package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
)

var targettedPods []v1.Pod

func GetTarggetedPods() []v1.Pod {
	return targettedPods
}

func SetTargettedPods(pods []v1.Pod) {
	targettedPods = pods
}

func PostTargettedPodsToWorkers() {
	podsMarshalled, err := json.Marshal(targettedPods)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal the targetted pods:")
	}

	RangeHosts(func(workerHost, v interface{}) bool {
		client := &http.Client{}
		setTargettedUrl := fmt.Sprintf("http://%s/pods/set-targetted", workerHost)
		log.Info().Str("url", setTargettedUrl).Msg("Doing set targetted pods request:")
		client.Post(setTargettedUrl, "application/json", bytes.NewBuffer(podsMarshalled))
		_, err := http.NewRequest(http.MethodPost, setTargettedUrl, nil)
		if err != nil {
			log.Error().Err(err).Str("url", setTargettedUrl).Msg("Set targetted pods request:")
		}

		return true
	})
}
