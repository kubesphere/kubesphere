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

// AlertNotification as described in the doc
// https://grafana.com/docs/grafana/latest/http_api/alerting_notification_channels/
type AlertNotification struct {
	ID                    int64       `json:"id,omitempty"`
	Name                  string      `json:"name"`
	Type                  string      `json:"type"`
	IsDefault             bool        `json:"isDefault"`
	DisableResolveMessage bool        `json:"disableResolveMessage"`
	SendReminder          bool        `json:"sendReminder"`
	Frequency             string      `json:"frequency"`
	Settings              interface{} `json:"settings"`
	UID                   string      `json:"uid,omitempty"`
}
