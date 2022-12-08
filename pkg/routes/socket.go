package routes

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

var (
	websocketUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func init() {
	websocketUpgrader.CheckOrigin = func(r *http.Request) bool { return true } // like cors for web socket
}

func WebSocketRoutes(app *gin.Engine) {
	app.GET("/ws", websocketHandler)
}

func websocketHandler(c *gin.Context) {
	ws, err := websocketUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to set WebSocket upgrade:")
		return
	}

	u := url.URL{Scheme: "ws", Host: "localhost:8897", Path: "/ws"}
	log.Printf("connecting to %s", u.String())

	wsd, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Error().Err(err).Msg("WebSocket client dial:")
		return
	}
	defer wsd.Close()

	for {
		_, msg, err := wsd.ReadMessage()
		if err != nil {
			log.Error().Err(err).Msg("WebSocket client read:")
			return
		}

		ws.WriteMessage(1, msg)
	}
}
