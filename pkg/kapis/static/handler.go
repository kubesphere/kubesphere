/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package static

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"time"

	"github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"

	ksapi "kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/rest"
	"kubesphere.io/kubesphere/pkg/simple/client/cache"
)

const (
	imageCacheKeyPrefix = "static:images:"
	mimePNG             = "image/png"
	mimeJPG             = "image/jpeg"
	mimeSVG             = "image/svg+xml"
	// https://mimesniff.spec.whatwg.org/#matching-an-image-type-pattern
	mimeICON      = "image/x-icon"
	maxUploadSize = 2 * 1024 * 1024 // 2M
)

var supportedMIMETypes = sets.NewString(mimePNG, mimeJPG, mimeSVG, mimeICON)

type handler struct {
	cache cache.Interface
}

func init() {
	_ = mime.AddExtensionType(".ico", mimeICON)
}

func NewHandler(cache cache.Interface) rest.Handler {
	return &handler{cache: cache}
}
func NewFakeHandler() rest.Handler {
	return &handler{}
}

func (h *handler) uploadImage(req *restful.Request, resp *restful.Response) {
	req.Request.Body = http.MaxBytesReader(resp, req.Request.Body, maxUploadSize)
	if err := req.Request.ParseMultipartForm(maxUploadSize); err != nil {
		ksapi.HandleBadRequest(resp, req, fmt.Errorf("failed to parse multipart form: %s", err))
		return
	}

	imagePart, existed := req.Request.MultipartForm.File["image"]
	if !existed || len(imagePart) == 0 {
		ksapi.HandleBadRequest(resp, req, fmt.Errorf("invalid form data"))
		return
	}

	header := imagePart[0]
	ext := filepath.Ext(header.Filename)
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = header.Header.Get("Content-Type")
	}

	if !supportedMIMETypes.Has(contentType) {
		ksapi.HandleBadRequest(resp, req, fmt.Errorf("the provided file format is not allowed"))
		return
	}
	file, err := header.Open()
	if err != nil {
		ksapi.HandleInternalError(resp, req, err)
		return
	}
	defer file.Close()
	rawContent, err := io.ReadAll(file)
	if err != nil {
		ksapi.HandleInternalError(resp, req, err)
		return
	}

	hasher := md5.New()
	hasher.Write(rawContent)
	hash := hasher.Sum(nil)

	md5Hex := hex.EncodeToString(hash)
	fileName := fmt.Sprintf("%s%s", md5Hex, ext)
	key := fmt.Sprintf("%s%s", imageCacheKeyPrefix, fileName)

	base64EncodedContent := base64.StdEncoding.EncodeToString(rawContent)
	if err = h.cache.Set(key, base64EncodedContent, cache.NeverExpire); err != nil {
		ksapi.HandleInternalError(resp, req, err)
		return
	}

	result := map[string]string{"image": fileName}
	_ = resp.WriteEntity(result)
}

func (h *handler) getImage(req *restful.Request, resp *restful.Response) {
	fileName := req.PathParameter("file")

	base64EncodedData, err := h.cache.Get(fmt.Sprintf("%s%s", imageCacheKeyPrefix, fileName))
	if err != nil {
		if errors.Is(err, cache.ErrNoSuchKey) {
			ksapi.HandleNotFound(resp, req, fmt.Errorf("image %s not found", fileName))
			return
		}
		ksapi.HandleInternalError(resp, req, err)
		return
	}

	rawContent, err := base64.StdEncoding.DecodeString(base64EncodedData)
	if err != nil {
		klog.Warningf("failed to decode base64 encoded image data: %s, %v", fileName, err)
		ksapi.HandleInternalError(resp, req, err)
		return
	}

	contentType := mime.TypeByExtension(filepath.Ext(fileName))
	resp.Header().Set("Content-Type", contentType)
	resp.Header().Set("Cache-Control", "public, max-age=86400")
	resp.Header().Set("Expires", time.Now().Add(86400*time.Second).Format(time.RFC1123))

	_, err = resp.Write(rawContent)
	if err != nil {
		klog.Warningf("failed to write image data: %s, %v", fileName, err)
	}
}
