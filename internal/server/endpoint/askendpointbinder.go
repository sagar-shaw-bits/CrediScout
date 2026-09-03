package endpoint

import (
	"github.com/gin-gonic/gin"

	askhandler "github.com/sagar-shaw-bits/CrediScout/internal/handlers/rest/ask"
)

type AskEndpointBinder struct {
	AskEndpointHandler *askhandler.AskEndpointHandler
}

func NewAskEndpointBinder(h *askhandler.AskEndpointHandler) *AskEndpointBinder {
	return &AskEndpointBinder{AskEndpointHandler: h}
}

func (b *AskEndpointBinder) AddEndpoints(routerGroup *gin.RouterGroup) {
	routerGroup.POST("/ask", b.AskEndpointHandler.AskEndpointHandler())
}
