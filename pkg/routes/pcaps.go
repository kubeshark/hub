package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kubeshark/hub/pkg/controllers"
)

func PcapsRoutes(ginApp *gin.Engine) {
	routeGroup := ginApp.Group("/pcaps")

	routeGroup.GET("/total-tcp-streams", controllers.GetTotalTcpStreams)
	routeGroup.GET("/merge", controllers.GetMerge)
}
