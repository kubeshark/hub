package api

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	baseApi "github.com/kubeshark/base/pkg/api"
	"github.com/kubeshark/base/pkg/models"
	"github.com/kubeshark/hub/pkg/utils"
	"github.com/rs/zerolog/log"
)

var (
	extensionsMap map[string]*baseApi.Extension // global
	protocolsMap  map[string]*baseApi.Protocol  //global
)

func InitMaps(extensions map[string]*baseApi.Extension, protocols map[string]*baseApi.Protocol) {
	extensionsMap = extensions
	protocolsMap = protocols
}

type EventHandlers interface {
	WebSocketConnect(c *gin.Context, socketId int, isWorker bool)
	WebSocketDisconnect(socketId int, isWorker bool)
	WebSocketMessage(socketId int, isWorker bool, message []byte)
}

type SocketConnection struct {
	connection    *websocket.Conn
	lock          *sync.Mutex
	eventHandlers EventHandlers
	isWorker      bool
}

type WebSocketParams struct {
	LeftOff           string `json:"leftOff"`
	Query             string `json:"query"`
	EnableFullEntries bool   `json:"enableFullEntries"`
	Fetch             int    `json:"fetch"`
	TimeoutMs         int    `json:"timeoutMs"`
}

var (
	websocketUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	websocketIdsLock            = sync.Mutex{}
	connectedWebsockets         map[int]*SocketConnection
	connectedWebsocketIdCounter = 0
	SocketGetBrowserHandler     gin.HandlerFunc
	SocketGetWorkerHandler      gin.HandlerFunc
)

func init() {
	websocketUpgrader.CheckOrigin = func(r *http.Request) bool { return true } // like cors for web socket
	connectedWebsockets = make(map[int]*SocketConnection)
}

func WebSocketRoutes(app *gin.Engine, eventHandlers EventHandlers) {
	SocketGetBrowserHandler = func(c *gin.Context) {
		websocketHandler(c, eventHandlers, false)
	}

	SocketGetWorkerHandler = func(c *gin.Context) {
		websocketHandler(c, eventHandlers, true)
	}

	app.GET("/ws", func(c *gin.Context) {
		SocketGetBrowserHandler(c)
	})

	app.GET("/wsWorker", func(c *gin.Context) { // TODO: add m2m authentication to this route
		SocketGetWorkerHandler(c)
	})
}

func websocketHandler(c *gin.Context, eventHandlers EventHandlers, isWorker bool) {
	ws, err := websocketUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to set WebSocket upgrade:")
		return
	}

	websocketIdsLock.Lock()

	connectedWebsocketIdCounter++
	socketId := connectedWebsocketIdCounter
	connectedWebsockets[socketId] = &SocketConnection{connection: ws, lock: &sync.Mutex{}, eventHandlers: eventHandlers, isWorker: isWorker}

	websocketIdsLock.Unlock()

	defer func() {
		if socketConnection := connectedWebsockets[socketId]; socketConnection != nil {
			socketCleanup(socketId, socketConnection)
		}
	}()

	eventHandlers.WebSocketConnect(c, socketId, isWorker)

	startTimeBytes, _ := models.CreateWebsocketStartTimeMessage(utils.StartTime)

	if err = SendToSocket(socketId, startTimeBytes); err != nil {
		log.Error().Err(err).Send()
	}

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			if _, ok := err.(*websocket.CloseError); ok {
				log.Debug().Int("socket-id", socketId).Msg("Received WebSocket close message.")
			} else {
				log.Error().Err(err).Int("socket-id", socketId).Msg("While reading WebSocket message!")
			}

			break
		}

		eventHandlers.WebSocketMessage(socketId, isWorker, msg)
	}
}

func SendToSocket(socketId int, message []byte) error {
	socketObj := connectedWebsockets[socketId]
	if socketObj == nil {
		return fmt.Errorf("socket %v is disconnected", socketId)
	}

	socketObj.lock.Lock() // gorilla socket panics from concurrent writes to a single socket
	defer socketObj.lock.Unlock()

	if connectedWebsockets[socketId] == nil {
		return fmt.Errorf("socket %v is disconnected", socketId)
	}

	if err := socketObj.connection.SetWriteDeadline(time.Now().Add(time.Second * 10)); err != nil {
		socketCleanup(socketId, socketObj)
		return fmt.Errorf("error setting timeout to socket %v, err: %v", socketId, err)
	}

	if err := socketObj.connection.WriteMessage(websocket.TextMessage, message); err != nil {
		socketCleanup(socketId, socketObj)
		return fmt.Errorf("failed to write message to socket %v, err: %v", socketId, err)
	}

	return nil
}

func socketCleanup(socketId int, socketConnection *SocketConnection) {
	err := socketConnection.connection.Close()
	if err != nil {
		log.Error().Err(err).Int("socket-id", socketId).Msg("Closing socket connection for:")
	}

	websocketIdsLock.Lock()
	connectedWebsockets[socketId] = nil
	websocketIdsLock.Unlock()

	socketConnection.eventHandlers.WebSocketDisconnect(socketId, socketConnection.isWorker)
}
