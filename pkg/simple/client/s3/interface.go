package s3

import (
	"io"
	"time"
)

type Interface interface {
	// Upload uploads a object to storage and returns object location if succeeded
	Upload(key string, body io.Reader) (string, error)

	// Get retrieves and object's downloadable location if succeeded
	Get(key string, fileName string, expire time.Duration) (string, error)

	// Delete deletes an object by its key
	Delete(key string) error
}
