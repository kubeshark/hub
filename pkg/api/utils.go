package api

import (
	"encoding/json"
	"log"

	"github.com/kubeshark/hub/pkg/providers/tappedPods"
	"github.com/kubeshark/worker/models"
)

func BroadcastTappedPodsStatus() {
	tappedPodsStatus := tappedPods.GetTappedPodsStatus()

	message := models.CreateWebSocketStatusMessage(tappedPodsStatus)
	if jsonBytes, err := json.Marshal(message); err != nil {
		log.Printf("Could not Marshal message %v", err)
	} else {
		BroadcastToBrowserClients(jsonBytes)
	}
}

func SendTappedPods(socketId int, nodeToTappedPodMap models.NodeToPodsMap) {
	message := models.CreateWebSocketTappedPodsMessage(nodeToTappedPodMap)
	if jsonBytes, err := json.Marshal(message); err != nil {
		log.Printf("Could not Marshal message %v", err)
	} else {
		if err := SendToSocket(socketId, jsonBytes); err != nil {
			log.Print(err)
		}
	}
}

func BroadcastTappedPodsToTappers(nodeToTappedPodMap models.NodeToPodsMap) {
	message := models.CreateWebSocketTappedPodsMessage(nodeToTappedPodMap)
	if jsonBytes, err := json.Marshal(message); err != nil {
		log.Printf("Could not Marshal message %v", err)
	} else {
		BroadcastToTapperClients(jsonBytes)
	}
}
