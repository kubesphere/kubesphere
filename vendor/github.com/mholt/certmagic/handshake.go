// Copyright 2015 Matthew Holt
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package certmagic

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/go-acme/lego/v3/challenge/tlsalpn01"
)

// GetCertificate gets a certificate to satisfy clientHello. In getting
// the certificate, it abides the rules and settings defined in the
// Config that matches clientHello.ServerName. It first checks the in-
// memory cache, then, if the config enables "OnDemand", it accesses
// disk, then accesses the network if it must obtain a new certificate
// via ACME.
//
// This method is safe for use as a tls.Config.GetCertificate callback.
func (cfg *Config) GetCertificate(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	if cfg.OnEvent != nil {
		cfg.OnEvent("tls_handshake_started", clientHello)
	}

	// special case: serve up the certificate for a TLS-ALPN ACME challenge
	// (https://tools.ietf.org/html/draft-ietf-acme-tls-alpn-05)
	for _, proto := range clientHello.SupportedProtos {
		if proto == tlsalpn01.ACMETLS1Protocol {
			cfg.certCache.mu.RLock()
			challengeCert, ok := cfg.certCache.cache[tlsALPNCertKeyName(clientHello.ServerName)]
			cfg.certCache.mu.RUnlock()
			if !ok {
				// see if this challenge was started in a cluster; try distributed challenge solver
				// (note that the tls.Config's ALPN settings must include the ACME TLS-ALPN challenge
				// protocol string, otherwise a valid certificate will not solve the challenge; we
				// should already have taken care of that when we made the tls.Config)
				challengeCert, ok, err := cfg.tryDistributedChallengeSolver(clientHello)
				if err != nil {
					log.Printf("[ERROR][%s] TLS-ALPN: %v", clientHello.ServerName, err)
				}
				if ok {
					return &challengeCert.Certificate, nil
				}

				return nil, fmt.Errorf("no certificate to complete TLS-ALPN challenge for SNI name: %s", clientHello.ServerName)
			}
			return &challengeCert.Certificate, nil
		}
	}

	// get the certificate and serve it up
	cert, err := cfg.getCertDuringHandshake(clientHello, true, true)
	if err == nil && cfg.OnEvent != nil {
		cfg.OnEvent("tls_handshake_completed", clientHello)
	}
	return &cert.Certificate, err
}

// getCertificate gets a certificate that matches name from the in-memory
// cache, according to the lookup table associated with cfg. The lookup then
// points to a certificate in the Instance certificate cache.
//
// The name is expected to already be normalized (e.g. lowercased).
//
// If there is no exact match for name, it will be checked against names of
// the form '*.example.com' (wildcard certificates) according to RFC 6125.
// If a match is found, matched will be true. If no matches are found, matched
// will be false and a "default" certificate will be returned with defaulted
// set to true. If defaulted is false, then no certificates were available.
//
// The logic in this function is adapted from the Go standard library,
// which is by the Go Authors.
//
// This function is safe for concurrent use.
func (cfg *Config) getCertificate(hello *tls.ClientHelloInfo) (cert Certificate, matched, defaulted bool) {
	name := NormalizedName(hello.ServerName)

	if name == "" {
		// if SNI is empty, prefer matching IP address
		if hello.Conn != nil {
			addr := hello.Conn.LocalAddr().String()
			ip, _, err := net.SplitHostPort(addr)
			if err == nil {
				addr = ip
			}
			cert, matched = cfg.selectCert(hello, addr)
			if matched {
				return
			}
		}

		// fall back to a "default" certificate, if specified
		if cfg.DefaultServerName != "" {
			normDefault := NormalizedName(cfg.DefaultServerName)
			cert, defaulted = cfg.selectCert(hello, normDefault)
			if defaulted {
				return
			}
		}
	} else {
		// if SNI is specified, try an exact match first
		cert, matched = cfg.selectCert(hello, name)
		if matched {
			return
		}

		// try replacing labels in the name with
		// wildcards until we get a match
		labels := strings.Split(name, ".")
		for i := range labels {
			labels[i] = "*"
			candidate := strings.Join(labels, ".")
			cert, matched = cfg.selectCert(hello, candidate)
			if matched {
				return
			}
		}

		// check the certCache directly to see if the SNI name is
		// already the key of the certificate it wants; this implies
		// that the SNI can contain the hash of a specific cert
		// (chain) it wants and we will still be able to serve it up
		// (this behavior, by the way, could be controversial as to
		// whether it complies with RFC 6066 about SNI, but I think
		// it does, soooo...)
		// (this is how we solved the former ACME TLS-SNI challenge)
		cfg.certCache.mu.RLock()
		directCert, ok := cfg.certCache.cache[name]
		cfg.certCache.mu.RUnlock()
		if ok {
			cert = directCert
			matched = true
			return
		}
	}

	// otherwise, we're bingo on ammo; see issues
	// mholt/caddy#2035 and mholt/caddy#1303 (any
	// change to certificate matching behavior must
	// account for hosts defined where the hostname
	// is empty or a catch-all, like ":443" or
	// "0.0.0.0:443")

	return
}

// selectCert uses hello to select a certificate from the
// cache for name. If cfg.CertSelection is set, it will be
// used to make the decision. Otherwise, the first matching
// unexpired cert is returned.
func (cfg *Config) selectCert(hello *tls.ClientHelloInfo, name string) (Certificate, bool) {
	choices := cfg.certCache.getAllMatchingCerts(name)
	if len(choices) == 0 {
		return Certificate{}, false
	}
	if cfg.CertSelection == nil {
		// by default, choose first non-expired cert
		now := time.Now()
		for _, choice := range choices {
			if now.Before(choice.NotAfter) {
				return choice, true
			}
		}
		// all matching certs are expired, oh well
		return choices[0], true
	}
	cert, err := cfg.CertSelection.SelectCertificate(hello, choices)
	return cert, err == nil
}

// getCertDuringHandshake will get a certificate for hello. It first tries
// the in-memory cache. If no certificate for hello is in the cache, the
// config most closely corresponding to hello will be loaded. If that config
// allows it (OnDemand==true) and if loadIfNecessary == true, it goes to disk
// to load it into the cache and serve it. If it's not on disk and if
// obtainIfNecessary == true, the certificate will be obtained from the CA,
// cached, and served. If obtainIfNecessary is true, then loadIfNecessary
// must also be set to true. An error will be returned if and only if no
// certificate is available.
//
// This function is safe for concurrent use.
func (cfg *Config) getCertDuringHandshake(hello *tls.ClientHelloInfo, loadIfNecessary, obtainIfNecessary bool) (Certificate, error) {
	name := NormalizedName(hello.ServerName)

	// First check our in-memory cache to see if we've already loaded it
	cert, matched, defaulted := cfg.getCertificate(hello)
	if matched {
		return cert, nil
	}

	// If OnDemand is enabled, then we might be able to load or
	// obtain a needed certificate
	if cfg.OnDemand != nil && loadIfNecessary {
		// Then check to see if we have one on disk
		loadedCert, err := cfg.CacheManagedCertificate(name)
		if err == nil {
			loadedCert, err = cfg.handshakeMaintenance(hello, loadedCert)
			if err != nil {
				log.Printf("[ERROR] Maintaining newly-loaded certificate for %s: %v", name, err)
			}
			return loadedCert, nil
		}
		if obtainIfNecessary {
			// By this point, we need to ask the CA for a certificate

			// Make sure the certificate should be obtained based on config
			err := cfg.checkIfCertShouldBeObtained(name)
			if err != nil {
				return Certificate{}, err
			}

			// Name has to qualify for a certificate
			if !HostQualifies(name) {
				return cert, fmt.Errorf("hostname '%s' does not qualify for certificate", name)
			}

			// Obtain certificate from the CA
			return cfg.obtainOnDemandCertificate(hello)
		}
	}

	// Fall back to the default certificate if there is one
	if defaulted {
		return cert, nil
	}

	return Certificate{}, fmt.Errorf("no certificate available for '%s'", name)
}

// checkIfCertShouldBeObtained checks to see if an on-demand TLS certificate
// should be obtained for a given domain based upon the config settings. If
// a non-nil error is returned, do not issue a new certificate for name.
func (cfg *Config) checkIfCertShouldBeObtained(name string) error {
	if cfg.OnDemand == nil {
		return fmt.Errorf("not configured for on-demand certificate issuance")
	}
	if cfg.OnDemand.DecisionFunc != nil {
		return cfg.OnDemand.DecisionFunc(name)
	}
	if len(cfg.OnDemand.hostWhitelist) > 0 &&
		!cfg.OnDemand.whitelistContains(name) {
		return fmt.Errorf("certificate for '%s' is not managed", name)
	}
	return nil
}

// obtainOnDemandCertificate obtains a certificate for hello.
// If another goroutine has already started obtaining a cert for
// hello, it will wait and use what the other goroutine obtained.
//
// This function is safe for use by multiple concurrent goroutines.
func (cfg *Config) obtainOnDemandCertificate(hello *tls.ClientHelloInfo) (Certificate, error) {
	name := NormalizedName(hello.ServerName)

	// We must protect this process from happening concurrently, so synchronize.
	obtainCertWaitChansMu.Lock()
	wait, ok := obtainCertWaitChans[name]
	if ok {
		// lucky us -- another goroutine is already obtaining the certificate.
		// wait for it to finish obtaining the cert and then we'll use it.
		obtainCertWaitChansMu.Unlock()
		<-wait
		return cfg.getCertDuringHandshake(hello, true, false)
	}

	// looks like it's up to us to do all the work and obtain the cert.
	// make a chan others can wait on if needed
	wait = make(chan struct{})
	obtainCertWaitChans[name] = wait
	obtainCertWaitChansMu.Unlock()

	// obtain the certificate
	log.Printf("[INFO] Obtaining new certificate for %s", name)
	err := cfg.ObtainCert(name, false)

	// immediately unblock anyone waiting for it; doing this in
	// a defer would risk deadlock because of the recursive call
	// to getCertDuringHandshake below when we return!
	obtainCertWaitChansMu.Lock()
	close(wait)
	delete(obtainCertWaitChans, name)
	obtainCertWaitChansMu.Unlock()

	if err != nil {
		// shucks; failed to solve challenge on-demand
		return Certificate{}, err
	}

	// success; certificate was just placed on disk, so
	// we need only restart serving the certificate
	return cfg.getCertDuringHandshake(hello, true, false)
}

// handshakeMaintenance performs a check on cert for expiration and OCSP validity.
//
// This function is safe for use by multiple concurrent goroutines.
func (cfg *Config) handshakeMaintenance(hello *tls.ClientHelloInfo, cert Certificate) (Certificate, error) {
	// Check cert expiration
	timeLeft := cert.NotAfter.Sub(time.Now().UTC())
	if timeLeft < cfg.RenewDurationBefore {
		log.Printf("[INFO] Certificate for %v expires in %v; attempting renewal", cert.Names, timeLeft)
		return cfg.renewDynamicCertificate(hello, cert)
	}

	// Check OCSP staple validity
	if cert.ocsp != nil {
		refreshTime := cert.ocsp.ThisUpdate.Add(cert.ocsp.NextUpdate.Sub(cert.ocsp.ThisUpdate) / 2)
		if time.Now().After(refreshTime) {
			_, err := stapleOCSP(cfg.Storage, &cert, nil)
			if err != nil {
				// An error with OCSP stapling is not the end of the world, and in fact, is
				// quite common considering not all certs have issuer URLs that support it.
				log.Printf("[ERROR] Getting OCSP for %s: %v", hello.ServerName, err)
			}
			cfg.certCache.mu.Lock()
			cfg.certCache.cache[cert.hash] = cert
			cfg.certCache.mu.Unlock()
		}
	}

	return cert, nil
}

// renewDynamicCertificate renews the certificate for name using cfg. It returns the
// certificate to use and an error, if any. name should already be lower-cased before
// calling this function. name is the name obtained directly from the handshake's
// ClientHello.
//
// This function is safe for use by multiple concurrent goroutines.
func (cfg *Config) renewDynamicCertificate(hello *tls.ClientHelloInfo, currentCert Certificate) (Certificate, error) {
	name := NormalizedName(hello.ServerName)

	obtainCertWaitChansMu.Lock()
	wait, ok := obtainCertWaitChans[name]
	if ok {
		// lucky us -- another goroutine is already renewing the certificate.
		// wait for it to finish, then we'll use the new one.
		obtainCertWaitChansMu.Unlock()
		<-wait
		return cfg.getCertDuringHandshake(hello, true, false)
	}

	// looks like it's up to us to do all the work and renew the cert
	wait = make(chan struct{})
	obtainCertWaitChans[name] = wait
	obtainCertWaitChansMu.Unlock()

	// renew and reload the certificate
	log.Printf("[INFO] Renewing certificate for %s", name)
	err := cfg.RenewCert(name, false)
	if err == nil {
		// even though the recursive nature of the dynamic cert loading
		// would just call this function anyway, we do it here to
		// make the replacement as atomic as possible.
		newCert, err := cfg.CacheManagedCertificate(name)
		if err != nil {
			log.Printf("[ERROR] loading renewed certificate for %s: %v", name, err)
		} else {
			// replace the old certificate with the new one
			cfg.certCache.replaceCertificate(currentCert, newCert)
		}
	}

	// immediately unblock anyone waiting for it; doing this in
	// a defer would risk deadlock because of the recursive call
	// to getCertDuringHandshake below when we return!
	obtainCertWaitChansMu.Lock()
	close(wait)
	delete(obtainCertWaitChans, name)
	obtainCertWaitChansMu.Unlock()

	if err != nil {
		return Certificate{}, err
	}

	return cfg.getCertDuringHandshake(hello, true, false)
}

// tryDistributedChallengeSolver is to be called when the clientHello pertains to
// a TLS-ALPN challenge and a certificate is required to solve it. This method
// checks the distributed store of challenge info files and, if a matching ServerName
// is present, it makes a certificate to solve this challenge and returns it.
// A boolean true is returned if a valid certificate is returned.
func (cfg *Config) tryDistributedChallengeSolver(clientHello *tls.ClientHelloInfo) (Certificate, bool, error) {
	tokenKey := distributedSolver{config: cfg}.challengeTokensKey(clientHello.ServerName)
	chalInfoBytes, err := cfg.Storage.Load(tokenKey)
	if err != nil {
		if _, ok := err.(ErrNotExist); ok {
			return Certificate{}, false, nil
		}
		return Certificate{}, false, fmt.Errorf("opening distributed challenge token file %s: %v", tokenKey, err)
	}

	var chalInfo challengeInfo
	err = json.Unmarshal(chalInfoBytes, &chalInfo)
	if err != nil {
		return Certificate{}, false, fmt.Errorf("decoding challenge token file %s (corrupted?): %v", tokenKey, err)
	}

	cert, err := tlsalpn01.ChallengeCert(chalInfo.Domain, chalInfo.KeyAuth)
	if err != nil {
		return Certificate{}, false, fmt.Errorf("making TLS-ALPN challenge certificate: %v", err)
	}
	if cert == nil {
		return Certificate{}, false, fmt.Errorf("got nil TLS-ALPN challenge certificate but no error")
	}

	return Certificate{Certificate: *cert}, true, nil
}

// NormalizedName returns a cleaned form of serverName that is
// used for consistency when referring to a SNI value.
func NormalizedName(serverName string) string {
	return strings.ToLower(strings.TrimSpace(serverName))
}

// obtainCertWaitChans is used to coordinate obtaining certs for each hostname.
var obtainCertWaitChans = make(map[string]chan struct{})
var obtainCertWaitChansMu sync.Mutex
