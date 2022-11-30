package api

import (
	"encoding/json"
	"log"

	"github.com/kubeshark/base/pkg/models"
	"github.com/kubeshark/hub/pkg/providers/targettedPods"
)

func BroadcastTargettedPodsStatus() {
	targettedPodsStatus := targettedPods.GetTargettedPodsStatus()

	message := models.CreateWebSocketStatusMessage(targettedPodsStatus)
	if jsonBytes, err := json.Marshal(message); err != nil {
		log.Printf("Could not Marshal message %v", err)
	} else {
		BroadcastToBrowserClients(jsonBytes)
	}
}

func SendTargettedPods(socketId int, nodeToTargettedPodMap models.NodeToPodsMap) {
	message := models.CreateWebSocketTargettedPodsMessage(nodeToTargettedPodMap)
	if jsonBytes, err := json.Marshal(message); err != nil {
		log.Printf("Could not Marshal message %v", err)
	} else {
		if err := SendToSocket(socketId, jsonBytes); err != nil {
			log.Print(err)
		}
	}
}

func BroadcastTargettedPodsToWorkers(nodeToTargettedPodMap models.NodeToPodsMap) {
	message := models.CreateWebSocketTargettedPodsMessage(nodeToTargettedPodMap)
	if jsonBytes, err := json.Marshal(message); err != nil {
		log.Printf("Could not Marshal message %v", err)
	} else {
		BroadcastToWorkerClients(jsonBytes)
	}
}
