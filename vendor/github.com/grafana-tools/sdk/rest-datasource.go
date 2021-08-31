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
	"context"
	"encoding/json"
	"fmt"
)

// GetAllDatasources gets all datasources.
// Reflects GET /api/datasources API call.
func (r *Client) GetAllDatasources(ctx context.Context) ([]Datasource, error) {
	var (
		raw  []byte
		ds   []Datasource
		code int
		err  error
	)
	if raw, code, err = r.get(ctx, "api/datasources", nil); err != nil {
		return nil, err
	}
	if code != 200 {
		return nil, fmt.Errorf("HTTP error %d: returns %s", code, raw)
	}
	err = json.Unmarshal(raw, &ds)
	return ds, err
}

// GetDatasource gets an datasource by ID.
// Reflects GET /api/datasources/:datasourceId API call.
func (r *Client) GetDatasource(ctx context.Context, id uint) (Datasource, error) {
	var (
		raw  []byte
		ds   Datasource
		code int
		err  error
	)
	if raw, code, err = r.get(ctx, fmt.Sprintf("api/datasources/%d", id), nil); err != nil {
		return ds, err
	}
	if code != 200 {
		return ds, fmt.Errorf("HTTP error %d: returns %s", code, raw)
	}
	err = json.Unmarshal(raw, &ds)
	return ds, err
}

// GetDatasourceByName gets an datasource by Name.
// Reflects GET /api/datasources/name/:datasourceName API call.
func (r *Client) GetDatasourceByName(ctx context.Context, name string) (Datasource, error) {
	var (
		raw  []byte
		ds   Datasource
		code int
		err  error
	)
	if raw, code, err = r.get(ctx, fmt.Sprintf("api/datasources/name/%s", name), nil); err != nil {
		return ds, err
	}
	if code != 200 {
		return ds, fmt.Errorf("HTTP error %d: returns %s", code, raw)
	}
	err = json.Unmarshal(raw, &ds)
	return ds, err
}

// CreateDatasource creates a new datasource.
// Reflects POST /api/datasources API call.
func (r *Client) CreateDatasource(ctx context.Context, ds Datasource) (StatusMessage, error) {
	var (
		raw  []byte
		resp StatusMessage
		err  error
	)
	if raw, err = json.Marshal(ds); err != nil {
		return StatusMessage{}, err
	}
	if raw, _, err = r.post(ctx, "api/datasources", nil, raw); err != nil {
		return StatusMessage{}, err
	}
	if err = json.Unmarshal(raw, &resp); err != nil {
		return StatusMessage{}, err
	}
	return resp, nil
}

// UpdateDatasource updates a datasource from data passed in argument.
// Reflects PUT /api/datasources/:datasourceId API call.
func (r *Client) UpdateDatasource(ctx context.Context, ds Datasource) (StatusMessage, error) {
	var (
		raw  []byte
		resp StatusMessage
		err  error
	)
	if raw, err = json.Marshal(ds); err != nil {
		return StatusMessage{}, err
	}
	if raw, _, err = r.put(ctx, fmt.Sprintf("api/datasources/%d", ds.ID), nil, raw); err != nil {
		return StatusMessage{}, err
	}
	if err = json.Unmarshal(raw, &resp); err != nil {
		return StatusMessage{}, err
	}
	return resp, nil
}

// DeleteDatasource deletes an existing datasource by ID.
// Reflects DELETE /api/datasources/:datasourceId API call.
func (r *Client) DeleteDatasource(ctx context.Context, id uint) (StatusMessage, error) {
	var (
		raw   []byte
		reply StatusMessage
		err   error
	)
	if raw, _, err = r.delete(ctx, fmt.Sprintf("api/datasources/%d", id)); err != nil {
		return StatusMessage{}, err
	}
	err = json.Unmarshal(raw, &reply)
	return reply, err
}

// DeleteDatasourceByName deletes an existing datasource by Name.
// Reflects DELETE /api/datasources/name/:datasourceName API call.
func (r *Client) DeleteDatasourceByName(ctx context.Context, name string) (StatusMessage, error) {
	var (
		raw   []byte
		reply StatusMessage
		err   error
	)
	if raw, _, err = r.delete(ctx, fmt.Sprintf("api/datasources/name/%s", name)); err != nil {
		return StatusMessage{}, err
	}
	err = json.Unmarshal(raw, &reply)
	return reply, err
}

// GetDatasourceTypes gets all available plugins for the datasources.
// Reflects GET /api/datasources/plugins API call.
func (r *Client) GetDatasourceTypes(ctx context.Context) (map[string]DatasourceType, error) {
	var (
		raw     []byte
		dsTypes = make(map[string]DatasourceType)
		code    int
		err     error
	)
	if raw, code, err = r.get(ctx, "api/datasources/plugins", nil); err != nil {
		return nil, err
	}
	if code != 200 {
		return nil, fmt.Errorf("HTTP error %d: returns %s", code, raw)
	}
	err = json.Unmarshal(raw, &dsTypes)
	return dsTypes, err
}
