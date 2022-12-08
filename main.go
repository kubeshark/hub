package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
	baseApi "github.com/kubeshark/base/pkg/api"
	"github.com/kubeshark/hub/pkg/api"
	"github.com/kubeshark/hub/pkg/app"
	"github.com/kubeshark/hub/pkg/config"
	"github.com/kubeshark/hub/pkg/dependency"
	"github.com/kubeshark/hub/pkg/entries"
	"github.com/kubeshark/hub/pkg/middlewares"
	"github.com/kubeshark/hub/pkg/oas"
	"github.com/kubeshark/hub/pkg/routes"
	"github.com/kubeshark/hub/pkg/servicemap"
	"github.com/kubeshark/hub/pkg/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var namespace = flag.String("namespace", "", "Resolve IPs if they belong to resources in this namespace (default is all)")
var port = flag.Int("port", 80, "Port number of the HTTP server")
var debug = flag.Bool("debug", false, "Enable debug mode")

func main() {
	flag.Parse()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log.Info().Msg("Initializing the Hub...")
	initializeDependencies()

	app.LoadExtensions()

	ginApp := runInApiServerMode(*namespace)

	utils.StartServer(ginApp, *port)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	log.Info().Msg("Exiting")
}

func hostApi(socketHarOutputChannel chan<- *baseApi.OutputChannelItem) *gin.Engine {
	ginApp := gin.Default()

	ginApp.GET("/echo", func(c *gin.Context) {
		c.String(http.StatusOK, "It's running.")
	})

	ginApp.Use(middlewares.CORSMiddleware())

	routes.OASRoutes(ginApp)
	routes.ServiceMapRoutes(ginApp)

	routes.QueryRoutes(ginApp)
	routes.ItemRoutes(ginApp)
	routes.WebSocketRoutes(ginApp)
	routes.MetadataRoutes(ginApp)
	routes.StatusRoutes(ginApp)
	routes.ReplayRoutes(ginApp)

	return ginApp
}

func runInApiServerMode(namespace string) *gin.Engine {
	if err := config.LoadConfig(); err != nil {
		log.Fatal().Err(err).Msg("While loading the config file!")
	}
	api.StartResolving(namespace)

	enableExpFeatures()

	return hostApi(app.GetEntryInputChannel())
}

func enableExpFeatures() {
	oasGenerator := dependency.GetInstance(dependency.OasGeneratorDependency).(oas.OasGenerator)
	oasGenerator.Start()

	serviceMapGenerator := dependency.GetInstance(dependency.ServiceMapGeneratorDependency).(servicemap.ServiceMap)
	serviceMapGenerator.Enable()
}

func initializeDependencies() {
	dependency.RegisterGenerator(dependency.ServiceMapGeneratorDependency, func() interface{} { return servicemap.GetDefaultServiceMapInstance() })
	dependency.RegisterGenerator(dependency.OasGeneratorDependency, func() interface{} { return oas.GetDefaultOasGeneratorInstance(10240) })
	dependency.RegisterGenerator(dependency.EntriesProvider, func() interface{} { return &entries.BasenineEntriesProvider{} })
}
