package s3

import (
	"io"
)

type Interface interface {
	// Upload uploads a object to storage and returns object location if succeeded
	Upload(key, fileName string, body io.Reader) error

	GetDownloadURL(key string, fileName string) (string, error)

	// Delete deletes an object by its key
	Delete(key string) error
}
