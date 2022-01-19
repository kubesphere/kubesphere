package elastic

import (
	"fmt"
	"io/ioutil"
	"log"
)

const (
	// REFRESH refresh
	REFRESH = "refresh"
	// FLUSH flush
	FLUSH = "flush"
	// OPTIMIZE optimize
	OPTIMIZE = "optimize"
)

// ShardMgmtOp a structure for creating shard management operations
type ShardMgmtOp struct {
	url    string
	params map[string]string
}

func newShardMgmtOp(operation string) *ShardMgmtOp {
	return &ShardMgmtOp{url: operation, params: make(map[string]string)}
}

// Refresh create a refresh API call in order to force recently added document to be visible to search calls
func (client *Elasticsearch) Refresh(index string) *ShardMgmtOp {
	url := client.request(index, "", -1, REFRESH)
	return &ShardMgmtOp{url: url}
}

// Flush creates a flush API call in order to force commit and trauncating the 'translog'
// See, chapter 11. Inside a shard (Elasticsearch Definitive Guide)
func (client *Elasticsearch) Flush(index string) *ShardMgmtOp {
	url := client.request(index, "", -1, FLUSH)
	return &ShardMgmtOp{url: url, params: make(map[string]string)}
}

// Optimize create an Optimize API call in order to force mering shards into a number of segments
func (client *Elasticsearch) Optimize(index string) *ShardMgmtOp {
	url := client.request(index, "", -1, OPTIMIZE)
	return &ShardMgmtOp{url: url, params: make(map[string]string)}
}

// AddParam adds a query parameter to ths Flush API url (e.g. wait_for_ongoing), or Optmize API (e.g. max_num_segment to 1)
func (op *ShardMgmtOp) AddParam(name, value string) *ShardMgmtOp {
	op.params[name] = value
	return op
}

// urlString get a string representation of this API url
func (op *ShardMgmtOp) urlString() string {
	return urlString(op.url, op.params)
}

// Post submit a shard managemnt request
// POST /:index/_refresh
func (op *ShardMgmtOp) Post() {
	url := op.urlString()
	log.Println("POST", url)
	reader, err := exec("POST", op.url, nil)
	if err != nil {
		log.Println(err)
		return
	}
	if data, err := ioutil.ReadAll(reader); err == nil {
		fmt.Println(string(data))
	}
}
