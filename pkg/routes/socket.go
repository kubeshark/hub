package routes

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/kubeshark/hub/pkg/worker"
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
	defer ws.Close()

	_, query, err := ws.ReadMessage()
	if err != nil {
		log.Error().Err(err).Msg("WebSocket recieve query:")
		return
	}

	done := make(chan bool, 1)

	go func() {
		for {
			_, _, err := ws.ReadMessage()
			if err != nil {
				log.Debug().Err(err).Msg("WebSocket read:")
				done <- true
				return
			}
		}
	}()

	var rangeCount uint64
	worker.RangeHosts(func(workerHost, v interface{}) bool {
		rangeCount++
		go func(host string) {
			u := url.URL{Scheme: "ws", Host: host, Path: "/ws"}

			q := u.Query()
			q.Add("q", string(query))
			q.Add("worker", host)
			u.RawQuery = q.Encode()

			log.Info().Str("url", u.String()).Msg("Connecting to the worker at:")

			wsc, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
			if err != nil {
				log.Error().Err(err).Msg("WebSocket client dial:")
				done <- true
				return
			}
			defer wsc.Close()

			log.Info().Str("url", u.String()).Msg("Connected to the worker at:")

			for {
				_, msg, err := wsc.ReadMessage()
				if err != nil {
					log.Error().Err(err).Msg("WebSocket client read:")
					break
				}

				var object map[string]interface{}
				if err := json.Unmarshal(msg, &object); err != nil {
					log.Error().Err(err).Msg("WebSocket failed unmarshalling item:")
					continue
				}

				var data []byte
				data, err = json.Marshal(object)
				if err != nil {
					log.Error().Err(err).Msg("WebSocket failed marshalling item:")
					break
				}

				err = ws.WriteMessage(1, data)
				if err != nil {
					log.Error().Err(err).Msg("WebSocket server write:")
					break
				}
			}

			done <- true
		}(workerHost.(string))

		return true
	})

	// Workaround for empty `workerHosts *sync.Map` case
	if rangeCount == 0 {
		done <- true
	}

	<-done
}
