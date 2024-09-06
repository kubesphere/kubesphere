/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package openapi

import (
	"fmt"
	"net/http"
	"time"
)

const (
	specDownloadTimeout = time.Minute
)

type CacheableDownloader interface {
	Name() string
	GetV2() ([]byte, error)
	GetV3() ([]byte, error)
	UpdateDownloader(apiService ApiService) error
}

type Downloader struct{}

func NewDownloader() *Downloader {
	return &Downloader{}
}

func (s *Downloader) Download(handler http.Handler, req *http.Request) (data []byte, err error) {
	handler = http.TimeoutHandler(handler, specDownloadTimeout, "request timed out")

	writer := newInMemoryResponseWriter()
	handler.ServeHTTP(writer, req)

	switch writer.respCode {
	case http.StatusNotModified:
		return nil, nil
	case http.StatusNotFound:
		return nil, ErrAPIServiceNotFound
	case http.StatusOK:
		return writer.data, nil
	default:
		return nil, fmt.Errorf("failed to retrieve openAPI spec, http error: %s", writer.String())
	}
}

type inMemoryResponseWriter struct {
	writeHeaderCalled bool
	header            http.Header
	respCode          int
	data              []byte
}

func newInMemoryResponseWriter() *inMemoryResponseWriter {
	return &inMemoryResponseWriter{header: http.Header{}}
}

func (r *inMemoryResponseWriter) Header() http.Header {
	return r.header
}

func (r *inMemoryResponseWriter) WriteHeader(code int) {
	r.writeHeaderCalled = true
	r.respCode = code
}

func (r *inMemoryResponseWriter) Write(in []byte) (int, error) {
	if !r.writeHeaderCalled {
		r.WriteHeader(http.StatusOK)
	}
	r.data = append(r.data, in...)
	return len(in), nil
}

func (r *inMemoryResponseWriter) String() string {
	s := fmt.Sprintf("ResponseCode: %d", r.respCode)
	if r.data != nil {
		s += fmt.Sprintf(", Body: %s", string(r.data))
	}
	if r.header != nil {
		s += fmt.Sprintf(", Header: %s", r.header)
	}
	return s
}

type cacheableDownloader struct {
	apiService ApiService
	downloader *Downloader
}

func NewCacheableDownloader(apiService ApiService, downloader *Downloader) (CacheableDownloader, error) {
	c := &cacheableDownloader{
		apiService: apiService,
		downloader: downloader,
	}
	if err := c.apiService.UpdateAPIService(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *cacheableDownloader) Name() string {
	return c.apiService.Name()
}

func (c *cacheableDownloader) Get(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")

	return c.downloader.Download(c.apiService, req)
}

func (c *cacheableDownloader) GetV2() ([]byte, error) {
	return c.Get("/openapi/v2")
}

func (c *cacheableDownloader) GetV3() ([]byte, error) {
	return c.Get("/openapi/v3")
}

func (c *cacheableDownloader) UpdateDownloader(apiService ApiService) error {
	c.apiService = apiService
	return c.apiService.UpdateAPIService()
}
