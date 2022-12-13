package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kubeshark/hub/pkg/controllers"
)

func PodsRoutes(ginApp *gin.Engine) {
	routeGroup := ginApp.Group("/pods")

	routeGroup.POST("/worker", controllers.PostWorker)

	routeGroup.GET("/targetted", controllers.GetTargetted)
	routeGroup.POST("/targetted", controllers.PostTargetted)
}
