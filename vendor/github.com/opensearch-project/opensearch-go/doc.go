// SPDX-License-Identifier: Apache-2.0
//
// The OpenSearch Contributors require contributions made to
// this file be licensed under the Apache-2.0 license or a
// compatible open source license.
//
// Modifications Copyright OpenSearch Contributors. See
// GitHub history for details.

// Licensed to opensearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

/*
Package opensearch provides a Go client for OpenSearch.

Create the client with the NewDefaultClient function:

		opensearch.NewDefaultClient()

The OPENSEARCH_URL/ELASTICSEARCH_URL environment variable is used instead of the default URL, when set.
Use a comma to separate multiple URLs.
It is an error to set both environment variable.

To configure the client, pass a Config object to the NewClient function:

		cfg := opensearch.Config{
		  Addresses: []string{
		    "http://localhost:9200",
		    "http://localhost:9201",
		  },
		  Username: "foo",
		  Password: "bar",
		  Transport: &http.Transport{
		    MaxIdleConnsPerHost:   10,
		    ResponseHeaderTimeout: time.Second,
		    DialContext:           (&net.Dialer{Timeout: time.Second}).DialContext,
		    TLSClientConfig: &tls.Config{
		      MinVersion:         tls.VersionTLS11,
		    },
		  },
		}

		opensearch.NewClient(cfg)

See the opensearch_integration_test.go file for more information.

Call the OpenSearch APIs by invoking the corresponding methods on the client:

		res, err := client.Info()
		if err != nil {
		  log.Fatalf("Error getting response: %s", err)
		}

		log.Println(res)

See the github.com/opensearch-project/opensearch-go/opensearchapi package for more information about using the API.

See the github.com/opensearch-project/opensearch-go/opensearchtransport package for more information about configuring the transport.
*/
package opensearch
