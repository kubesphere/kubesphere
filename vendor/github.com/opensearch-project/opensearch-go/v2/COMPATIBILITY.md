- [Compatibility with OpenSearch](#compatibility-with-opensearch)
- [Upgrading](#upgrading)

## Compatibility with OpenSearch

The below matrix shows the compatibility of the [`opensearch-go`](https://pkg.go.dev/github.com/opensearch-project/opensearch-go) with versions of [`OpenSearch`](https://opensearch.org/downloads.html#opensearch).

| OpenSearch Version | Client Version |
| --- | --- |
| 1.0.0 | 1.0.0 |
| 1.0.1 | 1.0.0 |
| 1.1.0 | 1.1.0 |
| 1.2.0 | 1.1.0 |
| 1.2.1 | 1.1.0 |
| 1.2.2 | 1.1.0 |
| 1.2.3 | 1.1.0 |
| 1.2.4 | 1.1.0 |
| 1.3.0 | 1.1.0 |
| 1.3.1 | 1.1.0 |
| 2.0.0 | 2.0.0 |

## Upgrading

Major versions of OpenSearch introduce breaking changes that require careful upgrades of the client. While `opensearch-go-client` 2.0.0 works against OpenSearch 1.3.1, certain deprecated features removed in OpenSearch 2.0 have also been removed from the client. Please refer to the [OpenSearch documentation](https://opensearch.org/docs/latest/clients/index/) for more information.