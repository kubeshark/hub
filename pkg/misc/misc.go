package misc

import (
	"net"

	"github.com/rs/zerolog/log"
)

func RemovePortFromWorkerHost(workerHost string) string {
	host, _, err := net.SplitHostPort(workerHost)
	if err != nil {
		log.Error().Err(err).Send()
		return workerHost
	}

	return host
}
