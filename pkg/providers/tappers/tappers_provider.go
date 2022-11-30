package tappers

import (
	"log"
	"os"
	"sync"

	"github.com/kubeshark/base/pkg/models"
	"github.com/kubeshark/hub/pkg/utils"
)

const FilePath = models.DataDirPath + "tappers-status.json"

var (
	lockStatus = &sync.Mutex{}
	syncOnce   sync.Once
	status     map[string]*models.TapperStatus

	lockConnectedCount = &sync.Mutex{}
	connectedCount     int
)

func GetStatus() map[string]*models.TapperStatus {
	initStatus()

	return status
}

func SetStatus(tapperStatus *models.TapperStatus) {
	initStatus()

	lockStatus.Lock()
	defer lockStatus.Unlock()

	status[tapperStatus.NodeName] = tapperStatus

	saveStatus()
}

func ResetStatus() {
	lockStatus.Lock()
	defer lockStatus.Unlock()

	status = make(map[string]*models.TapperStatus)

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
			status = make(map[string]*models.TapperStatus)

			if !os.IsNotExist(err) {
				log.Printf("Error reading tappers status from file, err: %v", err)
			}
		}
	})
}

func saveStatus() {
	if err := utils.SaveJsonFile(FilePath, status); err != nil {
		log.Printf("Error saving tappers status, err: %v", err)
	}
}
