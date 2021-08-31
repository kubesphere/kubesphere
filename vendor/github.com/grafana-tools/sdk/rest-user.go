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
	"net/url"
)

// GetActualUser gets an actual user.
// Reflects GET /api/user API call.
func (r *Client) GetActualUser(ctx context.Context) (User, error) {
	var (
		raw  []byte
		user User
		code int
		err  error
	)
	if raw, code, err = r.get(ctx, "api/user", nil); err != nil {
		return user, err
	}
	if code != 200 {
		return user, fmt.Errorf("HTTP error %d: returns %s", code, raw)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&user); err != nil {
		return user, fmt.Errorf("unmarshal user: %s\n%s", err, raw)
	}
	return user, err
}

// GetUser gets an user by ID.
// Reflects GET /api/users/:id API call.
func (r *Client) GetUser(ctx context.Context, id uint) (User, error) {
	var (
		raw  []byte
		user User
		code int
		err  error
	)
	if raw, code, err = r.get(ctx, fmt.Sprintf("api/users/%d", id), nil); err != nil {
		return user, err
	}
	if code != 200 {
		return user, fmt.Errorf("HTTP error %d: returns %s", code, raw)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&user); err != nil {
		return user, fmt.Errorf("unmarshal user: %s\n%s", err, raw)
	}
	return user, err
}

// GetAllUsers gets all users.
// Reflects GET /api/users API call.
func (r *Client) GetAllUsers(ctx context.Context) ([]User, error) {
	var (
		raw   []byte
		users []User
		code  int
		err   error
	)

	params := url.Values{}
	params.Set("perpage", "99999")
	if raw, code, err = r.get(ctx, "api/users", params); err != nil {
		return users, err
	}
	if code != 200 {
		return users, fmt.Errorf("HTTP error %d: returns %s", code, raw)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&users); err != nil {
		return users, fmt.Errorf("unmarshal users: %s\n%s", err, raw)
	}
	return users, err
}

// SearchUsersWithPaging search users with paging.
// query optional.  query value is contained in one of the name, login or email fields. Query values with spaces need to be url encoded e.g. query=Jane%20Doe
// perpage optional. default 1000
// page optional. default 1
// http://docs.grafana.org/http_api/user/#search-users
// http://docs.grafana.org/http_api/user/#search-users-with-paging
//
// Reflects GET /api/users/search API call.
func (r *Client) SearchUsersWithPaging(ctx context.Context, query *string, perpage, page *int) (PageUsers, error) {
	var (
		raw       []byte
		pageUsers PageUsers
		code      int
		err       error
	)

	var params url.Values = nil
	if perpage != nil && page != nil {
		if params == nil {
			params = url.Values{}
		}
		params["perpage"] = []string{fmt.Sprint(*perpage)}
		params["page"] = []string{fmt.Sprint(*page)}
	}

	if query != nil {
		if params == nil {
			params = url.Values{}
		}
		params["query"] = []string{*query}
	}

	if raw, code, err = r.get(ctx, "api/users/search", params); err != nil {
		return pageUsers, err
	}
	if code != 200 {
		return pageUsers, fmt.Errorf("HTTP error %d: returns %s", code, raw)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&pageUsers); err != nil {
		return pageUsers, fmt.Errorf("unmarshal users: %s\n%s", err, raw)
	}
	return pageUsers, err
}

// SwitchActualUserContext switches current user context to the given organization.
// Reflects POST /api/user/using/:organizationId API call.
func (r *Client) SwitchActualUserContext(ctx context.Context, oid uint) (StatusMessage, error) {
	var (
		raw  []byte
		resp StatusMessage
		err  error
	)

	if raw, _, err = r.post(ctx, fmt.Sprintf("/api/user/using/%d", oid), nil, raw); err != nil {
		return StatusMessage{}, err
	}
	if err = json.Unmarshal(raw, &resp); err != nil {
		return StatusMessage{}, err
	}
	return resp, nil
}
