package aws

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/open-policy-agent/opa/internal/version"
	"github.com/open-policy-agent/opa/logging"
)

// Values taken from
// https://docs.aws.amazon.com/kms/latest/APIReference/Welcome.html
// https://docs.aws.amazon.com/general/latest/gr/kms.html
const (
	kmsSignTarget  = "TrentService.Sign"
	kmsEndpointFmt = "https://kms.%s.amazonaws.com/"
)

// KMS is used to sign payloads using AWS Key Management Service.
type KMS struct {
	// endpoint returns the region-specifc KMS endpoint.
	// It can be overridden by tests.
	endpoint func(region string) string

	// client is used to send authorization tokens requests.
	client *http.Client

	logger logging.Logger
}

func NewKMS(logger logging.Logger) *KMS {
	return &KMS{
		endpoint: func(region string) string {
			return fmt.Sprintf(kmsEndpointFmt, region)
		},
		client: &http.Client{},
		logger: logger,
	}
}

func NewKMSWithURLClient(url string, client *http.Client, logger logging.Logger) *KMS {
	return &KMS{
		endpoint: func(string) string { return url },
		client:   client,
		logger:   logger,
	}
}

type KMSSignRequest struct {
	KeyID            string `json:"KeyId"`
	Message          string `json:"Message"`
	MessageType      string `json:"MessageType"`
	SigningAlgorithm string `json:"SigningAlgorithm"`
}
type KMSSignResponse struct {
	KeyID            string `json:"KeyId"`
	Signature        string `json:"Signature"`
	SigningAlgorithm string `json:"SigningAlgorithm"`
}

// SignDigest signs a digest using KMS.
func (k *KMS) SignDigest(ctx context.Context, digest []byte, keyID string, signingAlgorithm string, creds Credentials, signatureVersion string) (string, error) {
	endpoint := k.endpoint(creds.RegionName)

	kmsRequest := KMSSignRequest{
		KeyID:            keyID,
		Message:          base64.StdEncoding.EncodeToString(digest),
		MessageType:      "DIGEST",
		SigningAlgorithm: signingAlgorithm,
	}
	requestJSONBytes, err := json.Marshal(kmsRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshall request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(requestJSONBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Amz-Target", kmsSignTarget)
	req.Header.Set("Accept-Encoding", "identity")
	req.Header.Set("Content-Type", "application/x-amz-json-1.1")
	req.Header.Set("User-Agent", version.UserAgent)

	if err := SignRequest(req, "kms", creds, time.Now(), signatureVersion); err != nil {
		return "", fmt.Errorf("failed to sign request: %w", err)
	}

	resp, err := DoRequestWithClient(req, k.client, "kms sign digest", k.logger)
	if err != nil {
		return "", err
	}

	var data KMSSignResponse
	if err := json.Unmarshal(resp, &data); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return data.Signature, nil
}
