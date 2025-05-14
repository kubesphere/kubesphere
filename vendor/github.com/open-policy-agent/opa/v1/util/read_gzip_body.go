package util

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/open-policy-agent/opa/v1/util/decoding"
)

var gzipReaderPool = sync.Pool{
	New: func() interface{} {
		reader := new(gzip.Reader)
		return reader
	},
}

// Note(philipc): Originally taken from server/server.go
// The DecodingLimitHandler handles validating that the gzip payload is within the
// allowed max size limit. Thus, in the event of a forged payload size trailer,
// the worst that can happen is that we waste memory up to the allowed max gzip
// payload size, but not an unbounded amount of memory, as was potentially
// possible before.
func ReadMaybeCompressedBody(r *http.Request) ([]byte, error) {
	var content *bytes.Buffer
	// Note(philipc): If the request body is of unknown length (such as what
	// happens when 'Transfer-Encoding: chunked' is set), we have to do an
	// incremental read of the body. In this case, we can't be too clever, we
	// just do the best we can with whatever is streamed over to us.
	// Fetch gzip payload size limit from request context.
	if maxLength, ok := decoding.GetServerDecodingMaxLen(r.Context()); ok {
		bs, err := io.ReadAll(io.LimitReader(r.Body, maxLength))
		if err != nil {
			return bs, err
		}
		content = bytes.NewBuffer(bs)
	} else {
		// Read content from the request body into a buffer of known size.
		content = bytes.NewBuffer(make([]byte, 0, r.ContentLength))
		if _, err := io.CopyN(content, r.Body, r.ContentLength); err != nil {
			return content.Bytes(), err
		}
	}

	// Decompress gzip content by reading from the buffer.
	if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
		// Fetch gzip payload size limit from request context.
		gzipMaxLength, _ := decoding.GetServerDecodingGzipMaxLen(r.Context())

		// Note(philipc): The last 4 bytes of a well-formed gzip blob will
		// always be a little-endian uint32, representing the decompressed
		// content size, modulo 2^32. We validate that the size is safe,
		// earlier in DecodingLimitHandler.
		sizeTrailerField := binary.LittleEndian.Uint32(content.Bytes()[content.Len()-4:])
		if sizeTrailerField > uint32(gzipMaxLength) {
			return content.Bytes(), errors.New("gzip payload too large")
		}
		// Pull a gzip decompressor from the pool, and assign it to the current
		// buffer, using Reset(). Later, return it back to the pool for another
		// request to use.
		gzReader := gzipReaderPool.Get().(*gzip.Reader)
		if err := gzReader.Reset(content); err != nil {
			return nil, err
		}
		defer gzReader.Close()
		defer gzipReaderPool.Put(gzReader)
		decompressedContent := bytes.NewBuffer(make([]byte, 0, sizeTrailerField))
		if _, err := io.CopyN(decompressedContent, gzReader, int64(sizeTrailerField)); err != nil {
			return decompressedContent.Bytes(), err
		}
		return decompressedContent.Bytes(), nil
	}

	// Request was not compressed; return the content bytes.
	return content.Bytes(), nil
}
