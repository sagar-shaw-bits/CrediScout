package endpoint

import "github.com/gin-gonic/gin"

type EndpointBinder interface {
	AddEndpoints(routerGroup *gin.RouterGroup)
}
