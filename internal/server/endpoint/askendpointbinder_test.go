package endpoint

import (
	"testing"

	"github.com/gin-gonic/gin"

	askhandler "github.com/sagar-shaw-bits/CrediScout/internal/handlers/rest/ask"
)

func setupAskRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.Default()
	api := engine.Group("/api")
	binder := NewAskEndpointBinder(askhandler.NewAskEndpointHandler(nil, nil))
	binder.AddEndpoints(api)
	return engine
}

func TestAskEndpointBinder_RegistersRoutes(t *testing.T) {
	expectedRoutes := []Route{
		{Method: "POST", Path: "/api/ask"},
	}

	router := setupAskRouter()
	routesInfo := router.Routes()
	VerifyRoutes(t, expectedRoutes, routesInfo)
}
