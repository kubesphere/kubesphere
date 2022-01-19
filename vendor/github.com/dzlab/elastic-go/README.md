elastic-go
==============
[![Build Status](https://travis-ci.org/dzlab/elastic-go.png)](https://travis-ci.org/dzlab/elastic-go)
[![Coverage Status](https://coveralls.io/repos/github/dzlab/elastic-go/badge.svg?branch=master)](https://coveralls.io/github/dzlab/elastic-go?branch=master)
[![Code Climate](https://codeclimate.com/github/dzlab/elastic-go/badges/gpa.svg)](https://codeclimate.com/github/dzlab/elastic-go)
[![GoDoc](https://godoc.org/github.com/dzlab/elastic-go?status.svg)](https://godoc.org/github.com/dzlab/elastic-go)

elastic-go is a golang client library that wraps Elasticsearch REST API. It currently has support for:
* Search
* Index
* Mapping
* Analyze
* ... more to come

### Installation
```go get github.com/dzlab/elastic-go```

#### Tests and coverage
```
go test
gocov test | gocov report
```

### Documentation
http://godoc.org/github.com/dzlab/elastic-go

### Usage
```
import e "github.com/dzlab/elastic-go"
...
client := &e.Elasticsearch{Addr: "localhost:9200"}
client.Search("", "").Add("from", 30).Add("size", 10).Get()
// create an index example
client.Index("my_index").Delete()
cf := e.NewAnalyzer("char_filter")
  .Add1("&_to_and", "type", "mapping")
  .Add2("&_to_and", map[string]interface{}{
      "mappings": []string{"&=> and "},
    }
  )
f := e.NewAnalyzer("filter")
  .Add2("my_stopwords", map[string]interface{}{
      "type": "stop", 
      "stopwords": []string{"the", "a"},
    }
  )
a := e.NewAnalyzer("analyzer")
  .Add2("my_analyzer", e.Dict{
      "type": "custom", 
      "char_filter": []string{"html_strip", "&_to_and"}, 
      "tokenizer": "standard", 
      "filter": []string{"lowercase", "my_stopwords"},
    }
  )
client.Index("my_index").AddAnalyzer(cf).AddAnalyzer(f).AddAnalyzer(a).Put()

// try the analyzer with some data
c.Analyze("my_index").Analyzer("my_analyzer").Get("<p>a paragraph</p>")

// create mapping for a document
client.Mapping("my_index", "my_type")
  .AddField("title", e.Dict{"type":"string", "analyzer": "standard"})
  .AddField("body", e.Dict{"type":"string", "analyzer": "my_analyzer"})
  .Put()

// insert some data
c.Insert("my_index", "my_type").Document(1, e.Dict{"title": "some title", "body": "<p> a paragraph</p>"}).Put()
```

### Contribute
This library is still under very active development. Any contribution is welcome.

Some planned features:

* A REPL to interact easily with Elasticsearch
* ...
