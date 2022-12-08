package routes

import (
	"net/http"
	"net/url"
	"sync"

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

func WebSocketRoutes(app *gin.Engine, workerHosts []string) {
	app.GET("/ws", func(c *gin.Context) {
		websocketHandler(c, workerHosts)
	})
}

func websocketHandler(c *gin.Context, workerHosts []string) {
	ws, err := websocketUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to set WebSocket upgrade:")
		return
	}

	_, query, err := ws.ReadMessage()
	if err != nil {
		log.Error().Err(err).Msg("WebSocket recieve query:")
	}

	var wg sync.WaitGroup
	for _, workerHost := range workerHosts {
		go func(host string) {
			defer wg.Done()
			u := url.URL{Scheme: "ws", Host: host, Path: "/ws"}

			q := u.Query()
			q.Add("q", string(query))
			u.RawQuery = q.Encode()

			log.Info().Str("url", u.String()).Msg("Connecting to the worker at:")

			wsc, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
			if err != nil {
				log.Error().Err(err).Msg("WebSocket client dial:")
				return
			}
			defer wsc.Close()

			for {
				_, msg, err := wsc.ReadMessage()
				if err != nil {
					log.Error().Err(err).Msg("WebSocket client read:")
					continue
				}

				err = ws.WriteMessage(1, msg)
				if err != nil {
					log.Error().Err(err).Msg("WebSocket server write:")
					continue
				}
			}
		}(workerHost)
	}
	wg.Wait()
}
