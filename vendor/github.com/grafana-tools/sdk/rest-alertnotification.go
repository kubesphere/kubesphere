package sdk

/*
   Copyright 2016-2020 The Grafana SDK authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

	   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

import (
	"context"
	"encoding/json"
	"fmt"
)

// GetAllAlertNotifications gets all alert notification channels.
// Reflects GET /api/alert-notifications API call.
func (c *Client) GetAllAlertNotifications(ctx context.Context) ([]AlertNotification, error) {
	var (
		raw  []byte
		an   []AlertNotification
		code int
		err  error
	)
	if raw, code, err = c.get(ctx, "api/alert-notifications", nil); err != nil {
		return nil, err
	}
	if code != 200 {
		return nil, fmt.Errorf("HTTP error %d: returns %s", code, raw)
	}
	err = json.Unmarshal(raw, &an)
	return an, err
}

// GetAlertNotificationUID gets the alert notification channel which has the specified uid.
// Reflects GET /api/alert-notifications/uid/:uid API call.
func (c *Client) GetAlertNotificationUID(ctx context.Context, uid string) (AlertNotification, error) {
	var (
		raw  []byte
		an   AlertNotification
		code int
		err  error
	)
	if raw, code, err = c.get(ctx, fmt.Sprintf("api/alert-notifications/uid/%s", uid), nil); err != nil {
		return an, err
	}
	if code != 200 {
		return an, fmt.Errorf("HTTP error %d: returns %s", code, raw)
	}
	err = json.Unmarshal(raw, &an)
	return an, err
}

// GetAlertNotificationID gets the alert notification channel which has the specified id.
// Reflects GET /api/alert-notifications/:id API call.
func (c *Client) GetAlertNotificationID(ctx context.Context, id uint) (AlertNotification, error) {
	var (
		raw  []byte
		an   AlertNotification
		code int
		err  error
	)
	if raw, code, err = c.get(ctx, fmt.Sprintf("api/alert-notifications/%d", id), nil); err != nil {
		return an, err
	}
	if code != 200 {
		return an, fmt.Errorf("HTTP error %d: returns %s", code, raw)
	}
	err = json.Unmarshal(raw, &an)
	return an, err
}

// CreateAlertNotification creates a new alert notification channel.
// Reflects POST /api/alert-notifications API call.
func (c *Client) CreateAlertNotification(ctx context.Context, an AlertNotification) (int64, error) {
	var (
		raw  []byte
		code int
		err  error
	)
	if raw, err = json.Marshal(an); err != nil {
		return -1, err
	}
	if raw, code, err = c.post(ctx, "api/alert-notifications", nil, raw); err != nil {
		return -1, err
	}
	if code != 200 {
		return -1, fmt.Errorf("HTTP error %d: returns %s", code, raw)
	}
	result := struct {
		ID int64 `json:"id"`
	}{}
	err = json.Unmarshal(raw, &result)
	return result.ID, err
}

// UpdateAlertNotificationUID updates the specified alert notification channel.
// Reflects PUT /api/alert-notifications/uid/:uid API call.
func (c *Client) UpdateAlertNotificationUID(ctx context.Context, an AlertNotification, uid string) error {
	var (
		raw  []byte
		code int
		err  error
	)
	if raw, err = json.Marshal(an); err != nil {
		return err
	}
	if raw, code, err = c.put(ctx, fmt.Sprintf("api/alert-notifications/uid/%s", uid), nil, raw); err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("HTTP error %d: returns %s", code, raw)
	}
	return nil
}

// UpdateAlertNotificationID updates the specified alert notification channel.
// Reflects PUT /api/alert-notifications/:id API call.
func (c *Client) UpdateAlertNotificationID(ctx context.Context, an AlertNotification, id uint) error {
	var (
		raw  []byte
		code int
		err  error
	)
	if raw, err = json.Marshal(an); err != nil {
		return err
	}
	if raw, code, err = c.put(ctx, fmt.Sprintf("api/alert-notifications/%d", id), nil, raw); err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("HTTP error %d: returns %s", code, raw)
	}
	return nil
}

// DeleteAlertNotificationUID deletes the specified alert notification channel.
// Reflects DELETE /api/alert-notifications/uid/:uid API call.
func (c *Client) DeleteAlertNotificationUID(ctx context.Context, uid string) error {
	var (
		raw  []byte
		code int
		err  error
	)
	if raw, code, err = c.delete(ctx, fmt.Sprintf("api/alert-notifications/uid/%s", uid)); err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("HTTP error %d: returns %s", code, raw)
	}
	return nil
}

// DeleteAlertNotificationID deletes the specified alert notification channel.
// Reflects DELETE /api/alert-notifications/:id API call.
func (c *Client) DeleteAlertNotificationID(ctx context.Context, id uint) error {
	var (
		raw  []byte
		code int
		err  error
	)
	if raw, code, err = c.delete(ctx, fmt.Sprintf("api/alert-notifications/%d", id)); err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("HTTP error %d: returns %s", code, raw)
	}
	return nil
}
