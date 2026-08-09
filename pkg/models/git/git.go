/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package git

import (
	"context"
	"fmt"
	"net"
	nethttp "net/http"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	transportclient "github.com/go-git/go-git/v5/plumbing/transport/client"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type AuthInfo struct {
	RemoteUrl string                  `json:"remoteUrl" description:"git server url"`
	SecretRef *corev1.SecretReference `json:"secretRef,omitempty" description:"auth secret reference"`
}

type GitVerifier interface {
	VerifyGitCredential(remoteUrl, namespace, secretName string) error
}

type gitVerifier struct {
	cache runtimeclient.Reader
}

func NewGitVerifier(cacheReader runtimeclient.Reader) GitVerifier {
	return &gitVerifier{cache: cacheReader}
}

func (c *gitVerifier) VerifyGitCredential(remoteUrl, namespace, secretName string) error {
	var username, password string

	if len(secretName) > 0 {
		secret := &corev1.Secret{}
		if err := c.cache.Get(context.Background(),
			types.NamespacedName{Namespace: namespace, Name: secretName}, secret); err != nil {
			return err
		}
		usernameBytes, ok := secret.Data[corev1.BasicAuthUsernameKey]
		if !ok {
			return fmt.Errorf("could not get username in secret %s", secret.Name)
		}
		passwordBytes, ok := secret.Data[corev1.BasicAuthPasswordKey]
		if !ok {
			return fmt.Errorf("could not get password in secret %s", secret.Name)
		}
		username = string(usernameBytes)
		password = string(passwordBytes)
	}

	return c.gitReadVerifyWithBasicAuth(username, password, remoteUrl)
}

func (c *gitVerifier) gitReadVerifyWithBasicAuth(username string, password string, remote string) error {
	installSSRFGuard()

	r, _ := git.Init(memory.NewStorage(), nil)
	// Add a new remote, with the default fetch refspec
	origin, err := r.CreateRemote(&config.RemoteConfig{
		Name: git.DefaultRemoteName,
		URLs: []string{remote},
	})
	if err != nil {
		return err
	}
	_, err = origin.List(&git.ListOptions{Auth: &http.BasicAuth{Username: username, Password: password}})
	return err
}

var installSSRFGuardOnce sync.Once

// installSSRFGuard restricts go-git's http/https transport to dial only publicly
// routable addresses. Without this, VerifyGitCredential lets any authenticated user
// make ks-apiserver issue an arbitrary outbound HTTP request (with credentials from
// an attacker-chosen Secret attached, when secretRef is set) to a caller-supplied
// remoteUrl, and reflects the upstream response body back in the API error message.
// That allows server-side request forgery against loopback, link-local (including
// the 169.254.169.254 cloud metadata endpoint) and private in-cluster addresses.
func installSSRFGuard() {
	installSSRFGuardOnce.Do(func() {
		dialer := &net.Dialer{Timeout: 30 * time.Second}
		transport := &nethttp.Transport{
			Proxy:       nethttp.ProxyFromEnvironment,
			DialContext: safeDialContext(dialer),
		}
		safeClient := http.NewClient(&nethttp.Client{Transport: transport})
		transportclient.InstallProtocol("http", safeClient)
		transportclient.InstallProtocol("https", safeClient)
	})
}

// safeDialContext resolves the target host once, rejects it if any resolved address
// is disallowed, and then dials the validated IP directly (rather than the hostname)
// so a DNS response that changes between the check and the dial cannot bypass it.
func safeDialContext(dialer *net.Dialer) func(ctx context.Context, network, addr string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}
		ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
		if err != nil {
			return nil, err
		}
		var lastErr error
		for _, ip := range ips {
			if isDisallowedRemoteAddr(ip.IP) {
				lastErr = fmt.Errorf("connections to %s are not permitted", ip.IP)
				continue
			}
			conn, dialErr := dialer.DialContext(ctx, network, net.JoinHostPort(ip.IP.String(), port))
			if dialErr == nil {
				return conn, nil
			}
			lastErr = dialErr
		}
		if lastErr == nil {
			lastErr = fmt.Errorf("no route to host %s", host)
		}
		return nil, lastErr
	}
}

// isDisallowedRemoteAddr reports whether ip is a loopback, link-local, unspecified
// or RFC 1918 / RFC 4193 private address that a git remote must never resolve to.
func isDisallowedRemoteAddr(ip net.IP) bool {
	return ip.IsLoopback() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsUnspecified() ||
		ip.IsPrivate()
}
