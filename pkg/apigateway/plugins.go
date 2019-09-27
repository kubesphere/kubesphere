package apigateway

import (
	"github.com/mholt/caddy"

	"kubesphere.io/kubesphere/pkg/apigateway/caddy-plugin/authenticate"
	"kubesphere.io/kubesphere/pkg/apigateway/caddy-plugin/authentication"
	swagger "kubesphere.io/kubesphere/pkg/apigateway/caddy-plugin/swagger"
)

func RegisterPlugins() {
	caddy.RegisterPlugin("swagger", caddy.Plugin{
		ServerType: "http",
		Action:     swagger.Setup,
	})

	caddy.RegisterPlugin("authenticate", caddy.Plugin{
		ServerType: "http",
		Action:     authenticate.Setup,
	})

	caddy.RegisterPlugin("authentication", caddy.Plugin{
		ServerType: "http",
		Action:     authentication.Setup,
	})
}
