package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kubeshark/hub/pkg/config"
	"github.com/kubeshark/hub/pkg/middlewares"
	"github.com/kubeshark/hub/pkg/misc"
	"github.com/kubeshark/hub/pkg/routes"
	"github.com/kubeshark/hub/pkg/utils"
	"github.com/kubeshark/hub/pkg/worker"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var namespace = flag.String("namespace", "", "Resolve IPs if they belong to resources in this namespace (default is all)")
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

	ginApp := runInApiServerMode(*namespace)

	utils.StartServer(ginApp, *port)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	log.Info().Msg("Exiting")
}

func hostApi() *gin.Engine {
	ginApp := gin.New()
	ginApp.Use(middlewares.DefaultStructuredLogger())
	ginApp.Use(gin.Recovery())

	ginApp.GET("/echo", func(c *gin.Context) {
		c.String(http.StatusOK, "It's running.")
	})

	ginApp.Use(middlewares.CORSMiddleware())

	routes.QueryRoutes(ginApp)
	routes.ItemRoutes(ginApp)
	routes.WebSocketRoutes(ginApp)
	routes.MetadataRoutes(ginApp)
	routes.PodsRoutes(ginApp)
	routes.PcapsRoutes(ginApp)

	return ginApp
}

func runInApiServerMode(namespace string) *gin.Engine {
	if err := config.LoadConfig(); err != nil {
		log.Fatal().Err(err).Msg("While loading the config file!")
	}

	return hostApi()
}
