/*

 Copyright 2019 The KubeSphere Authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.

*/

package registries

import (
	"github.com/emicklei/go-restful"
	"net/http"

	"kubesphere.io/kubesphere/pkg/errors"
	"kubesphere.io/kubesphere/pkg/models/registries"

	log "github.com/golang/glog"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
)

func RegistryVerify(request *restful.Request, response *restful.Response) {

	authInfo := registries.AuthInfo{}

	err := request.ReadEntity(&authInfo)

	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	err = registries.RegistryVerify(authInfo)

	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	response.WriteAsJson(errors.None)
}

func RegistryImageBlob(request *restful.Request, response *restful.Response) {
	var statusUnauthorized = "Not found or unauthorized"

	imageName := request.QueryParameter("image")
	namespace := request.QueryParameter("namespace")
	secretName := request.QueryParameter("secret")

	// get entry
	entry, err := registries.GetEntryBySecret(namespace, secretName)
	if err != nil && k8serror.IsNotFound(err) {
		log.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusBadRequest, err.Error()), response)
		return
	}
	if err != nil {
		log.Errorf("%+v", err)
		response.WriteAsJson(&registries.ImageDetails{Status: registries.StatusFailed, Message: err.Error()})
		return
	}

	// parse image
	image, err := registries.ParseImage(imageName)
	if err != nil {
		log.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusBadRequest, err.Error()), response)
		return
	}

	// Create the registry client.
	r, err := registries.CreateRegistryClient(entry.Username, entry.Password, image.Domain)
	if err != nil {
		log.Errorf("%+v", err)
		response.WriteAsJson(&registries.ImageDetails{Status: registries.StatusFailed, Message: err.Error()})
		return
	}

	digestUrl := r.GetDigestUrl(image)

	// Get token.
	token, err := r.Token(digestUrl)
	if err != nil {
		log.Errorf("%+v", err)
		response.WriteAsJson(&registries.ImageDetails{Status: registries.StatusFailed, Message: err.Error()})
		return
	}

	// Get digest.
	imageManifest, err, statusCode := r.ImageManifest(image, token)
	if statusCode == http.StatusUnauthorized {
		log.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(statusCode, statusUnauthorized), response)
		return
	}
	if err != nil {
		log.Errorf("%+v", err)
		response.WriteAsJson(&registries.ImageDetails{Status: registries.StatusFailed, Message: err.Error()})
	}
	image.Digest = imageManifest.ManifestConfig.Digest

	// Get blob.
	imageBlob, err := r.ImageBlob(image, token)
	if err != nil {
		log.Errorf("%+v", err)
		response.WriteAsJson(&registries.ImageDetails{Status: registries.StatusFailed, Message: err.Error()})
		return
	}

	imageDetails := &registries.ImageDetails{
		Status:        registries.StatusSuccess,
		ImageManifest: imageManifest,
		ImageBlob:     imageBlob,
		ImageTag:      image.Tag,
		Registry:      image.Domain,
	}

	response.WriteAsJson(imageDetails)
}
