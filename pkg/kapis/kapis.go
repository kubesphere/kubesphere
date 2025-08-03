/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package kapis

import (
	"github.com/gin-gonic/gin"
)

// RegisterKubeSphereApis registers all KubeSphere APIs with the gin engine
func RegisterKubeSphereApis(g *gin.Engine) {
	// Basic health check endpoint
	g.GET("/healthz", func(c *gin.Context) {
		c.String(200, "ok")
	})

	// Other API registrations would go here
}
