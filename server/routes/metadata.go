package routes

import (
	"github.com/kubeshark/hub/server/controllers"

	"github.com/gin-gonic/gin"
)

// MetadataRoutes defines the group of metadata routes.
func MetadataRoutes(app *gin.Engine) {
	routeGroup := app.Group("/metadata")

	routeGroup.GET("/version", controllers.GetVersion)
}
