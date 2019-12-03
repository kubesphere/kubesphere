package dns01

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/go-acme/lego/v3/acme"
	"github.com/go-acme/lego/v3/acme/api"
	"github.com/go-acme/lego/v3/challenge"
	"github.com/go-acme/lego/v3/log"
	"github.com/go-acme/lego/v3/platform/wait"
	"github.com/miekg/dns"
)

const (
	// DefaultPropagationTimeout default propagation timeout
	DefaultPropagationTimeout = 60 * time.Second

	// DefaultPollingInterval default polling interval
	DefaultPollingInterval = 2 * time.Second

	// DefaultTTL default TTL
	DefaultTTL = 120
)

type ValidateFunc func(core *api.Core, domain string, chlng acme.Challenge) error

type ChallengeOption func(*Challenge) error

// CondOption Conditional challenge option.
func CondOption(condition bool, opt ChallengeOption) ChallengeOption {
	if !condition {
		// NoOp options
		return func(*Challenge) error {
			return nil
		}
	}
	return opt
}

// Challenge implements the dns-01 challenge
type Challenge struct {
	core       *api.Core
	validate   ValidateFunc
	provider   challenge.Provider
	preCheck   preCheck
	dnsTimeout time.Duration
}

func NewChallenge(core *api.Core, validate ValidateFunc, provider challenge.Provider, opts ...ChallengeOption) *Challenge {
	chlg := &Challenge{
		core:       core,
		validate:   validate,
		provider:   provider,
		preCheck:   newPreCheck(),
		dnsTimeout: 10 * time.Second,
	}

	for _, opt := range opts {
		err := opt(chlg)
		if err != nil {
			log.Infof("challenge option error: %v", err)
		}
	}

	return chlg
}

// PreSolve just submits the txt record to the dns provider.
// It does not validate record propagation, or do anything at all with the acme server.
func (c *Challenge) PreSolve(authz acme.Authorization) error {
	domain := challenge.GetTargetedDomain(authz)
	log.Infof("[%s] acme: Preparing to solve DNS-01", domain)

	chlng, err := challenge.FindChallenge(challenge.DNS01, authz)
	if err != nil {
		return err
	}

	if c.provider == nil {
		return fmt.Errorf("[%s] acme: no DNS Provider configured", domain)
	}

	// Generate the Key Authorization for the challenge
	keyAuth, err := c.core.GetKeyAuthorization(chlng.Token)
	if err != nil {
		return err
	}

	err = c.provider.Present(authz.Identifier.Value, chlng.Token, keyAuth)
	if err != nil {
		return fmt.Errorf("[%s] acme: error presenting token: %s", domain, err)
	}

	return nil
}

func (c *Challenge) Solve(authz acme.Authorization) error {
	domain := challenge.GetTargetedDomain(authz)
	log.Infof("[%s] acme: Trying to solve DNS-01", domain)

	chlng, err := challenge.FindChallenge(challenge.DNS01, authz)
	if err != nil {
		return err
	}

	// Generate the Key Authorization for the challenge
	keyAuth, err := c.core.GetKeyAuthorization(chlng.Token)
	if err != nil {
		return err
	}

	fqdn, value := GetRecord(authz.Identifier.Value, keyAuth)

	var timeout, interval time.Duration
	switch provider := c.provider.(type) {
	case challenge.ProviderTimeout:
		timeout, interval = provider.Timeout()
	default:
		timeout, interval = DefaultPropagationTimeout, DefaultPollingInterval
	}

	log.Infof("[%s] acme: Checking DNS record propagation using %+v", domain, recursiveNameservers)

	err = wait.For("propagation", timeout, interval, func() (bool, error) {
		stop, errP := c.preCheck.call(domain, fqdn, value)
		if !stop || errP != nil {
			log.Infof("[%s] acme: Waiting for DNS record propagation.", domain)
		}
		return stop, errP
	})
	if err != nil {
		return err
	}

	chlng.KeyAuthorization = keyAuth
	return c.validate(c.core, domain, chlng)
}

// CleanUp cleans the challenge.
func (c *Challenge) CleanUp(authz acme.Authorization) error {
	log.Infof("[%s] acme: Cleaning DNS-01 challenge", challenge.GetTargetedDomain(authz))

	chlng, err := challenge.FindChallenge(challenge.DNS01, authz)
	if err != nil {
		return err
	}

	keyAuth, err := c.core.GetKeyAuthorization(chlng.Token)
	if err != nil {
		return err
	}

	return c.provider.CleanUp(authz.Identifier.Value, chlng.Token, keyAuth)
}

func (c *Challenge) Sequential() (bool, time.Duration) {
	if p, ok := c.provider.(sequential); ok {
		return ok, p.Sequential()
	}
	return false, 0
}

type sequential interface {
	Sequential() time.Duration
}

// GetRecord returns a DNS record which will fulfill the `dns-01` challenge
func GetRecord(domain, keyAuth string) (fqdn string, value string) {
	keyAuthShaBytes := sha256.Sum256([]byte(keyAuth))
	// base64URL encoding without padding
	value = base64.RawURLEncoding.EncodeToString(keyAuthShaBytes[:sha256.Size])
	fqdn = fmt.Sprintf("_acme-challenge.%s.", domain)

	if ok, _ := strconv.ParseBool(os.Getenv("LEGO_EXPERIMENTAL_CNAME_SUPPORT")); ok {
		r, err := dnsQuery(fqdn, dns.TypeCNAME, recursiveNameservers, true)
		// Check if the domain has CNAME then return that
		if err == nil && r.Rcode == dns.RcodeSuccess {
			fqdn = updateDomainWithCName(r, fqdn)
		}
	}

	return
}
