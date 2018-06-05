/*
Copyright 2016 The Kubernetes Authors.

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

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	bootstrapapi "k8s.io/client-go/tools/bootstrap/token/api"
	kubeadmapiext "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1alpha1"
)

const (
	TokenExpectedRegex = "^\\S{6}\\.\\S{16}\n$"
	TestConfig         = `apiVersion: v1
clusters:
- cluster:
    certificate-authority-data:
    server: localhost:8000
  name: prod
contexts:
- context:
    cluster: prod
    namespace: default
    user: default-service-account
  name: default
current-context: default
kind: Config
preferences: {}
users:
- name: kubernetes-admin
  user:
    client-certificate-data:
    client-key-data:
`
	TestConfigCertAuthorityData = "certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN5RENDQWJDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRFM01USXhOREUxTlRFek1Gb1hEVEkzTVRJeE1qRTFOVEV6TUZvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBTlZrCnNkT0NjRDBIOG9ycXZ5djBEZ09jZEpjRGc4aTJPNGt3QVpPOWZUanJGRHJqbDZlVXRtdlMyZ1lZd0c4TGhPV2gKb0lkZ3AvbVkrbVlDakliUUJtTmE2Ums1V2JremhJRzM1c1lseE9NVUJJR0xXMzN0RTh4SlR1RVd3V0NmZnpLcQpyaU1UT1A3REF3MUxuM2xUNlpJNGRNM09NOE1IUk9Wd3lRMDVpbWo5eUx5R1lYdTlvSncwdTVXWVpFYmpUL3VpCjJBZ2QwVDMrZGFFb044aVBJOTlVQkQxMzRkc2VGSEJEY3hHcmsvVGlQdHBpSC9IOGoxRWZaYzRzTGlONzJmL2YKYUpacTROSHFiT2F5UkpITCtJejFNTW1DRkN3cjdHOHVENWVvWWp2dEdZN2xLc1pBTlUwK3VlUnJsTitxTzhQWQpxaTZNMDFBcmV1UzFVVHFuTkM4Q0F3RUFBYU1qTUNFd0RnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFNbXo4Nm9LMmFLa0owMnlLSC9ZcTlzaDZZcDEKYmhLS25mMFJCaTA1clRacWdhTi9oTnROdmQxSzJxZGRLNzhIT2pVdkpNRGp3NERieXF0Wll2V01XVFRCQnQrSgpPMGNyWkg5NXlqUW42YzRlcU1FTjFhOUFKNXRlclNnTDVhREJsK0FMTWxaNVpxTzBUOUJDdTJtNXV3dGNWaFZuCnh6cGpTT3V5WVdOQ3A5bW9mV2VPUTljNXhEcElWeUlMUkFvNmZ5Z2c3N25TSDN4ckVmd0VKUHFMd1RPYVk1bTcKeEZWWWJoR3dxUGU5V0I5aTR5cnNrZUFBWlpUSzdVbklKMXFkRmlHQk9aZlRtaDhYQ3BOTHZZcFBLQW9hWWlsRwpjOW1acVhpWVlESTV6R1IxMElpc2FWNXJUY2hDenNQVWRhQzRVbnpTZG01cTdKYTAyb0poQlU1TE1FMD0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo="
	TestConfigNoCluster         = `apiVersion: v1
clusters:
- cluster:
    server:
  name: prod
contexts:
- context:
    namespace: default
    user: default-service-account
  name: default
kind: Config
preferences: {}
`
)

func TestRunGenerateToken(t *testing.T) {
	var buf bytes.Buffer

	err := RunGenerateToken(&buf)
	if err != nil {
		t.Errorf("RunGenerateToken returned an error: %v", err)
	}

	output := buf.String()

	matched, err := regexp.MatchString(TokenExpectedRegex, output)
	if err != nil {
		t.Fatalf("Encountered an error while trying to match RunGenerateToken's output: %v", err)
	}
	if !matched {
		t.Errorf("RunGenerateToken's output did not match expected regex; wanted: [%s], got: [%s]", TokenExpectedRegex, output)
	}
}

func TestRunCreateToken(t *testing.T) {
	var buf bytes.Buffer
	fakeClient := &fake.Clientset{}
	fakeClient.AddReactor("get", "secrets", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, errors.NewNotFound(v1.Resource("secrets"), "foo")
	})

	testCases := []struct {
		name          string
		token         string
		usages        []string
		extraGroups   []string
		printJoin     bool
		expectedError bool
	}{
		{
			name:          "valid: empty token",
			token:         "",
			usages:        []string{"signing", "authentication"},
			extraGroups:   []string{"system:bootstrappers:foo"},
			expectedError: false,
		},
		{
			name:          "valid: non-empty token",
			token:         "abcdef.1234567890123456",
			usages:        []string{"signing", "authentication"},
			extraGroups:   []string{"system:bootstrappers:foo"},
			expectedError: false,
		},
		{
			name:          "valid: no extraGroups",
			token:         "abcdef.1234567890123456",
			usages:        []string{"signing", "authentication"},
			extraGroups:   []string{},
			expectedError: false,
		},
		{
			name:          "invalid: incorrect token",
			token:         "123456.AABBCCDDEEFFGGHH",
			usages:        []string{"signing", "authentication"},
			extraGroups:   []string{},
			expectedError: true,
		},
		{
			name:          "invalid: incorrect extraGroups",
			token:         "abcdef.1234567890123456",
			usages:        []string{"signing", "authentication"},
			extraGroups:   []string{"foo"},
			expectedError: true,
		},
		{
			name:          "invalid: specifying --groups when --usages doesn't include authentication",
			token:         "abcdef.1234567890123456",
			usages:        []string{"signing"},
			extraGroups:   []string{"foo"},
			expectedError: true,
		},
		{
			name:          "invalid: partially incorrect usages",
			token:         "abcdef.1234567890123456",
			usages:        []string{"foo", "authentication"},
			extraGroups:   []string{"system:bootstrappers:foo"},
			expectedError: true,
		},
		{
			name:          "invalid: all incorrect usages",
			token:         "abcdef.1234567890123456",
			usages:        []string{"foo", "bar"},
			extraGroups:   []string{"system:bootstrappers:foo"},
			expectedError: true,
		},
		{
			name:          "invalid: print join command",
			token:         "",
			usages:        []string{"signing", "authentication"},
			extraGroups:   []string{"system:bootstrappers:foo"},
			printJoin:     true,
			expectedError: true,
		},
	}
	for _, tc := range testCases {
		cfg := &kubeadmapiext.MasterConfiguration{

			// KubernetesVersion is not used by bootstrap-token, but we set this explicitly to avoid
			// the lookup of the version from the internet when executing ConfigFileAndDefaultsToInternalConfig
			KubernetesVersion: "v1.9.0",
			Token:             tc.token,
			TokenTTL:          &metav1.Duration{Duration: 0},
			TokenUsages:       tc.usages,
			TokenGroups:       tc.extraGroups,
		}

		err := RunCreateToken(&buf, fakeClient, "", cfg, "", tc.printJoin, "")
		if (err != nil) != tc.expectedError {
			t.Errorf("Test case %s: RunCreateToken expected error: %v, saw: %v", tc.name, tc.expectedError, (err != nil))
		}
	}
}

func TestNewCmdTokenGenerate(t *testing.T) {
	var buf bytes.Buffer
	args := []string{}

	cmd := NewCmdTokenGenerate(&buf)
	cmd.SetArgs(args)

	if err := cmd.Execute(); err != nil {
		t.Errorf("Cannot execute token command: %v", err)
	}
}

func TestNewCmdToken(t *testing.T) {
	var buf, bufErr bytes.Buffer
	testConfigFile := "test-config-file"

	tmpDir, err := ioutil.TempDir("", "kubeadm-token-test")
	if err != nil {
		t.Errorf("Unable to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	fullPath := filepath.Join(tmpDir, testConfigFile)

	f, err := os.Create(fullPath)
	if err != nil {
		t.Errorf("Unable to create test file %q: %v", fullPath, err)
	}
	defer f.Close()

	testCases := []struct {
		name          string
		args          []string
		configToWrite string
		expectedError bool
	}{
		{
			name:          "valid: generate",
			args:          []string{"generate"},
			configToWrite: "",
			expectedError: false,
		},
		{
			name:          "valid: delete",
			args:          []string{"delete", "abcdef.1234567890123456", "--dry-run", "--kubeconfig=" + fullPath},
			configToWrite: TestConfig,
			expectedError: false,
		},
	}

	cmd := NewCmdToken(&buf, &bufErr)
	for _, tc := range testCases {
		if _, err = f.WriteString(tc.configToWrite); err != nil {
			t.Errorf("Unable to write test file %q: %v", fullPath, err)
		}
		cmd.SetArgs(tc.args)
		err := cmd.Execute()
		if (err != nil) != tc.expectedError {
			t.Errorf("Test case %q: NewCmdToken expected error: %v, saw: %v", tc.name, tc.expectedError, (err != nil))
		}
	}
}

func TestGetSecretString(t *testing.T) {
	secret := v1.Secret{}
	key := "test-key"
	if str := getSecretString(&secret, key); str != "" {
		t.Errorf("getSecretString() did not return empty string for a nil v1.Secret.Data")
	}
	secret.Data = make(map[string][]byte)
	if str := getSecretString(&secret, key); str != "" {
		t.Errorf("getSecretString() did not return empty string for missing v1.Secret.Data key")
	}
	secret.Data[key] = []byte("test-value")
	if str := getSecretString(&secret, key); str == "" {
		t.Errorf("getSecretString() failed for a valid v1.Secret.Data key")
	}
}

func TestGetClientset(t *testing.T) {
	testConfigFile := "test-config-file"

	tmpDir, err := ioutil.TempDir("", "kubeadm-token-test")
	if err != nil {
		t.Errorf("Unable to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	fullPath := filepath.Join(tmpDir, testConfigFile)

	// test dryRun = false on a non-exisiting file
	if _, err = getClientset(fullPath, false); err == nil {
		t.Errorf("getClientset(); dry-run: false; did no fail for test file %q: %v", fullPath, err)
	}

	// test dryRun = true on a non-exisiting file
	if _, err = getClientset(fullPath, true); err == nil {
		t.Errorf("getClientset(); dry-run: true; did no fail for test file %q: %v", fullPath, err)
	}

	f, err := os.Create(fullPath)
	if err != nil {
		t.Errorf("Unable to create test file %q: %v", fullPath, err)
	}
	defer f.Close()

	if _, err = f.WriteString(TestConfig); err != nil {
		t.Errorf("Unable to write test file %q: %v", fullPath, err)
	}

	// test dryRun = true on an exisiting file
	if _, err = getClientset(fullPath, true); err != nil {
		t.Errorf("getClientset(); dry-run: true; failed for test file %q: %v", fullPath, err)
	}
}

func TestRunDeleteToken(t *testing.T) {
	var buf bytes.Buffer

	tmpDir, err := ioutil.TempDir("", "kubeadm-token-test")
	if err != nil {
		t.Errorf("Unable to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	fullPath := filepath.Join(tmpDir, "test-config-file")

	f, err := os.Create(fullPath)
	if err != nil {
		t.Errorf("Unable to create test file %q: %v", fullPath, err)
	}
	defer f.Close()

	if _, err = f.WriteString(TestConfig); err != nil {
		t.Errorf("Unable to write test file %q: %v", fullPath, err)
	}

	client, err := getClientset(fullPath, true)
	if err != nil {
		t.Errorf("Unable to run getClientset() for test file %q: %v", fullPath, err)
	}

	// test valid; should not fail
	// for some reason Secrets().Delete() does not fail even for this dummy config
	if err = RunDeleteToken(&buf, client, "abcdef.1234567890123456"); err != nil {
		t.Errorf("RunDeleteToken() failed for a valid token: %v", err)
	}

	// test invalid token; should fail
	if err = RunDeleteToken(&buf, client, "invalid-token"); err == nil {
		t.Errorf("RunDeleteToken() succeeded for an invalid token: %v", err)
	}
}

var httpTestItr uint32
var httpSentResponse uint32 = 1

func TestRunListTokens(t *testing.T) {
	var err error
	var bufOut, bufErr bytes.Buffer

	tmpDir, err := ioutil.TempDir("", "kubeadm-token-test")
	if err != nil {
		t.Errorf("Unable to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	fullPath := filepath.Join(tmpDir, "test-config-file")

	f, err := os.Create(fullPath)
	if err != nil {
		t.Errorf("Unable to create test file %q: %v", fullPath, err)
	}
	defer f.Close()

	// test config without secrets; should fail
	if _, err = f.WriteString(TestConfig); err != nil {
		t.Errorf("Unable to write test file %q: %v", fullPath, err)
	}

	client, err := getClientset(fullPath, true)
	if err != nil {
		t.Errorf("Unable to run getClientset() for test file %q: %v", fullPath, err)
	}

	if err = RunListTokens(&bufOut, &bufErr, client); err == nil {
		t.Errorf("RunListTokens() did not fail for a config without secrets: %v", err)
	}

	// test config without secrets but use a dummy API server that returns secrets
	portString := "9008"
	http.HandleFunc("/", httpHandler)
	httpServer := &http.Server{Addr: "localhost:" + portString}
	go func() {
		err := httpServer.ListenAndServe()
		if err != nil {
			t.Errorf("Failed to start dummy API server: localhost:%s", portString)
		}
	}()

	fmt.Printf("dummy API server listening on localhost:%s\n", portString)
	testConfigOpenPort := strings.Replace(TestConfig, "server: localhost:8000", "server: localhost:"+portString, -1)

	if _, err = f.WriteString(testConfigOpenPort); err != nil {
		t.Errorf("Unable to write test file %q: %v", fullPath, err)
	}

	client, err = getClientset(fullPath, true)
	if err != nil {
		t.Errorf("Unable to run getClientset() for test file %q: %v", fullPath, err)
	}

	// the order of these tests should match the case check
	// for httpTestItr in httpHandler
	testCases := []struct {
		name          string
		expectedError bool
	}{
		{
			name:          "token-id not defined",
			expectedError: true,
		},
		{
			name:          "secret name not formatted correctly",
			expectedError: true,
		},
		{
			name:          "token-secret not defined",
			expectedError: true,
		},
		{
			name:          "token expiration not formatted correctly",
			expectedError: true,
		},
		{
			name:          "token expiration formatted correctly",
			expectedError: false,
		},
		{
			name:          "token usage constant not true",
			expectedError: false,
		},
		{
			name:          "token usage constant set to true",
			expectedError: false,
		},
	}
	for _, tc := range testCases {
		bufErr.Reset()
		atomic.StoreUint32(&httpSentResponse, 0)
		fmt.Printf("Running HTTP test case (%d) %q\n", atomic.LoadUint32(&httpTestItr), tc.name)
		// should always return nil here if a valid list of secrets if fetched
		err := RunListTokens(&bufOut, &bufErr, client)
		if err != nil {
			t.Errorf("HTTP test case %d: Was unable to fetch a list of secrets", atomic.LoadUint32(&httpTestItr))
		}
		// wait for a response from the dummy HTTP server
		timeSpent := 0 * time.Millisecond
		timeToSleep := 50 * time.Millisecond
		timeMax := 2000 * time.Millisecond
		for {
			if atomic.LoadUint32(&httpSentResponse) == 1 {
				break
			}
			if timeSpent >= timeMax {
				t.Errorf("HTTP test case %d: The server did not respond within %d ms", atomic.LoadUint32(&httpTestItr), timeMax)
			}
			timeSpent += timeToSleep
			time.Sleep(timeToSleep)
		}
		// check if an error is written in the error buffer
		hasError := bufErr.Len() != 0
		if hasError != tc.expectedError {
			t.Errorf("HTTP test case %d: RunListTokens expected error: %v, saw: %v; %v", atomic.LoadUint32(&httpTestItr), tc.expectedError, hasError, bufErr.String())
		}
	}
}

// only one of these should run at a time in a goroutine
func httpHandler(w http.ResponseWriter, r *http.Request) {
	tokenID := []byte("07401b")
	tokenSecret := []byte("f395accd246ae52d")
	tokenExpire := []byte("2012-11-01T22:08:41+00:00")
	badValue := "bad-value"
	name := bootstrapapi.BootstrapTokenSecretPrefix + string(tokenID)
	tokenUsageKey := bootstrapapi.BootstrapTokenUsagePrefix + "test"

	secret := v1.Secret{}
	secret.Type = bootstrapapi.SecretTypeBootstrapToken
	secret.TypeMeta = metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"}
	secret.Data = map[string][]byte{}

	switch atomic.LoadUint32(&httpTestItr) {
	case 0:
		secret.Data[bootstrapapi.BootstrapTokenIDKey] = []byte("")
	case 1:
		secret.Data[bootstrapapi.BootstrapTokenIDKey] = tokenID
		secret.ObjectMeta = metav1.ObjectMeta{Name: badValue}
	case 2:
		secret.Data[bootstrapapi.BootstrapTokenIDKey] = tokenID
		secret.Data[bootstrapapi.BootstrapTokenSecretKey] = []byte("")
		secret.ObjectMeta = metav1.ObjectMeta{Name: name}
	case 3:
		secret.Data[bootstrapapi.BootstrapTokenIDKey] = tokenID
		secret.Data[bootstrapapi.BootstrapTokenSecretKey] = tokenSecret
		secret.Data[bootstrapapi.BootstrapTokenExpirationKey] = []byte(badValue)
		secret.ObjectMeta = metav1.ObjectMeta{Name: name}
	case 4:
		secret.Data[bootstrapapi.BootstrapTokenIDKey] = tokenID
		secret.Data[bootstrapapi.BootstrapTokenSecretKey] = tokenSecret
		secret.Data[bootstrapapi.BootstrapTokenExpirationKey] = tokenExpire
		secret.ObjectMeta = metav1.ObjectMeta{Name: name}
	case 5:
		secret.Data[bootstrapapi.BootstrapTokenIDKey] = tokenID
		secret.Data[bootstrapapi.BootstrapTokenSecretKey] = tokenSecret
		secret.Data[bootstrapapi.BootstrapTokenExpirationKey] = tokenExpire
		secret.Data[tokenUsageKey] = []byte("false")
		secret.ObjectMeta = metav1.ObjectMeta{Name: name}
	case 6:
		secret.Data[bootstrapapi.BootstrapTokenIDKey] = tokenID
		secret.Data[bootstrapapi.BootstrapTokenSecretKey] = tokenSecret
		secret.Data[bootstrapapi.BootstrapTokenExpirationKey] = tokenExpire
		secret.Data[tokenUsageKey] = []byte("true")
		secret.ObjectMeta = metav1.ObjectMeta{Name: name}
	}

	secretList := v1.SecretList{}
	secretList.Items = []v1.Secret{secret}
	secretList.TypeMeta = metav1.TypeMeta{APIVersion: "v1", Kind: "SecretList"}

	output, err := json.Marshal(secretList)
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(output))
	}
	atomic.AddUint32(&httpTestItr, 1)
	atomic.StoreUint32(&httpSentResponse, 1)
}
