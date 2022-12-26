package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kubeshark/hub/server/controllers"
)

func ItemRoutes(ginApp *gin.Engine) {
	routeGroup := ginApp.Group("/item")

	routeGroup.GET("/:worker/:id", controllers.GetItem)
}
