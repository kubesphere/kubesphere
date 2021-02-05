package cas

import (
	"crypto/rand"
	"encoding/xml"
	"strings"
	"time"
)

// Represents the XML CAS Single Log Out Request data
type logoutRequest struct {
	XMLName         xml.Name  `xml:"urn:oasis:names:tc:SAML:2.0:protocol LogoutRequest"`
	Version         string    `xml:"Version,attr"`
	IssueInstant    time.Time `xml:"-"`
	RawIssueInstant string    `xml:"IssueInstant,attr"`
	ID              string    `xml:"ID,attr"`
	NameID          string    `xml:"urn:oasis:names:tc:SAML:2.0:assertion NameID"`
	SessionIndex    string    `xml:"SessionIndex"`
}

func parseLogoutRequest(data []byte) (*logoutRequest, error) {
	l := &logoutRequest{}
	if err := xml.Unmarshal(data, &l); err != nil {
		return nil, err
	}

	t, err := parseDate(l.RawIssueInstant)
	if err != nil {
		return nil, err
	}

	l.IssueInstant = t
	l.NameID = strings.TrimSpace(l.NameID)
	l.SessionIndex = strings.TrimSpace(l.SessionIndex)

	return l, nil
}

func parseDate(raw string) (time.Time, error) {
	t, err := time.Parse(time.RFC1123Z, raw)
	if err != nil {
		// if RFC1123Z does not match, we will try iso8601
		t, err = time.Parse("2006-01-02T15:04:05Z0700", raw)
		if err != nil {
			return t, err
		}
	}
	return t, nil
}

func newLogoutRequestID() string {
	const alphabet = "abcdef0123456789"

	// generate 64 character string
	bytes := make([]byte, 64)
	rand.Read(bytes)

	for k, v := range bytes {
		bytes[k] = alphabet[v%byte(len(alphabet))]
	}

	return string(bytes)
}

func xmlLogoutRequest(ticket string) ([]byte, error) {
	l := &logoutRequest{
		Version:      "2.0",
		IssueInstant: time.Now().UTC(),
		ID:           newLogoutRequestID(),
		NameID:       "@NOT_USED@",
		SessionIndex: ticket,
	}

	l.RawIssueInstant = l.IssueInstant.Format(time.RFC1123Z)

	return xml.MarshalIndent(l, "", "  ")
}
