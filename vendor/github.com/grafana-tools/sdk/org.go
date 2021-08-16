package sdk

/*
   Copyright 2016-2017 Alexander I.Grafov <grafov@gmail.com>

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

type Org struct {
	ID      uint    `json:"id"`
	Name    string  `json:"name"`
	Address Address `json:"address"`
}

type OrgUser struct {
	ID    uint   `json:"userId"`
	OrgId uint   `json:"orgId"`
	Email string `json:"email"`
	Login string `json:"login"`
	Role  string `json:"role"`
}
