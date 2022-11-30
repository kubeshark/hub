package targettedPods

import (
	"os"
	"strings"
	"sync"

	"github.com/kubeshark/base/pkg/models"
	"github.com/kubeshark/hub/pkg/providers/workers"
	"github.com/kubeshark/hub/pkg/utils"
	"github.com/rs/zerolog/log"
)

const FilePath = models.DataDirPath + "targetted-pods.json"

var (
	lock                       = &sync.Mutex{}
	syncOnce                   sync.Once
	targettedPods              []*models.PodInfo
	nodeHostToTargettedPodsMap models.NodeToPodsMap
)

func Get() []*models.PodInfo {
	syncOnce.Do(func() {
		if err := utils.ReadJsonFile(FilePath, &targettedPods); err != nil {
			if !os.IsNotExist(err) {
				log.Error().Err(err).Msg("While reading targetted pods form file.")
			}
		}
	})

	return targettedPods
}

func Set(targettedPodsToSet []*models.PodInfo) {
	lock.Lock()
	defer lock.Unlock()

	targettedPods = targettedPodsToSet
	if err := utils.SaveJsonFile(FilePath, targettedPods); err != nil {
		log.Error().Err(err).Msg("While saving targetted pods.")
	}
}

func GetTargettedPodsStatus() []models.TargettedPodStatus {
	targettedPodsStatus := make([]models.TargettedPodStatus, 0)
	workerStatus := workers.GetStatus()
	for _, pod := range Get() {
		var status string
		if workerStatus, ok := workerStatus[pod.NodeName]; ok {
			status = strings.ToLower(workerStatus.Status)
		}

		IsTargetted := status == "running"
		targettedPodsStatus = append(targettedPodsStatus, models.TargettedPodStatus{Name: pod.Name, Namespace: pod.Namespace, IsTargetted: IsTargetted})
	}

	return targettedPodsStatus
}

func SetNodeToTargettedPodMap(nodeToTargettedPodsMap models.NodeToPodsMap) {
	summary := nodeToTargettedPodsMap.Summary()
	log.Debug().Interface("summary", summary).Msg("Setting node to targetted pods map:")

	nodeHostToTargettedPodsMap = nodeToTargettedPodsMap
}

func GetNodeToTargettedPodsMap() models.NodeToPodsMap {
	return nodeHostToTargettedPodsMap
}
