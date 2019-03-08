# go-restful-openapi

[![Build Status](https://travis-ci.org/emicklei/go-restful-openapi.png)](https://travis-ci.org/emicklei/go-restful-openapi)
[![GoDoc](https://godoc.org/github.com/emicklei/go-restful-openapi?status.svg)](https://godoc.org/github.com/emicklei/go-restful-openapi)

[openapi](https://www.openapis.org) extension to the go-restful package, targeting [version 2.0](https://github.com/OAI/OpenAPI-Specification)

## The following Go field tags are translated to OpenAPI equivalents
- description
- minimum
- maximum
- optional ( if set to "true" then it is not listed in `required`)
- unique
- modelDescription
- type (overrides the Go type String())
- enum
- readOnly

See TestThatExtraTagsAreReadIntoModel for examples.

## dependencies

- [go-restful](https://github.com/emicklei/go-restful)
- [go-openapi](https://github.com/go-openapi/spec)

© 2018, ernestmicklei.com.  MIT License. Contributions welcome.
