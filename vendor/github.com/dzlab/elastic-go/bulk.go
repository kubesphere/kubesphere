package elastic

import ()

const (
	// BULK constant name of Elasticsearch bulk operations
	BULK = "bulk"
)

// Bulk a strcuture representing bulk operations
type Bulk struct {
	client *Elasticsearch
	parser Parser
	url    string
	ops    []Dict
}

// Operation a structure representing a bulk operation
type Operation struct {
	_id int
	doc Dict
}

// newBulk creates a new Bulk of operations
func newBulk() *Bulk {
	return &Bulk{url: "", ops: []Dict{}}
}

// NewOperation creates a new operation with the given id
func NewOperation(id int) *Operation {
	return &Operation{_id: id, doc: make(Dict)}
}

// Add adds a field to this document
func (op *Operation) Add(name string, value interface{}) *Operation {
	op.doc[name] = value
	return op
}

// AddMultiple adds a field with multiple values to this document
func (op *Operation) AddMultiple(name string, values ...interface{}) *Operation {
	op.doc[name] = values
	return op
}

// String get a string representation of this operation
func (op *Operation) String() string {
	return String(op.doc)
}

// Bulk creates a new Bulk operations
func (client *Elasticsearch) Bulk(index, docType string) *Bulk {
	url := client.request(index, docType, -1, BULK)
	return &Bulk{
		client: client,
		parser: &BulkResultParser{},
		url:    url,
		ops:    []Dict{},
	}
}

// AddOperation adds an operation to this bulk
func (bulk *Bulk) AddOperation(op *Operation) *Bulk {
	indexOp := Dict{"index": Dict{"_id": op._id}}
	bulk.ops = append(bulk.ops, indexOp)
	bulk.ops = append(bulk.ops, op.doc)
	return bulk
}

// String gets a string representation of the list of operations in this bulk
func (bulk *Bulk) String() string {
	ops := ""
	for _, op := range bulk.ops {
		ops += String(op) + "\n"
	}
	ops = ops[:len(ops)-len("\n")]
	return ops
}

// Post submits a bulk that consists of a list of operations
// POST /:index/:type/_bulk
func (bulk *Bulk) Post() {
	bulk.client.Execute("POST", bulk.url, bulk.String(), bulk.parser)
}
