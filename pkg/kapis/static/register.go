/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package static

import (
	"github.com/emicklei/go-restful/v3"
)

func (h *handler) AddToContainer(c *restful.Container) error {
	webservice := new(restful.WebService)
	webservice.Path("/static")
	webservice.Route(webservice.POST("/images").
		Doc("Upload image").
		Consumes("multipart/form-data").
		Produces(restful.MIME_JSON).
		To(h.uploadImage).
		Param(webservice.FormParameter("image", "Image content, support JPG, PNG, SVG; size limit is 2MB.")))
	webservice.Route(webservice.GET("/images/{file}").
		Doc("Get image file").
		To(h.getImage).
		Param(webservice.PathParameter("file", "File name of the image.")))
	c.Add(webservice)
	return nil
}
