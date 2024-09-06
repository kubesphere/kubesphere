package aws

import (
	"errors"
	"io"
	"net/http"

	"github.com/open-policy-agent/opa/logging"
)

// DoRequestWithClient is a convenience function to get the body of an http response with
// appropriate error-handling boilerplate and logging.
func DoRequestWithClient(req *http.Request, client *http.Client, desc string, logger logging.Logger) ([]byte, error) {
	resp, err := client.Do(req)
	if err != nil {
		// some kind of catastrophe talking to the service
		return nil, errors.New(desc + " HTTP request failed: " + err.Error())
	}
	defer resp.Body.Close()

	logger.WithFields(map[string]interface{}{
		"url":     req.URL.String(),
		"status":  resp.Status,
		"headers": resp.Header,
	}).Debug("Received response from " + desc + " service.")

	if resp.StatusCode != 200 {
		if logger.GetLevel() == logging.Debug {
			body, err := io.ReadAll(resp.Body)
			if err == nil {
				logger.Debug("Error response with response body: %s", body)
			}
		}
		// could be 404 for role that's not available, but cover all the bases
		return nil, errors.New(desc + " HTTP request returned unexpected status: " + resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// deal with problems reading the body, whatever those might be
		return nil, errors.New(desc + " HTTP response body could not be read: " + err.Error())
	}

	return body, nil
}
