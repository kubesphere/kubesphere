package rest

import "github.com/emicklei/go-restful/v3"

type Handler interface {
	AddToContainer(c *restful.Container) error
}
