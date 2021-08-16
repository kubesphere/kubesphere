package sdk

/*
   Copyright 2016 Alexander I.Grafov <grafov@gmail.com>
   Copyright 2016-2019 The Grafana SDK authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

	   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.

   ॐ तारे तुत्तारे तुरे स्व
*/

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// CreateOrg creates a new organization.
// It reflects POST /api/orgs API call.
func (r *Client) CreateOrg(ctx context.Context, org Org) (StatusMessage, error) {
	var (
		raw  []byte
		resp StatusMessage
		err  error
	)
	if raw, err = json.Marshal(org); err != nil {
		return StatusMessage{}, err
	}
	if raw, _, err = r.post(ctx, "api/orgs", nil, raw); err != nil {
		return StatusMessage{}, err
	}
	if err = json.Unmarshal(raw, &resp); err != nil {
		return StatusMessage{}, err
	}
	return resp, nil
}

// GetAllOrgs returns all organizations.
// It reflects GET /api/orgs API call.
func (r *Client) GetAllOrgs(ctx context.Context) ([]Org, error) {
	var (
		raw  []byte
		orgs []Org
		code int
		err  error
	)
	if raw, code, err = r.get(ctx, "api/orgs", nil); err != nil {
		return orgs, err
	}

	if code != http.StatusOK {
		return orgs, fmt.Errorf("HTTP error %d: returns %s", code, raw)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&orgs); err != nil {
		return orgs, fmt.Errorf("unmarshal orgs: %s\n%s", err, raw)
	}
	return orgs, err
}

// GetActualOrg gets current organization.
// It reflects GET /api/org API call.
func (r *Client) GetActualOrg(ctx context.Context) (Org, error) {
	var (
		raw  []byte
		org  Org
		code int
		err  error
	)
	if raw, code, err = r.get(ctx, "api/org", nil); err != nil {
		return org, err
	}
	if code != http.StatusOK {
		return org, fmt.Errorf("HTTP error %d: returns %s", code, raw)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&org); err != nil {
		return org, fmt.Errorf("unmarshal org: %s\n%s", err, raw)
	}
	return org, err
}

// GetOrgById gets organization by organization Id.
// It reflects GET /api/orgs/:orgId API call.
func (r *Client) GetOrgById(ctx context.Context, oid uint) (Org, error) {
	var (
		raw  []byte
		org  Org
		code int
		err  error
	)
	if raw, code, err = r.get(ctx, fmt.Sprintf("api/orgs/%d", oid), nil); err != nil {
		return org, err
	}

	if code != http.StatusOK {
		return org, fmt.Errorf("HTTP error %d: returns %s", code, raw)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&org); err != nil {
		return org, fmt.Errorf("unmarshal org: %s\n%s", err, raw)
	}
	return org, err
}

// GetOrgByOrgName gets organization by organization name.
// It reflects GET /api/orgs/name/:orgName API call.
func (r *Client) GetOrgByOrgName(ctx context.Context, name string) (Org, error) {
	var (
		raw  []byte
		org  Org
		code int
		err  error
	)
	if raw, code, err = r.get(ctx, fmt.Sprintf("api/orgs/name/%s", name), nil); err != nil {
		return org, err
	}

	if code != http.StatusOK {
		return org, fmt.Errorf("HTTP error %d: returns %s", code, raw)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&org); err != nil {
		return org, fmt.Errorf("unmarshal org: %s\n%s", err, raw)
	}
	return org, err
}

// UpdateActualOrg updates current organization.
// It reflects PUT /api/org API call.
func (r *Client) UpdateActualOrg(ctx context.Context, org Org) (StatusMessage, error) {
	var (
		raw  []byte
		resp StatusMessage
		err  error
	)
	if raw, err = json.Marshal(org); err != nil {
		return StatusMessage{}, err
	}
	if raw, _, err = r.put(ctx, "api/org", nil, raw); err != nil {
		return StatusMessage{}, err
	}
	if err = json.Unmarshal(raw, &resp); err != nil {
		return StatusMessage{}, err
	}
	return resp, nil
}

// UpdateOrg updates the organization identified by oid.
// It reflects PUT /api/orgs/:orgId API call.
func (r *Client) UpdateOrg(ctx context.Context, org Org, oid uint) (StatusMessage, error) {
	var (
		raw  []byte
		resp StatusMessage
		err  error
	)
	if raw, err = json.Marshal(org); err != nil {
		return StatusMessage{}, err
	}
	if raw, _, err = r.put(ctx, fmt.Sprintf("api/orgs/%d", oid), nil, raw); err != nil {
		return StatusMessage{}, err
	}
	if err = json.Unmarshal(raw, &resp); err != nil {
		return StatusMessage{}, err
	}
	return resp, nil
}

// DeleteOrg deletes the organization identified by the oid.
// Reflects DELETE /api/orgs/:orgId API call.
func (r *Client) DeleteOrg(ctx context.Context, oid uint) (StatusMessage, error) {
	var (
		raw  []byte
		resp StatusMessage
		err  error
	)
	if raw, _, err = r.delete(ctx, fmt.Sprintf("api/orgs/%d", oid)); err != nil {
		return StatusMessage{}, err
	}
	if err = json.Unmarshal(raw, &resp); err != nil {
		return StatusMessage{}, err
	}
	return resp, nil
}

// GetActualOrgUsers get all users within the actual organisation.
// Reflects GET /api/org/users API call.
func (r *Client) GetActualOrgUsers(ctx context.Context) ([]OrgUser, error) {
	var (
		raw   []byte
		users []OrgUser
		code  int
		err   error
	)
	if raw, code, err = r.get(ctx, "api/org/users", nil); err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("HTTP error %d: returns %s", code, raw)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&users); err != nil {
		return nil, fmt.Errorf("unmarshal org: %s\n%s", err, raw)
	}
	return users, err
}

// GetOrgUsers gets the users for the organization specified by oid.
// Reflects GET /api/orgs/:orgId/users API call.
func (r *Client) GetOrgUsers(ctx context.Context, oid uint) ([]OrgUser, error) {
	var (
		raw   []byte
		users []OrgUser
		code  int
		err   error
	)
	if raw, code, err = r.get(ctx, fmt.Sprintf("api/orgs/%d/users", oid), nil); err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("HTTP error %d: returns %s", code, raw)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&users); err != nil {
		return nil, fmt.Errorf("unmarshal org: %s\n%s", err, raw)
	}
	return users, err
}

// AddActualOrgUser adds a global user to the current organization.
// Reflects POST /api/org/users API call.
func (r *Client) AddActualOrgUser(ctx context.Context, userRole UserRole) (StatusMessage, error) {
	var (
		raw  []byte
		resp StatusMessage
		err  error
	)
	if raw, err = json.Marshal(userRole); err != nil {
		return StatusMessage{}, err
	}
	if raw, _, err = r.post(ctx, "api/org/users", nil, raw); err != nil {
		return StatusMessage{}, err
	}
	if err = json.Unmarshal(raw, &resp); err != nil {
		return StatusMessage{}, err
	}
	return resp, nil
}

// UpdateActualOrgUser updates the existing user.
// Reflects POST /api/org/users/:userId API call.
func (r *Client) UpdateActualOrgUser(ctx context.Context, user UserRole, uid uint) (StatusMessage, error) {
	var (
		raw  []byte
		resp StatusMessage
		err  error
	)
	if raw, err = json.Marshal(user); err != nil {
		return StatusMessage{}, err
	}
	if raw, _, err = r.post(ctx, fmt.Sprintf("api/org/users/%d", uid), nil, raw); err != nil {
		return StatusMessage{}, err
	}
	if err = json.Unmarshal(raw, &resp); err != nil {
		return StatusMessage{}, err
	}
	return resp, nil
}

// DeleteActualOrgUser delete user in actual organization.
// Reflects DELETE /api/org/users/:userId API call.
func (r *Client) DeleteActualOrgUser(ctx context.Context, uid uint) (StatusMessage, error) {
	var (
		raw   []byte
		reply StatusMessage
		err   error
	)
	if raw, _, err = r.delete(ctx, fmt.Sprintf("api/org/users/%d", uid)); err != nil {
		return StatusMessage{}, err
	}
	err = json.Unmarshal(raw, &reply)
	return reply, err
}

// AddOrgUser add user to organization with oid.
// Reflects POST /api/orgs/:orgId/users API call.
func (r *Client) AddOrgUser(ctx context.Context, user UserRole, oid uint) (StatusMessage, error) {
	var (
		raw   []byte
		reply StatusMessage
		err   error
	)
	if raw, err = json.Marshal(user); err != nil {
		return StatusMessage{}, err
	}
	if raw, _, err = r.post(ctx, fmt.Sprintf("api/orgs/%d/users", oid), nil, raw); err != nil {
		return StatusMessage{}, err
	}
	err = json.Unmarshal(raw, &reply)
	return reply, err
}

// UpdateOrgUser updates the user specified by uid within the organization specified by oid.
// Reflects PATCH /api/orgs/:orgId/users/:userId API call.
func (r *Client) UpdateOrgUser(ctx context.Context, user UserRole, oid, uid uint) (StatusMessage, error) {
	var (
		raw   []byte
		reply StatusMessage
		err   error
	)
	if raw, err = json.Marshal(user); err != nil {
		return StatusMessage{}, err
	}
	if raw, _, err = r.patch(ctx, fmt.Sprintf("api/orgs/%d/users/%d", oid, uid), nil, raw); err != nil {
		return StatusMessage{}, err
	}
	err = json.Unmarshal(raw, &reply)
	return reply, err
}

// DeleteOrgUser deletes the user specified by uid within the organization specified by oid.
// Reflects DELETE /api/orgs/:orgId/users/:userId API call.
func (r *Client) DeleteOrgUser(ctx context.Context, oid, uid uint) (StatusMessage, error) {
	var (
		raw   []byte
		reply StatusMessage
		err   error
	)
	if raw, _, err = r.delete(ctx, fmt.Sprintf("api/orgs/%d/users/%d", oid, uid)); err != nil {
		return StatusMessage{}, err
	}
	err = json.Unmarshal(raw, &reply)
	return reply, err
}

// UpdateActualOrgPreferences updates preferences of the actual organization.
// Reflects PUT /api/org/preferences API call.
func (r *Client) UpdateActualOrgPreferences(ctx context.Context, prefs Preferences) (StatusMessage, error) {
	var (
		raw  []byte
		resp StatusMessage
		err  error
	)
	if raw, err = json.Marshal(prefs); err != nil {
		return StatusMessage{}, err
	}
	if raw, _, err = r.put(ctx, "api/org/preferences/", nil, raw); err != nil {
		return StatusMessage{}, err
	}
	if err = json.Unmarshal(raw, &resp); err != nil {
		return StatusMessage{}, err
	}
	return resp, nil
}

// GetActualOrgPreferences gets preferences of the actual organization.
// It reflects GET /api/org/preferences API call.
func (r *Client) GetActualOrgPreferences(ctx context.Context) (Preferences, error) {
	var (
		raw  []byte
		pref Preferences
		code int
		err  error
	)
	if raw, code, err = r.get(ctx, "/api/org/preferences", nil); err != nil {
		return pref, err
	}

	if code != http.StatusOK {
		return pref, fmt.Errorf("HTTP error %d: returns %s", code, raw)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&pref); err != nil {
		return pref, fmt.Errorf("unmarshal prefs: %s\n%s", err, raw)
	}
	return pref, err
}

// UpdateActualOrgAddress updates current organization's address.
// It reflects PUT /api/org/address API call.
func (r *Client) UpdateActualOrgAddress(ctx context.Context, address Address) (StatusMessage, error) {
	var (
		raw  []byte
		resp StatusMessage
		err  error
	)
	if raw, err = json.Marshal(address); err != nil {
		return StatusMessage{}, err
	}
	if raw, _, err = r.put(ctx, "api/org/address", nil, raw); err != nil {
		return StatusMessage{}, err
	}
	if err = json.Unmarshal(raw, &resp); err != nil {
		return StatusMessage{}, err
	}
	return resp, nil
}

// UpdateOrgAddress updates the address of the organization identified by oid.
// It reflects PUT /api/orgs/:orgId/address API call.
func (r *Client) UpdateOrgAddress(ctx context.Context, address Address, oid uint) (StatusMessage, error) {
	var (
		raw  []byte
		resp StatusMessage
		err  error
	)
	if raw, err = json.Marshal(address); err != nil {
		return StatusMessage{}, err
	}
	if raw, _, err = r.put(ctx, fmt.Sprintf("api/orgs/%d/address", oid), nil, raw); err != nil {
		return StatusMessage{}, err
	}
	if err = json.Unmarshal(raw, &resp); err != nil {
		return StatusMessage{}, err
	}
	return resp, nil
}
