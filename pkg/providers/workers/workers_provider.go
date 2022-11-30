package workers

import (
	"os"
	"sync"

	"github.com/kubeshark/base/pkg/models"
	"github.com/kubeshark/hub/pkg/utils"
	"github.com/rs/zerolog/log"
)

const FilePath = models.DataDirPath + "workers-status.json"

var (
	lockStatus = &sync.Mutex{}
	syncOnce   sync.Once
	status     map[string]*models.WorkerStatus

	lockConnectedCount = &sync.Mutex{}
	connectedCount     int
)

func GetStatus() map[string]*models.WorkerStatus {
	initStatus()

	return status
}

func SetStatus(workerStatus *models.WorkerStatus) {
	initStatus()

	lockStatus.Lock()
	defer lockStatus.Unlock()

	status[workerStatus.NodeName] = workerStatus

	saveStatus()
}

func ResetStatus() {
	lockStatus.Lock()
	defer lockStatus.Unlock()

	status = make(map[string]*models.WorkerStatus)

	saveStatus()
}

func GetConnectedCount() int {
	return connectedCount
}

func Connected() {
	lockConnectedCount.Lock()
	defer lockConnectedCount.Unlock()

	connectedCount++
}

func Disconnected() {
	lockConnectedCount.Lock()
	defer lockConnectedCount.Unlock()

	connectedCount--
}

func initStatus() {
	syncOnce.Do(func() {
		if err := utils.ReadJsonFile(FilePath, &status); err != nil {
			status = make(map[string]*models.WorkerStatus)

			if !os.IsNotExist(err) {
				log.Error().Err(err).Msg("While reading workers status from file.")
			}
		}
	})
}

func saveStatus() {
	if err := utils.SaveJsonFile(FilePath, status); err != nil {
		log.Error().Err(err).Msg("While saving workers status.")
	}
}
