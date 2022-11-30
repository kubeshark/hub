package api

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/gin-gonic/gin"
	baseApi "github.com/kubeshark/base/pkg/api"
	"github.com/kubeshark/base/pkg/models"
	"github.com/kubeshark/hub/pkg/dependency"
	"github.com/kubeshark/hub/pkg/providers/targettedPods"
	"github.com/kubeshark/hub/pkg/providers/workers"
	"github.com/rs/zerolog/log"
)

type BrowserClient struct {
	dataStreamCancelFunc context.CancelFunc
}

var browserClients = make(map[int]*BrowserClient, 0)
var workerClientSocketUUIDs = make([]int, 0)
var socketListLock = sync.Mutex{}

type RoutesEventHandlers struct {
	EventHandlers
	SocketOutChannel chan<- *baseApi.OutputChannelItem
}

func (h *RoutesEventHandlers) WebSocketConnect(_ *gin.Context, socketId int, isWorker bool) {
	if isWorker {
		log.Info().Int("socket-id", socketId).Msg("Worker connected.")
		workers.Connected()

		socketListLock.Lock()
		workerClientSocketUUIDs = append(workerClientSocketUUIDs, socketId)
		socketListLock.Unlock()

		nodeToTargettedPodsMap := targettedPods.GetNodeToTargettedPodsMap()
		SendTargettedPods(socketId, nodeToTargettedPodsMap)
	} else {
		log.Info().Int("socket-id", socketId).Msg("Browser connected.")

		socketListLock.Lock()
		browserClients[socketId] = &BrowserClient{}
		socketListLock.Unlock()

		BroadcastTargettedPodsStatus()
	}
}

func (h *RoutesEventHandlers) WebSocketDisconnect(socketId int, isWorker bool) {
	if isWorker {
		log.Info().Int("socket-id", socketId).Msg("Worker disconnected.")
		workers.Disconnected()

		socketListLock.Lock()
		removeSocketUUIDFromWorkerSlice(socketId)
		socketListLock.Unlock()
	} else {
		log.Info().Int("socket-id", socketId).Msg("Browser disconnected.")
		socketListLock.Lock()
		if browserClients[socketId] != nil && browserClients[socketId].dataStreamCancelFunc != nil {
			browserClients[socketId].dataStreamCancelFunc()
		}
		delete(browserClients, socketId)
		socketListLock.Unlock()
	}
}

func BroadcastToBrowserClients(message []byte) {
	for socketId := range browserClients {
		go func(socketId int) {
			if err := SendToSocket(socketId, message); err != nil {
				log.Error().Err(err).Int("socket-id", socketId).Send()
			}
		}(socketId)
	}
}

func BroadcastToWorkerClients(message []byte) {
	for _, socketId := range workerClientSocketUUIDs {
		go func(socketId int) {
			if err := SendToSocket(socketId, message); err != nil {
				log.Error().Err(err).Int("socket-id", socketId).Send()
			}
		}(socketId)
	}
}

func (h *RoutesEventHandlers) WebSocketMessage(socketId int, isWorker bool, message []byte) {
	if isWorker {
		HandleWorkerIncomingMessage(message, h.SocketOutChannel, BroadcastToBrowserClients)
	} else {
		// we initiate the basenine stream after the first websocket message we receive (it contains the entry query), we then store a cancelfunc to later cancel this stream
		if browserClients[socketId] != nil && browserClients[socketId].dataStreamCancelFunc == nil {
			var params WebSocketParams
			if err := json.Unmarshal(message, &params); err != nil {
				log.Error().Err(err).Int("socket-id", socketId).Send()
				return
			}

			entriesStreamer := dependency.GetInstance(dependency.EntriesSocketStreamer).(EntryStreamer)
			ctx, cancelFunc := context.WithCancel(context.Background())
			err := entriesStreamer.Get(ctx, socketId, &params)

			if err != nil {
				log.Error().Err(err).Int("socket-id", socketId).Msg("While initializing a Basenine stream for the browser socket!")
				cancelFunc()
			} else {
				browserClients[socketId].dataStreamCancelFunc = cancelFunc
			}
		}
	}
}

func HandleWorkerIncomingMessage(message []byte, socketOutChannel chan<- *baseApi.OutputChannelItem, broadcastMessageFunc func([]byte)) {
	var socketMessageBase models.WebSocketMessageMetadata
	err := json.Unmarshal(message, &socketMessageBase)
	if err != nil {
		log.Error().Err(err).Msg("Couldn't unmarshal WebSocket message:")
	} else {
		switch socketMessageBase.MessageType {
		case models.WebSocketMessageTypeWorkerEntry:
			var workerEntryMessage models.WebSocketWorkerEntryMessage
			err := json.Unmarshal(message, &workerEntryMessage)
			if err != nil {
				log.Error().Err(err).Str("msg-type", string(socketMessageBase.MessageType)).Msg("Couldn't unmarshal message of message type:")
			} else {
				// NOTE: This is where the message comes back from the intermediate WebSocket to code.
				socketOutChannel <- workerEntryMessage.Data
			}
		case models.WebSocketMessageTypeUpdateStatus:
			var statusMessage models.WebSocketStatusMessage
			err := json.Unmarshal(message, &statusMessage)
			if err != nil {
				log.Error().Err(err).Str("msg-type", string(socketMessageBase.MessageType)).Msg("Couldn't unmarshal message of message type:")
			} else {
				broadcastMessageFunc(message)
			}
		default:
			log.Error().Str("msg-type", string(socketMessageBase.MessageType)).Msg("Received a socket message type which no handlers are defined for!")
		}
	}
}

func removeSocketUUIDFromWorkerSlice(uuidToRemove int) {
	newUUIDSlice := make([]int, 0, len(workerClientSocketUUIDs))
	for _, uuid := range workerClientSocketUUIDs {
		if uuid != uuidToRemove {
			newUUIDSlice = append(newUUIDSlice, uuid)
		}
	}
	workerClientSocketUUIDs = newUUIDSlice
}
