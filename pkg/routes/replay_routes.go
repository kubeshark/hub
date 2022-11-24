package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kubeshark/hub/pkg/controllers"
)

// ReplayRoutes defines the group of replay routes.
func ReplayRoutes(app *gin.Engine) {
	routeGroup := app.Group("/replay")

	routeGroup.POST("/", controllers.ReplayRequest)
}
