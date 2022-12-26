package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kubeshark/hub/server/controllers"
)

func WebSocketRoutes(app *gin.Engine) {
	app.GET("/ws", controllers.WebsocketHandler)
}
