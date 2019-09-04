/*
Package estransport provides the transport layer for the Elasticsearch client.

It is automatically included in the client provided by the github.com/elastic/go-elasticsearch package
and is not intended for direct use: to configure the client, use the elasticsearch.Config struct.

The default HTTP transport of the client is http.Transport.

The package defines the "Selector" interface for getting a URL from the list. At the moment,
the implementation is rather minimal: the client takes a slice of url.URL pointers,
and round-robins across them when performing the request.

The package defines the "Logger" interface for logging information about request and response.
It comes with several bundled loggers for logging in text and JSON.

*/
package estransport
