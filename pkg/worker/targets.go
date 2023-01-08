package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
)

var targetedPods []v1.Pod

func GetTarggetedPods() []v1.Pod {
	return targetedPods
}

func SetTargetedPods(pods []v1.Pod) {
	targetedPods = pods
}

func PostTargetedPodsToWorkers() error {
	podsMarshalled, err := json.Marshal(targetedPods)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal the targeted pods:")
		return err
	}

	RangeHosts(func(workerHost, v interface{}) bool {
		client := &http.Client{}
		setTargetedUrl := fmt.Sprintf("http://%s/pods/set-targeted", workerHost)
		log.Info().Str("url", setTargetedUrl).Msg("Doing set targeted pods request:")
		_, err = client.Post(setTargetedUrl, "application/json", bytes.NewBuffer(podsMarshalled))
		if err != nil {
			log.Error().Err(err).Str("url", setTargetedUrl).Msg("Set targeted pods request:")
		}

		return true
	})

	return err
}
