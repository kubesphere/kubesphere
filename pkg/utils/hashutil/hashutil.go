/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package hashutil

import (
	"encoding/hex"
	"hash/fnv"
	"io"

	"code.cloudfoundry.org/bytefmt"
	"k8s.io/klog/v2"

	"kubesphere.io/kubesphere/pkg/utils/readerutils"
)

func GetMD5(reader io.ReadCloser) (string, error) {
	md5reader := readerutils.NewMD5Reader(reader)
	data := make([]byte, bytefmt.KILOBYTE)
	for {
		_, err := md5reader.Read(data)
		if err != nil {
			if err == io.EOF {
				break
			}
			klog.Error(err)
			return "", err
		}
	}
	err := reader.Close()
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(md5reader.MD5()), nil
}

func FNVString(text []byte) string {
	h := fnv.New64a()
	if _, err := h.Write(text); err != nil {
		klog.Error(err)
	}
	return hex.EncodeToString(h.Sum(nil))
}
