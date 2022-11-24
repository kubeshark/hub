package api

import (
	"encoding/json"
	"log"

	"github.com/kubeshark/hub/pkg/providers/tappedPods"
	"github.com/kubeshark/kubeshark/shared"
)

func BroadcastTappedPodsStatus() {
	tappedPodsStatus := tappedPods.GetTappedPodsStatus()

	message := shared.CreateWebSocketStatusMessage(tappedPodsStatus)
	if jsonBytes, err := json.Marshal(message); err != nil {
		log.Printf("Could not Marshal message %v", err)
	} else {
		BroadcastToBrowserClients(jsonBytes)
	}
}

func SendTappedPods(socketId int, nodeToTappedPodMap shared.NodeToPodsMap) {
	message := shared.CreateWebSocketTappedPodsMessage(nodeToTappedPodMap)
	if jsonBytes, err := json.Marshal(message); err != nil {
		log.Printf("Could not Marshal message %v", err)
	} else {
		if err := SendToSocket(socketId, jsonBytes); err != nil {
			log.Print(err)
		}
	}
}

func BroadcastTappedPodsToTappers(nodeToTappedPodMap shared.NodeToPodsMap) {
	message := shared.CreateWebSocketTappedPodsMessage(nodeToTappedPodMap)
	if jsonBytes, err := json.Marshal(message); err != nil {
		log.Printf("Could not Marshal message %v", err)
	} else {
		BroadcastToTapperClients(jsonBytes)
	}
}
