package endpoint

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type Route struct {
	Method string
	Path   string
}

func VerifyRoutes(t *testing.T, expectedRoutes []Route, routesInfo gin.RoutesInfo) {
	expectedRoutesCount := len(expectedRoutes)
	actualRouteInfoCount := len(routesInfo)

	assert.Equal(t, expectedRoutesCount, actualRouteInfoCount)

	for i := 0; i < expectedRoutesCount; i++ {
		routeFound := false
		expectedRoute := expectedRoutes[i]
		for j := 0; j < actualRouteInfoCount; j++ {
			routeMethodMatch := expectedRoutes[i].Method == routesInfo[j].Method
			routePathMatch := expectedRoutes[i].Path == routesInfo[j].Path
			routeFound = routeMethodMatch && routePathMatch
			if routeFound {
				break
			}
		}
		assert.True(t, routeFound, "Route '%s %s' not found.", expectedRoute.Method, expectedRoute.Path)
	}
}
