package misc

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
)

var dataDir = "data"

func InitDataDir() {
	body, err := os.ReadFile("/etc/machine-id")
	newDataDir := dataDir
	if err == nil {
		machineId := strings.TrimSpace(string(body))
		log.Info().Str("id", machineId).Msg("Machine ID is:")
		newDataDir = fmt.Sprintf("%s/%s", dataDir, machineId)
	}
	err = os.MkdirAll(newDataDir, os.ModePerm)
	if err != nil {
		log.Error().Err(err).Str("data-dir", newDataDir).Msg("Unable to create the new data directory:")
	} else {
		dataDir = newDataDir
		log.Info().Str("data-dir", dataDir).Msg("Set the data directory to:")
	}
}

func GetDataDir() string {
	return dataDir
}

func BuildDataFilePath(dir string, filename string) string {
	if dir == "" {
		return fmt.Sprintf("%s/%s", GetDataDir(), filename)
	} else {
		return fmt.Sprintf("%s/%s/%s", GetDataDir(), dir, filename)
	}
}
