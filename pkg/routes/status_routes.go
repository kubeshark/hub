package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kubeshark/hub/pkg/controllers"
)

func StatusRoutes(ginApp *gin.Engine) {
	routeGroup := ginApp.Group("/status")

	routeGroup.GET("/health", controllers.HealthCheck)

	routeGroup.POST("/targettedPods", controllers.PostTargettedPods)
	routeGroup.POST("/workerStatus", controllers.PostWorkerStatus)
	routeGroup.GET("/connectedWorkersCount", controllers.GetConnectedWorkersCount)
	routeGroup.GET("/target", controllers.GetTargettingStatus)

	routeGroup.GET("/general", controllers.GetGeneralStats)
	routeGroup.GET("/trafficStats", controllers.GetTrafficStats)

	routeGroup.GET("/resolving", controllers.GetCurrentResolvingInformation)
}
