package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kubeshark/hub/server/controllers"
)

func PodsRoutes(ginApp *gin.Engine) {
	routeGroup := ginApp.Group("/pods")

	routeGroup.POST("/worker", controllers.PostWorker)

	routeGroup.GET("/targeted", controllers.GetTargeted)
	routeGroup.POST("/targeted", controllers.PostTargeted)

	// For backward compatibility (38.0, 38.1 has this typo)
	routeGroup.GET("/targetted", controllers.GetTargeted)
	routeGroup.POST("/targetted", controllers.PostTargeted)
}
