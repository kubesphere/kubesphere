/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package readerutils

import (
	"crypto/md5"
	"hash"
	"io"
)

type MD5Reader struct {
	md5  hash.Hash
	body io.Reader
}

func (reader *MD5Reader) Read(b []byte) (int, error) {
	n, err := reader.body.Read(b)
	if err != nil {
		return n, err
	}
	return reader.md5.Write(b[:n])
}

func (reader *MD5Reader) MD5() []byte {
	return reader.md5.Sum(nil)
}

func NewMD5Reader(reader io.Reader) *MD5Reader {
	return &MD5Reader{
		md5:  md5.New(),
		body: reader,
	}
}
