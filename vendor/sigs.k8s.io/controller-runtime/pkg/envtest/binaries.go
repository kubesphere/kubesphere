/*
Copyright 2025 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package envtest

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/blang/semver/v4"
	"sigs.k8s.io/yaml"
)

// DefaultBinaryAssetsIndexURL is the default index used in HTTPClient.
var DefaultBinaryAssetsIndexURL = "https://raw.githubusercontent.com/kubernetes-sigs/controller-tools/HEAD/envtest-releases.yaml"

// SetupEnvtestDefaultBinaryAssetsDirectory returns the default location that setup-envtest uses to store envtest binaries.
// Setting BinaryAssetsDirectory to this directory allows sharing envtest binaries with setup-envtest.
//
// The directory is dependent on operating system:
//
// - Windows: %LocalAppData%\kubebuilder-envtest
// - OSX: ~/Library/Application Support/io.kubebuilder.envtest
// - Others: ${XDG_DATA_HOME:-~/.local/share}/kubebuilder-envtest
//
// Otherwise, it errors out.  Note that these paths must not be relied upon
// manually.
func SetupEnvtestDefaultBinaryAssetsDirectory() (string, error) {
	var baseDir string

	// find the base data directory
	switch runtime.GOOS {
	case "windows":
		baseDir = os.Getenv("LocalAppData")
		if baseDir == "" {
			return "", errors.New("%LocalAppData% is not defined")
		}
	case "darwin":
		homeDir := os.Getenv("HOME")
		if homeDir == "" {
			return "", errors.New("$HOME is not defined")
		}
		baseDir = filepath.Join(homeDir, "Library/Application Support")
	default:
		baseDir = os.Getenv("XDG_DATA_HOME")
		if baseDir == "" {
			homeDir := os.Getenv("HOME")
			if homeDir == "" {
				return "", errors.New("neither $XDG_DATA_HOME nor $HOME are defined")
			}
			baseDir = filepath.Join(homeDir, ".local/share")
		}
	}

	// append our program-specific dir to it (OSX has a slightly different
	// convention so try to follow that).
	switch runtime.GOOS {
	case "darwin", "ios":
		return filepath.Join(baseDir, "io.kubebuilder.envtest", "k8s"), nil
	default:
		return filepath.Join(baseDir, "kubebuilder-envtest", "k8s"), nil
	}
}

// index represents an index of envtest binary archives. Example:
//
//	releases:
//		v1.28.0:
//			envtest-v1.28.0-darwin-amd64.tar.gz:
//	    		hash: <sha512-hash>
//				selfLink: <url-to-archive-with-envtest-binaries>
type index struct {
	// Releases maps Kubernetes versions to Releases (envtest archives).
	Releases map[string]release `json:"releases"`
}

// release maps an archive name to an archive.
type release map[string]archive

// archive contains the self link to an archive and its hash.
type archive struct {
	Hash     string `json:"hash"`
	SelfLink string `json:"selfLink"`
}

func downloadBinaryAssets(ctx context.Context, binaryAssetsDirectory, binaryAssetsVersion, binaryAssetsIndexURL string) (string, string, string, error) {
	if binaryAssetsIndexURL == "" {
		binaryAssetsIndexURL = DefaultBinaryAssetsIndexURL
	}

	downloadRootDir := binaryAssetsDirectory
	if downloadRootDir == "" {
		var err error
		if downloadRootDir, err = os.MkdirTemp("", "envtest-binaries-"); err != nil {
			return "", "", "", fmt.Errorf("failed to create tmp directory for envtest binaries: %w", err)
		}
	}

	var binaryAssetsIndex *index
	if binaryAssetsVersion == "" {
		var err error
		binaryAssetsIndex, err = getIndex(ctx, binaryAssetsIndexURL)
		if err != nil {
			return "", "", "", err
		}

		binaryAssetsVersion, err = latestStableVersionFromIndex(binaryAssetsIndex)
		if err != nil {
			return "", "", "", err
		}
	}

	// Storing the envtest binaries in a directory structure that is compatible with setup-envtest.
	// This makes it possible to share the envtest binaries with setup-envtest if the BinaryAssetsDirectory is set to SetupEnvtestDefaultBinaryAssetsDirectory().
	downloadDir := path.Join(downloadRootDir, fmt.Sprintf("%s-%s-%s", strings.TrimPrefix(binaryAssetsVersion, "v"), runtime.GOOS, runtime.GOARCH))
	if !fileExists(downloadDir) {
		if err := os.MkdirAll(downloadDir, 0700); err != nil {
			return "", "", "", fmt.Errorf("failed to create directory %q for envtest binaries: %w", downloadDir, err)
		}
	}

	apiServerPath := path.Join(downloadDir, "kube-apiserver")
	etcdPath := path.Join(downloadDir, "etcd")
	kubectlPath := path.Join(downloadDir, "kubectl")

	if fileExists(apiServerPath) && fileExists(etcdPath) && fileExists(kubectlPath) {
		// Nothing to do if the binaries already exist.
		return apiServerPath, etcdPath, kubectlPath, nil
	}

	// Get Index if we didn't have to get it above to get the latest stable version.
	if binaryAssetsIndex == nil {
		var err error
		binaryAssetsIndex, err = getIndex(ctx, binaryAssetsIndexURL)
		if err != nil {
			return "", "", "", err
		}
	}

	buf := &bytes.Buffer{}
	if err := downloadBinaryAssetsArchive(ctx, binaryAssetsIndex, binaryAssetsVersion, buf); err != nil {
		return "", "", "", err
	}

	gzStream, err := gzip.NewReader(buf)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create gzip reader to extract envtest binaries: %w", err)
	}
	tarReader := tar.NewReader(gzStream)

	var header *tar.Header
	for header, err = tarReader.Next(); err == nil; header, err = tarReader.Next() {
		if header.Typeflag != tar.TypeReg {
			// Skip non-regular file entry in archive.
			continue
		}

		// Just dump all files directly into the download directory, ignoring the prefixed directory paths.
		// We also ignore bits for the most part (except for X).
		fileName := filepath.Base(header.Name)
		perms := 0555 & header.Mode // make sure we're at most r+x

		// Setting O_EXCL to get an error if the file already exists.
		f, err := os.OpenFile(path.Join(downloadDir, fileName), os.O_RDWR|os.O_CREATE|os.O_EXCL|os.O_TRUNC, os.FileMode(perms))
		if err != nil {
			if os.IsExist(err) {
				// Nothing to do if the file already exists. We assume another process created the file concurrently.
				continue
			}
			return "", "", "", fmt.Errorf("failed to create file %s in directory %s: %w", fileName, downloadDir, err)
		}
		if err := func() error {
			defer f.Close()
			if _, err := io.Copy(f, tarReader); err != nil {
				return fmt.Errorf("failed to write file %s in directory %s: %w", fileName, downloadDir, err)
			}
			return nil
		}(); err != nil {
			return "", "", "", fmt.Errorf("failed to close file %s in directory %s: %w", fileName, downloadDir, err)
		}
	}

	return apiServerPath, etcdPath, kubectlPath, nil
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

func downloadBinaryAssetsArchive(ctx context.Context, index *index, version string, out io.Writer) error {
	archives, ok := index.Releases[version]
	if !ok {
		return fmt.Errorf("failed to find envtest binaries for version %s", version)
	}

	archiveName := fmt.Sprintf("envtest-%s-%s-%s.tar.gz", version, runtime.GOOS, runtime.GOARCH)
	archive, ok := archives[archiveName]
	if !ok {
		return fmt.Errorf("failed to find envtest binaries for version %s with archiveName %s", version, archiveName)
	}

	archiveURL, err := url.Parse(archive.SelfLink)
	if err != nil {
		return fmt.Errorf("failed to parse envtest binaries archive URL %q: %w", archiveURL, err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", archiveURL.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create request to download %s: %w", archiveURL.String(), err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download %s: %w", archiveURL.String(), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to download %s, got status %q", archiveURL.String(), resp.Status)
	}

	return readBody(resp, out, archiveName, archive.Hash)
}

func latestStableVersionFromIndex(index *index) (string, error) {
	if len(index.Releases) == 0 {
		return "", fmt.Errorf("failed to find latest stable version from index: index is empty")
	}

	parsedVersions := []semver.Version{}
	for releaseVersion := range index.Releases {
		v, err := semver.ParseTolerant(releaseVersion)
		if err != nil {
			return "", fmt.Errorf("failed to parse version %q: %w", releaseVersion, err)
		}

		// Filter out pre-releases.
		if len(v.Pre) > 0 {
			continue
		}

		parsedVersions = append(parsedVersions, v)
	}

	if len(parsedVersions) == 0 {
		return "", fmt.Errorf("failed to find latest stable version from index: index does not have stable versions")
	}

	sort.Slice(parsedVersions, func(i, j int) bool {
		return parsedVersions[i].GT(parsedVersions[j])
	})
	return "v" + parsedVersions[0].String(), nil
}

func getIndex(ctx context.Context, indexURL string) (*index, error) {
	loc, err := url.Parse(indexURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse index URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", loc.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to construct request to get index: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to perform request to get index: %w", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unable to get index -- got status %q", resp.Status)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to get index -- unable to read body %w", err)
	}

	var index index
	if err := yaml.Unmarshal(responseBody, &index); err != nil {
		return nil, fmt.Errorf("unable to unmarshal index: %w", err)
	}
	return &index, nil
}

func readBody(resp *http.Response, out io.Writer, archiveName string, expectedHash string) error {
	// Stream in chunks to do the checksum
	buf := make([]byte, 32*1024) // 32KiB, same as io.Copy
	hasher := sha512.New()

	for cont := true; cont; {
		amt, err := resp.Body.Read(buf)
		if err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("unable read next chunk of %s: %w", archiveName, err)
		}
		if amt > 0 {
			// checksum never returns errors according to docs
			hasher.Write(buf[:amt])
			if _, err := out.Write(buf[:amt]); err != nil {
				return fmt.Errorf("unable write next chunk of %s: %w", archiveName, err)
			}
		}
		cont = amt > 0 && !errors.Is(err, io.EOF)
	}

	actualHash := hex.EncodeToString(hasher.Sum(nil))
	if actualHash != expectedHash {
		return fmt.Errorf("checksum mismatch for %s: %s (computed) != %s (expected)", archiveName, actualHash, expectedHash)
	}

	return nil
}
