package elastic

import (
	"fmt"
)

// Insert a request representing a document insert query
type Insert struct {
	client *Elasticsearch
	parser Parser
	url    string
	id     int64
	doc    interface{}
}

// Insert Create an Insert request, that will submit a new document to elastic search
func (client *Elasticsearch) Insert(index, doctype string) *Insert {
	url := fmt.Sprintf("http://%s/%s/%s", client.Addr, index, doctype)
	return &Insert{
		client: client,
		parser: &InsertResultParser{},
		url:    url,
	}
}

// newInsert Create a new Insert query (for test purpose)
func newInsert() *Insert {
	return &Insert{}
}

// Document set the document to insert
func (insert *Insert) Document(id int64, doc interface{}) *Insert {
	insert.id = id
	insert.doc = doc
	return insert
}

// String returns a string representation of the document
func (insert *Insert) String() string {
	return String(insert.doc)
}

// Put submits a request mappings between the json fields and how Elasticsearch store them
// PUT /:index/:type/:id
func (insert *Insert) Put() {
	// construct the url
	url := fmt.Sprintf("%s/%d", insert.url, insert.id)
	// construct the body
	query := insert.String()
	insert.client.Execute("PUT", url, query, insert.parser)
}
