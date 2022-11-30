package tappedPods

import (
	"log"
	"os"
	"strings"
	"sync"

	"github.com/kubeshark/base/pkg/models"
	"github.com/kubeshark/hub/pkg/providers/tappers"
	"github.com/kubeshark/hub/pkg/utils"
)

const FilePath = models.DataDirPath + "tapped-pods.json"

var (
	lock                    = &sync.Mutex{}
	syncOnce                sync.Once
	tappedPods              []*models.PodInfo
	nodeHostToTappedPodsMap models.NodeToPodsMap
)

func Get() []*models.PodInfo {
	syncOnce.Do(func() {
		if err := utils.ReadJsonFile(FilePath, &tappedPods); err != nil {
			if !os.IsNotExist(err) {
				log.Printf("Error reading tapped pods from file, err: %v", err)
			}
		}
	})

	return tappedPods
}

func Set(tappedPodsToSet []*models.PodInfo) {
	lock.Lock()
	defer lock.Unlock()

	tappedPods = tappedPodsToSet
	if err := utils.SaveJsonFile(FilePath, tappedPods); err != nil {
		log.Printf("Error saving tapped pods, err: %v", err)
	}
}

func GetTappedPodsStatus() []models.TappedPodStatus {
	tappedPodsStatus := make([]models.TappedPodStatus, 0)
	tapperStatus := tappers.GetStatus()
	for _, pod := range Get() {
		var status string
		if tapperStatus, ok := tapperStatus[pod.NodeName]; ok {
			status = strings.ToLower(tapperStatus.Status)
		}

		isTapped := status == "running"
		tappedPodsStatus = append(tappedPodsStatus, models.TappedPodStatus{Name: pod.Name, Namespace: pod.Namespace, IsTapped: isTapped})
	}

	return tappedPodsStatus
}

func SetNodeToTappedPodMap(nodeToTappedPodsMap models.NodeToPodsMap) {
	summary := nodeToTappedPodsMap.Summary()
	log.Printf("Setting node to tapped pods map to %v", summary)

	nodeHostToTappedPodsMap = nodeToTappedPodsMap
}

func GetNodeToTappedPodMap() models.NodeToPodsMap {
	return nodeHostToTappedPodsMap
}
