package alerting

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
)
import "github.com/prometheus/client_golang/api"

const (
	apiPrefix      = "/api/v1"
	epRules        = apiPrefix + "/rules"
	statusAPIError = 422

	statusSuccess status = "success"
	statusError   status = "error"

	ErrBadData     ErrorType = "bad_data"
	ErrTimeout     ErrorType = "timeout"
	ErrCanceled    ErrorType = "canceled"
	ErrExec        ErrorType = "execution"
	ErrBadResponse ErrorType = "bad_response"
	ErrServer      ErrorType = "server_error"
	ErrClient      ErrorType = "client_error"
)

type status string

type ErrorType string

type Error struct {
	Type   ErrorType
	Msg    string
	Detail string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Type, e.Msg)
}

type response struct {
	Status    status          `json:"status"`
	Data      json.RawMessage `json:"data,omitempty"`
	ErrorType ErrorType       `json:"errorType,omitempty"`
	Error     string          `json:"error,omitempty"`
	Warnings  []string        `json:"warnings,omitempty"`
}

type RuleClient interface {
	PrometheusRules(ctx context.Context) ([]*RuleGroup, error)
	ThanosRules(ctx context.Context) ([]*RuleGroup, error)
}

type ruleClient struct {
	prometheus  api.Client
	thanosruler api.Client
}

func (c *ruleClient) PrometheusRules(ctx context.Context) ([]*RuleGroup, error) {
	if c.prometheus != nil {
		return c.rules(c.prometheus, ctx)
	}
	return nil, nil
}

func (c *ruleClient) ThanosRules(ctx context.Context) ([]*RuleGroup, error) {
	if c.thanosruler != nil {
		return c.rules(c.thanosruler, ctx)
	}
	return nil, nil
}

func (c *ruleClient) rules(client api.Client, ctx context.Context) ([]*RuleGroup, error) {
	u := client.URL(epRules, nil)
	q := u.Query()
	q.Add("type", "alert")
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "error creating request: ")
	}

	_, body, _, err := c.do(client, ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "error doing request: ")
	}

	var result struct {
		Groups []*RuleGroup
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	return result.Groups, nil
}

func (c *ruleClient) do(client api.Client, ctx context.Context, req *http.Request) (*http.Response, []byte, []string, error) {
	resp, body, e := client.Do(ctx, req)
	if e != nil {
		return resp, body, nil, e
	}

	code := resp.StatusCode

	if code/100 != 2 && !apiError(code) {
		errorType, errorMsg := errorTypeAndMsgFor(resp)
		return resp, body, nil, &Error{
			Type:   errorType,
			Msg:    errorMsg,
			Detail: string(body),
		}
	}

	var result response
	if http.StatusNoContent != code {
		if jsonErr := json.Unmarshal(body, &result); jsonErr != nil {
			return resp, body, nil, &Error{
				Type: ErrBadResponse,
				Msg:  jsonErr.Error(),
			}
		}
	}

	var err error
	if apiError(code) && result.Status == "success" {
		err = &Error{
			Type: ErrBadResponse,
			Msg:  "inconsistent body for response code",
		}
	}
	if result.Status == "error" {
		err = &Error{
			Type: result.ErrorType,
			Msg:  result.Error,
		}
	}

	return resp, []byte(result.Data), result.Warnings, err
}

func errorTypeAndMsgFor(resp *http.Response) (ErrorType, string) {
	switch resp.StatusCode / 100 {
	case 4:
		return ErrClient, fmt.Sprintf("client error: %d", resp.StatusCode)
	case 5:
		return ErrServer, fmt.Sprintf("server error: %d", resp.StatusCode)
	}
	return ErrBadResponse, fmt.Sprintf("bad response code %d", resp.StatusCode)
}

func apiError(code int) bool {
	// These are the codes that rule server sends when it returns an error.
	return code == statusAPIError || code == http.StatusBadRequest ||
		code == http.StatusServiceUnavailable || code == http.StatusInternalServerError
}

func NewRuleClient(options *Options) (RuleClient, error) {
	var (
		c ruleClient
		e error
	)
	if options.PrometheusEndpoint != "" {
		c.prometheus, e = api.NewClient(api.Config{Address: options.PrometheusEndpoint})
	}
	if options.ThanosRulerEndpoint != "" {
		c.thanosruler, e = api.NewClient(api.Config{Address: options.ThanosRulerEndpoint})
	}
	return &c, e
}
