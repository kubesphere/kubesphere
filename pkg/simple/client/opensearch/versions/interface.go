package versions

// versioned es client interface
type Client interface {
	Search(indices string, body []byte, scroll bool) ([]byte, error)
	Scroll(id string) ([]byte, error)
	ClearScroll(id string)
	GetTotalHitCount(v interface{}) int64
}
