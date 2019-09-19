package esclient

import "time"

type Client interface {
	// Perform Search API
	Search(body []byte, scrollTimeout time.Duration) ([]byte, error)
	Scroll(scrollId string, scrollTimeout time.Duration) ([]byte, error)
	ClearScroll(scrollId string)
	GetTotalHitCount(v interface{}) int64
}
