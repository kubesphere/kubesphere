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

// Folder as described in the doc
// https://grafana.com/docs/grafana/latest/http_api/folder/#get-all-folders
type Folder struct {
	ID        int    `json:"id"`
	UID       string `json:"uid"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	HasAcl    bool   `json:"hasAcl"`
	CanSave   bool   `json:"canSave"`
	CanEdit   bool   `json:"canEdit"`
	CanAdmin  bool   `json:"canAdmin"`
	CreatedBy string `json:"createdBy"`
	Created   string `json:"created"`
	UpdatedBy string `json:"updatedBy"`
	Updated   string `json:"updated"`
	Version   int    `json:"version"`
	Overwrite bool   `json:"overwrite"`
}
