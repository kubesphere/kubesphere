/*
Copyright 2019 The KubeSphere Authors.

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

package params

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"gotest.tools/assert"

	"github.com/emicklei/go-restful"
)

func TestParseConditions(t *testing.T) {
	type args struct {
		req *restful.Request
	}
	tests := []struct {
		name    string
		args    args
		want    *Conditions
		wantErr bool
	}{
		{
			"good case 1",
			args{&restful.Request{Request: &http.Request{URL: &url.URL{
				RawQuery: "conditions=status%3Ddraft%7Cactive%7Csuspended%7Cpassed",
			}}}},
			&Conditions{
				Match: map[string]string{
					"status": "draft|active|suspended|passed",
				},
				Fuzzy: map[string]string{},
			},
			false,
		},
		{
			"good case 2",
			args{&restful.Request{Request: &http.Request{URL: &url.URL{
				RawQuery: "conditions=status1%3Ddraft%7Cactive%7Csuspended%7Cpassed,status2~draft%7Cactive,status3",
			}}}},
			&Conditions{
				Match: map[string]string{
					"status1": "draft|active|suspended|passed",
					"status3": "",
				},
				Fuzzy: map[string]string{
					"status2": "draft|active",
				},
			},
			false,
		},
		{
			"good case 3",
			args{&restful.Request{Request: &http.Request{URL: &url.URL{
				RawQuery: "conditions=status%3Ddraft%7Cactive%7Csuspended%7Cpassed%28%29,",
			}}}},
			&Conditions{
				Match: map[string]string{
					"status": "draft|active|suspended|passed()",
				},
				Fuzzy: map[string]string{},
			},
			false,
		},

		{
			"bad case 1",
			args{&restful.Request{Request: &http.Request{URL: &url.URL{
				RawQuery: "conditions=%28select+status%3D%29",
			}}}},
			nil,
			true,
		},
		{
			"bad case 2",
			args{&restful.Request{Request: &http.Request{URL: &url.URL{
				RawQuery: "conditions=%28select+status%3D%2C",
			}}}},
			nil,
			true,
		},
		{
			"bad case 3",
			args{&restful.Request{Request: &http.Request{URL: &url.URL{
				RawQuery: "conditions=status%3D%2C%28select+status%3D%29",
			}}}},
			nil,
			true,
		},
		{
			"bad case 4",
			args{&restful.Request{Request: &http.Request{URL: &url.URL{
				RawQuery: "conditions=%28select+status%3Ddraft%7Cactive%7Csuspended%7Cpassed%29",
			}}}},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseConditions(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseConditions() error = %+v, wantErr %+v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseConditions() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func Test_parseConditions(t *testing.T) {
	type args struct {
		conditionsStr string
	}
	tests := []struct {
		name    string
		args    args
		want    *Conditions
		wantErr bool
	}{
		{
			"good case 1",
			args{"key1=value1,key2~value2,key3=,key4~,key5"},
			&Conditions{
				Match: map[string]string{
					"key1": "value1",
					"key3": "",
					"key5": "",
				},
				Fuzzy: map[string]string{
					"key2": "value2",
					"key4": "",
				},
			},
			false,
		},
		{
			"bad case 1",
			args{"key1 error=value1,key2~value2,key3=,key4~,key5"},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseConditions(tt.args.conditionsStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseConditions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseConditions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParsePaging(t *testing.T) {
	type testData struct {
		req    *restful.Request
		limit  int
		offset int
	}

	table := []testData{{
		req: &restful.Request{Request: &http.Request{URL: &url.URL{
			RawQuery: "",
		}}},
		limit: 10, offset: 0,
	}, {
		req: &restful.Request{Request: &http.Request{URL: &url.URL{
			RawQuery: "paging=limit=11,page=1",
		}}},
		limit: 11, offset: 0,
	}, {
		req: &restful.Request{Request: &http.Request{URL: &url.URL{
			RawQuery: "paging=limit=10,page=2",
		}}},
		limit: 10, offset: 10,
	}, {
		req: &restful.Request{Request: &http.Request{URL: &url.URL{
			RawQuery: "paging=limit=a,page=2",
		}}},
		limit: 10, offset: 0,
	}, {
		req: &restful.Request{Request: &http.Request{URL: &url.URL{
			RawQuery: "paging=limit=10,page=a",
		}}},
		limit: 10, offset: 0,
	}, {
		req: &restful.Request{Request: &http.Request{URL: &url.URL{
			RawQuery: "paging=limit=a,page=a",
		}}},
		limit: 10, offset: 0,
	}, {
		req: &restful.Request{Request: &http.Request{URL: &url.URL{
			RawQuery: "limit=10&page=2",
		}}},
		limit: 10, offset: 10,
	}, {
		req: &restful.Request{Request: &http.Request{URL: &url.URL{
			RawQuery: "page=2",
		}}},
		limit: 10, offset: 10,
	}, {
		req: &restful.Request{Request: &http.Request{URL: &url.URL{
			RawQuery: "page=3",
		}}},
		limit: 10, offset: 20,
	}, {
		req: &restful.Request{Request: &http.Request{URL: &url.URL{
			RawQuery: "start=10&limit10",
		}}},
		limit: 10, offset: 10,
	}, {
		req: &restful.Request{Request: &http.Request{URL: &url.URL{
			RawQuery: "page=3&start=10&limit10",
		}}},
		limit: 10, offset: 10,
	}}

	for index, item := range table {
		limit, offset := ParsePaging(item.req)
		assert.Equal(t, item.limit, limit, "index: [%d], wrong limit", index)
		assert.Equal(t, item.offset, offset, "index: [%d], wrong offset", index)
	}
}

func TestAtoiOrDefault(t *testing.T) {
	type testData struct {
		msg      string
		str      string
		def      int
		expected int
	}
	table := []testData{{
		msg:      "non-numerical",
		str:      "a",
		def:      1,
		expected: 1,
	}, {
		msg:      "non-numerical, empty string",
		str:      "",
		def:      1,
		expected: 1,
	}, {
		msg:      "numerical",
		str:      "2",
		def:      1,
		expected: 2,
	}}

	for index, item := range table {
		result := AtoiOrDefault(item.str, item.def)
		assert.Equal(t, item.expected, result, "test case[%d]: '%s' is wrong", index, item.msg)
	}
}
