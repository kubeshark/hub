package main

import (
	"flag"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/kubeshark/hub/pkg/config"
	"github.com/kubeshark/hub/pkg/misc"
	"github.com/kubeshark/hub/pkg/worker"
	"github.com/kubeshark/hub/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var port = flag.Int("port", 80, "Port number of the HTTP server")
var debug = flag.Bool("debug", false, "Enable debug mode")
var workerHostsFlag = flag.String("worker-hosts", worker.HostWithPort(worker.DefaultWorkerHost), "hostname:port pairs of worker instances to access their WebSocket and HTTP endpoints")

func main() {
	flag.Parse()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).With().Caller().Logger()

	misc.InitDataDir()
	worker.InitHosts()
	worker.AddHosts(strings.Split(*workerHostsFlag, " "))

	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log.Info().Msg("Initializing the Hub...")

	if err := config.LoadConfig(); err != nil {
		log.Fatal().Err(err).Msg("While loading the config file!")
	}

	ginApp := server.Build()

	server.Start(ginApp, *port)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	log.Info().Msg("Exiting")
}
