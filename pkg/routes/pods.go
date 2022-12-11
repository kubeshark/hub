package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kubeshark/hub/pkg/controllers"
)

func PodsRoutes(ginApp *gin.Engine) {
	routeGroup := ginApp.Group("/pods")

	routeGroup.POST("/set-worker", controllers.PostSetWorker)
}
