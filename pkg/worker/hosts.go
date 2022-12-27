package worker

import (
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"
)

const DefaultWorkerHost = "127.0.0.1"
const DefaultWorkerPort = 8897
const DefaultNodeName = "localhost"

var RemovedDefaultHost bool

var workerHosts *sync.Map

func HostWithPort(host string) string {
	return fmt.Sprintf("%s:%d", host, DefaultWorkerPort)
}

func InitHosts() {
	workerHosts = &sync.Map{}
}

func AddHost(host string, name string) {
	workerHosts.Store(host, name)
	log.Info().Str("host", host).Msg("Added worker host:")
}

func AddHosts(hosts []string, name string) {
	for _, host := range hosts {
		AddHost(host, name)
	}
}

func GetHostName(host string) string {
	v, ok := workerHosts.Load(host)
	if ok {
		return v.(string)
	} else {
		return host
	}
}

func RangeHosts(f func(key, value interface{}) bool) {
	workerHosts.Range(f)
}

func RemoveHost(host string) {
	workerHosts.Delete(host)
	log.Warn().Str("host", host).Msg("Removed worker host:")
}
