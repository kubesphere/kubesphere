package cas

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"github.com/golang/glog"
)

// NewServiceTicketValidator create a new *ServiceTicketValidator
func NewServiceTicketValidator(client *http.Client, casURL *url.URL) *ServiceTicketValidator {
	return &ServiceTicketValidator{
		client: client,
		casURL: casURL,
	}
}

// ServiceTicketValidator is responsible for the validation of a service ticket
type ServiceTicketValidator struct {
	client *http.Client
	casURL *url.URL
}

// ValidateTicket validates the service ticket for the given server. The method will try to use the service validate
// endpoint of the cas >= 2 protocol, if the service validate endpoint not available, the function will use the cas 1
// validate endpoint.
func (validator *ServiceTicketValidator) ValidateTicket(serviceURL *url.URL, ticket string) (*AuthenticationResponse, error) {
	if glog.V(2) {
		glog.Infof("Validating ticket %v for service %v", ticket, serviceURL)
	}

	u, err := validator.ServiceValidateUrl(serviceURL, ticket)
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	r.Header.Add("User-Agent", "Golang CAS client gopkg.in/cas")

	if glog.V(2) {
		glog.Infof("Attempting ticket validation with %v", r.URL)
	}

	resp, err := validator.client.Do(r)
	if err != nil {
		return nil, err
	}

	if glog.V(2) {
		glog.Infof("Request %v %v returned %v",
			r.Method, r.URL,
			resp.Status)
	}

	if resp.StatusCode == http.StatusNotFound {
		return validator.validateTicketCas1(serviceURL, ticket)
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cas: validate ticket: %v", string(body))
	}

	if glog.V(2) {
		glog.Infof("Received authentication response\n%v", string(body))
	}

	success, err := ParseServiceResponse(body)
	if err != nil {
		return nil, err
	}

	if glog.V(2) {
		glog.Infof("Parsed ServiceResponse: %#v", success)
	}

	return success, nil
}

// ServiceValidateUrl creates the service validation url for the cas >= 2 protocol.
// TODO the function is only exposed, because of the clients ServiceValidateUrl function
func (validator *ServiceTicketValidator) ServiceValidateUrl(serviceURL *url.URL, ticket string) (string, error) {
	u, err := validator.casURL.Parse(path.Join(validator.casURL.Path, "serviceValidate"))
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Add("service", sanitisedURLString(serviceURL))
	q.Add("ticket", ticket)
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func (validator *ServiceTicketValidator) validateTicketCas1(serviceURL *url.URL, ticket string) (*AuthenticationResponse, error) {
	u, err := validator.ValidateUrl(serviceURL, ticket)
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	r.Header.Add("User-Agent", "Golang CAS client gopkg.in/cas")

	if glog.V(2) {
		glog.Infof("Attempting ticket validation with %v", r.URL)
	}

	resp, err := validator.client.Do(r)
	if err != nil {
		return nil, err
	}

	if glog.V(2) {
		glog.Infof("Request %v %v returned %v",
			r.Method, r.URL,
			resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	if err != nil {
		return nil, err
	}

	body := string(data)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cas: validate ticket: %v", body)
	}

	if glog.V(2) {
		glog.Infof("Received authentication response\n%v", body)
	}

	if body == "no\n\n" {
		return nil, nil // not logged in
	}

	success := &AuthenticationResponse{
		User: body[4 : len(body)-1],
	}

	if glog.V(2) {
		glog.Infof("Parsed ServiceResponse: %#v", success)
	}

	return success, nil
}

// ValidateUrl creates the validation url for the cas >= 1 protocol.
// TODO the function is only exposed, because of the clients ValidateUrl function
func (validator *ServiceTicketValidator) ValidateUrl(serviceURL *url.URL, ticket string) (string, error) {
	u, err := validator.casURL.Parse(path.Join(validator.casURL.Path, "validate"))
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Add("service", sanitisedURLString(serviceURL))
	q.Add("ticket", ticket)
	u.RawQuery = q.Encode()

	return u.String(), nil
}
