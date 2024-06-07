/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package rest

import "github.com/emicklei/go-restful/v3"

type Handler interface {
	AddToContainer(c *restful.Container) error
}
