package aws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/open-policy-agent/opa/internal/version"
	"github.com/open-policy-agent/opa/v1/logging"
)

// Values taken from
// https://docs.aws.amazon.com/AmazonECR/latest/APIReference/API_GetAuthorizationToken.html
const (
	ecrGetAuthorizationTokenTarget = "AmazonEC2ContainerRegistry_V20150921.GetAuthorizationToken"
	ecrEndpointFmt                 = "https://ecr.%s.amazonaws.com/"
)

// ECR is used to request tokens from Elastic Container Registry.
type ECR struct {
	// endpoint returns the region-specifc ECR endpoint.
	// It can be overridden by tests.
	endpoint func(region string) string

	// client is used to send authorization tokens requests.
	client *http.Client

	logger logging.Logger
}

func NewECR(logger logging.Logger) *ECR {
	return &ECR{
		endpoint: func(region string) string {
			return fmt.Sprintf(ecrEndpointFmt, region)
		},
		client: &http.Client{},
		logger: logger,
	}
}

// GetAuthorizationToken requests a token that can be used to authenticate image pull requests.
func (e *ECR) GetAuthorizationToken(ctx context.Context, creds Credentials, signatureVersion string) (ECRAuthorizationToken, error) {
	endpoint := e.endpoint(creds.RegionName)
	body := strings.NewReader("{}")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, body)
	if err != nil {
		return ECRAuthorizationToken{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Amz-Target", ecrGetAuthorizationTokenTarget)
	req.Header.Set("Accept-Encoding", "identity")
	req.Header.Set("Content-Type", "application/x-amz-json-1.1")
	req.Header.Set("User-Agent", version.UserAgent)

	e.logger.Debug("Signing ECR authorization token request")

	if err := SignRequest(req, "ecr", creds, time.Now(), signatureVersion); err != nil {
		return ECRAuthorizationToken{}, fmt.Errorf("failed to sign request: %w", err)
	}

	resp, err := DoRequestWithClient(req, e.client, "ecr get authorization token", e.logger)
	if err != nil {
		return ECRAuthorizationToken{}, err
	}

	var data struct {
		AuthorizationData []struct {
			AuthorizationToken string      `json:"authorizationToken"`
			ExpiresAt          json.Number `json:"expiresAt"`
		} `json:"authorizationData"`
	}
	if err := json.Unmarshal(resp, &data); err != nil {
		return ECRAuthorizationToken{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(data.AuthorizationData) < 1 {
		return ECRAuthorizationToken{}, errors.New("empty authorization data")
	}

	// The GetAuthorizationToken request returns a list of tokens for
	// backwards compatibility reasons. We should only ever get one token back
	// because we don't define any registryIDs in the request.
	// See https://docs.aws.amazon.com/AmazonECR/latest/APIReference/API_GetAuthorizationToken.html#API_GetAuthorizationToken_ResponseSyntax
	resultToken := data.AuthorizationData[0]

	expiresAt, err := parseTimestamp(resultToken.ExpiresAt)
	if err != nil {
		return ECRAuthorizationToken{}, fmt.Errorf("failed to parse expiresAt: %w", err)
	}

	return ECRAuthorizationToken{
		AuthorizationToken: resultToken.AuthorizationToken,
		ExpiresAt:          expiresAt,
	}, nil
}

// ECRAuthorizationToken can sign requests to AWS ECR.
//
// It corresponds to data returned by the AWS GetAuthorizationToken API.
// See https://docs.aws.amazon.com/AmazonECR/latest/APIReference/API_AuthorizationData.html
type ECRAuthorizationToken struct {
	AuthorizationToken string
	ExpiresAt          time.Time
}

// IsValid returns true if the token is set and not expired.
// It respects a margin of error for time handling and will mark it as expired early.
func (t *ECRAuthorizationToken) IsValid() bool {
	const tokenExpirationMargin = 5 * time.Minute

	expired := time.Now().Add(tokenExpirationMargin).After(t.ExpiresAt)
	return t.AuthorizationToken != "" && !expired
}

var millisecondsFloat = new(big.Float).SetInt64(1e3)

// parseTimestamp parses the AWS format for timestamps.
// The time precision is in milliseconds.
//
// The logic is taken from
// https://github.com/aws/aws-sdk-go/blob/41717ba2c04d3fd03f94d09ea984a10899574935/private/protocol/json/jsonutil/unmarshal.go#L294-L302
func parseTimestamp(raw json.Number) (time.Time, error) {
	s := raw.String()

	float, ok := new(big.Float).SetString(s)
	if !ok {
		return time.Time{}, fmt.Errorf("not a float: %q", raw)
	}

	// The float is expected to be in second resolution with millisecond
	// decimal places.
	// Multiply by millisecondsFloat to obtain an integer in millisecond
	// resolution
	ms, _ := float.Mul(float, millisecondsFloat).Int64()

	// Multiply again to obtain nanosecond resolution for time.Unix
	ns := ms * 1e6

	t := time.Unix(0, ns).UTC()

	return t, nil
}
