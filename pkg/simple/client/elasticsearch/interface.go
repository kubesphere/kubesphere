package esclient

type Client interface {
	// Perform Search API
	Search(body []byte) ([]byte, error)
	GetTotalHitCount(v interface{}) int64
}
