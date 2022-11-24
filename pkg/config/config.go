package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kubeshark/worker/models"
)

// these values are used when the config.json file is not present
const (
	defaultMaxDatabaseSizeBytes int64  = 200 * 1000 * 1000
	DefaultDatabasePath         string = "./entries"
)

var Config *models.Config

func LoadConfig() error {
	if Config != nil {
		return nil
	}
	filePath := fmt.Sprintf("%s%s", models.ConfigDirPath, models.ConfigFileName)

	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return applyDefaultConfig()
		}
		return err
	}

	if err = json.Unmarshal(content, &Config); err != nil {
		return err
	}
	return nil
}

func applyDefaultConfig() error {
	defaultConfig, err := getDefaultConfig()
	if err != nil {
		return err
	}
	Config = defaultConfig
	return nil
}

func getDefaultConfig() (*models.Config, error) {
	return &models.Config{
		MaxDBSizeBytes:    defaultMaxDatabaseSizeBytes,
		AgentDatabasePath: DefaultDatabasePath,
	}, nil
}
