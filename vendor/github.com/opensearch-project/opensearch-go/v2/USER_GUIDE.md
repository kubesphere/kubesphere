- [User Guide](#user-guide)
  - [Example](#example)
  - [How to use IAMs as authentication method](#how-to-use-iams-as-authentication-method)
# User Guide

## Example
In the example below, we create a client, an index with non-default settings, insert a document to the index,
search for the document, delete the document and finally delete the index.

```go
package main

import (
    "os"
    "context"
    "crypto/tls"
    "fmt"
    opensearch "github.com/opensearch-project/opensearch-go/v2"
    opensearchapi "github.com/opensearch-project/opensearch-go/v2/opensearchapi"
    "net/http"
    "strings"
)

const IndexName = "go-test-index1"

func main() {

    // Initialize the client with SSL/TLS enabled.
    client, err := opensearch.NewClient(opensearch.Config{
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // For testing only. Use certificate for validation.
        },
        Addresses: []string{"https://localhost:9200"},
        Username:  "admin", // For testing only. Don't store credentials in code.
        Password:  "admin",
    })
    if err != nil {
        fmt.Println("cannot initialize", err)
        os.Exit(1)
    }
    
    // Print OpenSearch version information on console.
    fmt.Println(client.Info())
    
    // Define index mapping.
    mapping := strings.NewReader(`{
     'settings': {
       'index': {
            'number_of_shards': 4
            }
          }
     }`)
    
    // Create an index with non-default settings.
    res := opensearchapi.CreateRequest{
        Index: IndexName,
        Body:  mapping,
    }
    fmt.Println("creating index", res)
    
    // Add a document to the index.
    document := strings.NewReader(`{
        "title": "Moneyball",
        "director": "Bennett Miller",
        "year": "2011"
    }`)
    
    docId := "1"
    req := opensearchapi.IndexRequest{
        Index:      IndexName,
        DocumentID: docId,
        Body:       document,
    }
    insertResponse, err := req.Do(context.Background(), client)
    if err != nil {
        fmt.Println("failed to insert document ", err)
        os.Exit(1)
    }
    fmt.Println(insertResponse)
    
    // Search for the document.
    content := strings.NewReader(`{
       "size": 5,
       "query": {
           "multi_match": {
           "query": "miller",
           "fields": ["title^2", "director"]
           }
      }
    }`)
    
    search := opensearchapi.SearchRequest{
        Body: content,
    }
    
    searchResponse, err := search.Do(context.Background(), client)
    if err != nil {
        fmt.Println("failed to search document ", err)
        os.Exit(1)
    }
    fmt.Println(searchResponse)
    
    // Delete the document.
    delete := opensearchapi.DeleteRequest{
        Index:      IndexName,
        DocumentID: docId,
    }
    
    deleteResponse, err := delete.Do(context.Background(), client)
    if err != nil {
        fmt.Println("failed to delete document ", err)
        os.Exit(1)
    }
    fmt.Println("deleting document")
    fmt.Println(deleteResponse)
    
    // Delete previously created index.
    deleteIndex := opensearchapi.IndicesDeleteRequest{
        Index: []string{IndexName},
    }
    
    deleteIndexResponse, err := deleteIndex.Do(context.Background(), client)
    if err != nil {
        fmt.Println("failed to delete index ", err)
        os.Exit(1)
    }
    fmt.Println("deleting index", deleteIndexResponse)
}

```

## How to use IAMs as authentication method

Before starting, we strongly recommend reading the full AWS documentation regarding using IAM credentials to sign requests to OpenSearch APIs.
See [Identity and Access Management in Amazon OpenSearch Service.](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/ac.html)

> Even if you configure a completely open resource-based access policy, all requests to the OpenSearch Service configuration API must be signed. If your policies specify IAM users or roles, requests to the OpenSearch APIs also must be signed using AWS Signature Version 4.
See [Managed Domains signing-service requests.](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/ac.html#managedomains-signing-service-requests)

Here is some sample Go code that shows how to sign each OpenSearch request and automatically search for AWS credentials from the ~/.aws folder or environment variables:

```go
package main

import (
	"context"
	"io"
	"log"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	requestsigner "github.com/opensearch-project/opensearch-go/v2/signer/aws"
)

const endpoint = "" // e.g. https://opensearch-domain.region.com

func main() {
	ctx := context.Background()

	// Create an AWS request Signer and load AWS configuration using default config folder or env vars.
	// See https://docs.aws.amazon.com/opensearch-service/latest/developerguide/request-signing.html#request-signing-go
	signer, err := requestsigner.NewSigner(session.Options{SharedConfigState: session.SharedConfigEnable})
	if err != nil {
		log.Fatal(err) // Do not log.fatal in a production ready app.
	}

	// Create an opensearch client and use the request-signer
	client, err := opensearch.NewClient(opensearch.Config{
		Addresses: []string{endpoint},
		Signer: signer,
	})
	if err != nil {
		log.Fatal("client creation err", err)
	}

	ping := opensearchapi.PingRequest{}

	resp, err := ping.Do(ctx, client)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.IsError() {
		log.Println("ping response status ", resp.Status())

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal("response body read err", err)
		}

		log.Fatal("ping resp body", respBody)
	}

	log.Println("PING OK")
}
```
