package api

import (
	"encoding/json"

	"github.com/kubeshark/base/pkg/models"
	"github.com/kubeshark/hub/pkg/providers/targettedPods"
	"github.com/rs/zerolog/log"
)

func BroadcastTargettedPodsStatus() {
	targettedPodsStatus := targettedPods.GetTargettedPodsStatus()

	message := models.CreateWebSocketStatusMessage(targettedPodsStatus)
	if jsonBytes, err := json.Marshal(message); err != nil {
		log.Error().Err(err).Msg("Couldn't marshal message:")
	} else {
		BroadcastToBrowserClients(jsonBytes)
	}
}

func SendTargettedPods(socketId int, nodeToTargettedPodMap models.NodeToPodsMap) {
	message := models.CreateWebSocketTargettedPodsMessage(nodeToTargettedPodMap)
	if jsonBytes, err := json.Marshal(message); err != nil {
		log.Error().Err(err).Msg("Couldn't marshal message:")
	} else {
		if err := SendToSocket(socketId, jsonBytes); err != nil {
			log.Error().Err(err).Send()
		}
	}
}

func BroadcastTargettedPodsToWorkers(nodeToTargettedPodMap models.NodeToPodsMap) {
	message := models.CreateWebSocketTargettedPodsMessage(nodeToTargettedPodMap)
	if jsonBytes, err := json.Marshal(message); err != nil {
		log.Error().Err(err).Msg("Couldn't marshal message:")
	} else {
		BroadcastToWorkerClients(jsonBytes)
	}
}
