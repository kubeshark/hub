package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
)

type PostStorageLimit struct {
	Limit int64 `json:"limit"`
}

func PostStorageLimitToWorkers(payload PostStorageLimit) error {
	payloadMarshalled, err := json.Marshal(payload)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal the targeted pods:")
		return err
	}

	RangeHosts(func(workerHost, v interface{}) bool {
		client := &http.Client{}
		setStorageLimitUrl := fmt.Sprintf("http://%s/pcaps/set-storage-limit", workerHost)
		log.Info().Str("url", setStorageLimitUrl).Msg("Doing set targeted pods request:")
		_, err = client.Post(setStorageLimitUrl, "application/json", bytes.NewBuffer(payloadMarshalled))
		if err != nil {
			log.Error().Err(err).Str("url", setStorageLimitUrl).Msg("Set targeted pods request:")
		}

		return true
	})

	return err
}
