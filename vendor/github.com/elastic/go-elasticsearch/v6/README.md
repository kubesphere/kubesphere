# go-elasticsearch

The official Go client for [Elasticsearch](https://www.elastic.co/products/elasticsearch).

[![GoDoc](https://godoc.org/github.com/elastic/go-elasticsearch?status.svg)](http://godoc.org/github.com/elastic/go-elasticsearch)
[![Travis-CI](https://travis-ci.org/elastic/go-elasticsearch.svg?branch=master)](https://travis-ci.org/elastic/go-elasticsearch)
[![Go Report Card](https://goreportcard.com/badge/github.com/elastic/go-elasticsearch)](https://goreportcard.com/report/github.com/elastic/go-elasticsearch)
[![codecov.io](https://codecov.io/github/elastic/go-elasticsearch/coverage.svg?branch=master)](https://codecov.io/gh/elastic/go-elasticsearch?branch=master)

## Compatibility

The client major versions correspond to the compatible Elasticsearch major versions: to connect to Elasticsearch `7.x`, use a [`7.x`](https://github.com/elastic/go-elasticsearch/tree/7.x) version of the client, to connect to Elasticsearch `6.x`, use a [`6.x`](https://github.com/elastic/go-elasticsearch/tree/6.x) version of the client.

When using Go modules, include the version in the import path, and specify either an explicit version or a branch:

    require github.com/elastic/go-elasticsearch/v7 7.x
    require github.com/elastic/go-elasticsearch/v7 7.0.0

It's possible to use multiple versions of the client in a single project:

    // go.mod
    github.com/elastic/go-elasticsearch/v6 6.x
    github.com/elastic/go-elasticsearch/v7 7.x

    // main.go
    import (
      elasticsearch6 "github.com/elastic/go-elasticsearch/v6"
      elasticsearch7 "github.com/elastic/go-elasticsearch/v7"
    )
    // ...
    es6, _ := elasticsearch6.NewDefaultClient()
    es7, _ := elasticsearch7.NewDefaultClient()

The `master` branch of the client is compatible with the current `master` branch of Elasticsearch.

<!-- ----------------------------------------------------------------------------------------------- -->

## Installation

Add the package to your `go.mod` file:

    require github.com/elastic/go-elasticsearch/v6 6.x

Or, clone the repository:

    git clone --branch 6.x https://github.com/elastic/go-elasticsearch.git $GOPATH/src/github.com/elastic/go-elasticsearch

A complete example:

```bash
mkdir my-elasticsearch-app && cd my-elasticsearch-app

cat > go.mod <<-END
  module my-elasticsearch-app

  require github.com/elastic/go-elasticsearch/v6 6.x
END

cat > main.go <<-END
  package main

  import (
    "log"

    "github.com/elastic/go-elasticsearch/v6"
  )

  func main() {
    es, _ := elasticsearch.NewDefaultClient()
    log.Println(elasticsearch.Version)
    log.Println(es.Info())
  }
END

go run main.go
```


<!-- ----------------------------------------------------------------------------------------------- -->

## Usage

The `elasticsearch` package ties together two separate packages for calling the Elasticsearch APIs and transferring data over HTTP: `esapi` and `estransport`, respectively.

Use the `elasticsearch.NewDefaultClient()` function to create the client with the default settings.

```golang
es, err := elasticsearch.NewDefaultClient()
if err != nil {
  log.Fatalf("Error creating the client: %s", err)
}

res, err := es.Info()
if err != nil {
  log.Fatalf("Error getting response: %s", err)
}

log.Println(res)

// [200 OK] {
//   "name" : "node-1",
//   "cluster_name" : "go-elasticsearch"
// ...
```

When you export the `ELASTICSEARCH_URL` environment variable,
it will be used to set the cluster endpoint(s). Separate multiple adresses by a comma.

To set the cluster endpoint(s) programatically, pass them in the configuration object
to the `elasticsearch.NewClient()` function.

```golang
cfg := elasticsearch.Config{
  Addresses: []string{
    "http://localhost:9200",
    "http://localhost:9201",
  },
}
es, err := elasticsearch.NewClient(cfg)
// ...
```

To configure the HTTP settings, pass a [`http.Transport`](https://golang.org/pkg/net/http/#Transport)
object in the configuration object (the values are for illustrative purposes only).

```golang
cfg := elasticsearch.Config{
  Transport: &http.Transport{
    MaxIdleConnsPerHost:   10,
    ResponseHeaderTimeout: time.Second,
    DialContext:           (&net.Dialer{Timeout: time.Second}).DialContext,
    TLSClientConfig: &tls.Config{
      MinVersion: tls.VersionTLS11,
      // ...
    },
  },
}

es, err := elasticsearch.NewClient(cfg)
// ...
```

See the [`_examples/configuration.go`](_examples/configuration.go) and
[`_examples/customization.go`](_examples/customization.go) files for
more examples of configuration and customization of the client.

The following example demonstrates a more complex usage. It fetches the Elasticsearch version from the cluster, indexes a couple of documents concurrently, and prints the search results, using a lightweight wrapper around the response body.

```golang
// $ go run _examples/main.go

package main

import (
  "bytes"
  "context"
  "encoding/json"
  "log"
  "strconv"
  "strings"
  "sync"

  "github.com/elastic/go-elasticsearch/v6"
  "github.com/elastic/go-elasticsearch/v6/esapi"
)

func main() {
  log.SetFlags(0)

  var (
    r  map[string]interface{}
    wg sync.WaitGroup
  )

  // Initialize a client with the default settings.
  //
  // An `ELASTICSEARCH_URL` environment variable will be used when exported.
  //
  es, err := elasticsearch.NewDefaultClient()
  if err != nil {
    log.Fatalf("Error creating the client: %s", err)
  }

  // 1. Get cluster info
  //
  res, err := es.Info()
  if err != nil {
    log.Fatalf("Error getting response: %s", err)
  }
  // Check response status
  if res.IsError() {
    log.Fatalf("Error: %s", res.String())
  }
  // Deserialize the response into a map.
  if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
    log.Fatalf("Error parsing the response body: %s", err)
  }
  // Print client and server version numbers.
  log.Printf("Client: %s", elasticsearch.Version)
  log.Printf("Server: %s", r["version"].(map[string]interface{})["number"])
  log.Println(strings.Repeat("~", 37))

  // 2. Index documents concurrently
  //
  for i, title := range []string{"Test One", "Test Two"} {
    wg.Add(1)

    go func(i int, title string) {
      defer wg.Done()

      // Build the request body.
      var b strings.Builder
      b.WriteString(`{"title" : "`)
      b.WriteString(title)
      b.WriteString(`"}`)

      // Set up the request object.
      req := esapi.IndexRequest{
        Index:      "test",
        DocumentID: strconv.Itoa(i + 1),
        Body:       strings.NewReader(b.String()),
        Refresh:    "true",
      }

      // Perform the request with the client.
      res, err := req.Do(context.Background(), es)
      if err != nil {
        log.Fatalf("Error getting response: %s", err)
      }
      defer res.Body.Close()

      if res.IsError() {
        log.Printf("[%s] Error indexing document ID=%d", res.Status(), i+1)
      } else {
        // Deserialize the response into a map.
        var r map[string]interface{}
        if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
          log.Printf("Error parsing the response body: %s", err)
        } else {
          // Print the response status and indexed document version.
          log.Printf("[%s] %s; version=%d", res.Status(), r["result"], int(r["_version"].(float64)))
        }
      }
    }(i, title)
  }
  wg.Wait()

  log.Println(strings.Repeat("-", 37))

  // 3. Search for the indexed documents
  //
  // Build the request body.
  var buf bytes.Buffer
  query := map[string]interface{}{
    "query": map[string]interface{}{
      "match": map[string]interface{}{
        "title": "test",
      },
    },
  }
  if err := json.NewEncoder(&buf).Encode(query); err != nil {
    log.Fatalf("Error encoding query: %s", err)
  }

  // Perform the search request.
  res, err = es.Search(
    es.Search.WithContext(context.Background()),
    es.Search.WithIndex("test"),
    es.Search.WithBody(&buf),
    es.Search.WithTrackTotalHits(true),
    es.Search.WithPretty(),
  )
  if err != nil {
    log.Fatalf("Error getting response: %s", err)
  }
  defer res.Body.Close()

  if res.IsError() {
    var e map[string]interface{}
    if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
      log.Fatalf("Error parsing the response body: %s", err)
    } else {
      // Print the response status and error information.
      log.Fatalf("[%s] %s: %s",
        res.Status(),
        e["error"].(map[string]interface{})["type"],
        e["error"].(map[string]interface{})["reason"],
      )
    }
  }

  if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
    log.Fatalf("Error parsing the response body: %s", err)
  }
  // Print the response status, number of results, and request duration.
  log.Printf(
    "[%s] %d hits; took: %dms",
    res.Status(),
    int(r["hits"].(map[string]interface{})["total"].(float64)),
    int(r["took"].(float64)),
  )
  // Print the ID and document source for each hit.
  for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
    log.Printf(" * ID=%s, %s", hit.(map[string]interface{})["_id"], hit.(map[string]interface{})["_source"])
  }

  log.Println(strings.Repeat("=", 37))
}

// Client: 6.7.0-SNAPSHOT
// Server: 6.7.2
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// [201 Created] updated; version=1
// [201 Created] updated; version=1
// -------------------------------------
// [200 OK] 2 hits; took: 5ms
//  * ID=1, map[title:Test One]
//  * ID=2, map[title:Test Two]
// =====================================
```

As you see in the example above, the `esapi` package allows to call the Elasticsearch APIs in two distinct ways: either by creating a struct, such as `IndexRequest`, and calling its `Do()` method by passing it a context and the client, or by calling the `Search()` function on the client directly, using the option functions such as `WithIndex()`. See more information and examples in the
[package documentation](https://godoc.org/github.com/elastic/go-elasticsearch/esapi).

The `estransport` package handles the transfer of data to and from Elasticsearch. At the moment, the implementation is really minimal: it only round-robins across the configured cluster endpoints. In future, more features — retrying failed requests, ignoring certain status codes, auto-discovering nodes in the cluster, and so on — will be added.

<!-- ----------------------------------------------------------------------------------------------- -->

## Helpers

The `esutil` package provides convenience helpers for working with the client. At the moment, it provides the
`esutil.JSONReader()` helper function.

<!-- ----------------------------------------------------------------------------------------------- -->

## Examples

The **[`_examples`](./_examples)** folder contains a number of recipes and comprehensive examples to get you started with the client, including configuration and customization of the client, mocking the transport for unit tests, embedding the client in a custom type, building queries, performing requests, and parsing the responses.

<!-- ----------------------------------------------------------------------------------------------- -->

## License

(c) 2019 Elasticsearch. Licensed under the Apache License, Version 2.0.
