package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kubeshark/hub/server/controllers"
)

func QueryRoutes(ginApp *gin.Engine) {
	routeGroup := ginApp.Group("/query")

	routeGroup.GET("/validate", controllers.GetValidate)
}
