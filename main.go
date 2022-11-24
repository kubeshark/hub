package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/kubeshark/hub/pkg/db"
	"github.com/kubeshark/hub/pkg/dependency"
	"github.com/kubeshark/hub/pkg/entries"
	"github.com/kubeshark/hub/pkg/middlewares"
	"github.com/kubeshark/hub/pkg/oas"
	"github.com/kubeshark/hub/pkg/routes"
	"github.com/kubeshark/hub/pkg/servicemap"
	"github.com/kubeshark/hub/pkg/utils"

	"github.com/kubeshark/hub/pkg/api"
	"github.com/kubeshark/hub/pkg/app"
	"github.com/kubeshark/hub/pkg/config"

	tapApi "github.com/kubeshark/worker/api"
)

var namespace = flag.String("namespace", "", "Resolve IPs if they belong to resources in this namespace (default is all)")
var port = flag.Int("port", 80, "Port number of the HTTP server")
var profiler = flag.Bool("profiler", false, "Run pprof server")

func main() {
	fmt.Println("Initializing the Hub")
	initializeDependencies()
	flag.Parse()

	app.LoadExtensions()

	ginApp := runInApiServerMode(*namespace)

	if *profiler {
		pprof.Register(ginApp)
	}

	utils.StartServer(ginApp, *port)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	log.Print("Exiting")
}

func hostApi(socketHarOutputChannel chan<- *tapApi.OutputChannelItem) *gin.Engine {
	ginApp := gin.Default()

	ginApp.GET("/echo", func(c *gin.Context) {
		c.JSON(http.StatusOK, "Here is Kubeshark agent")
	})

	eventHandlers := api.RoutesEventHandlers{
		SocketOutChannel: socketHarOutputChannel,
	}

	ginApp.Use(middlewares.CORSMiddleware())

	api.WebSocketRoutes(ginApp, &eventHandlers)

	if config.Config.OAS.Enable {
		routes.OASRoutes(ginApp)
	}

	if config.Config.ServiceMap {
		routes.ServiceMapRoutes(ginApp)
	}

	routes.QueryRoutes(ginApp)
	routes.EntriesRoutes(ginApp)
	routes.MetadataRoutes(ginApp)
	routes.StatusRoutes(ginApp)
	routes.DbRoutes(ginApp)
	routes.ReplayRoutes(ginApp)

	return ginApp
}

func runInApiServerMode(namespace string) *gin.Engine {
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Error loading config file %v", err)
	}
	app.ConfigureBasenineServer(db.BasenineHost, db.BaseninePort, config.Config.MaxDBSizeBytes, config.Config.LogLevel, config.Config.InsertionFilter)
	api.StartResolving(namespace)

	enableExpFeatureIfNeeded()

	return hostApi(app.GetEntryInputChannel())
}

func enableExpFeatureIfNeeded() {
	if config.Config.OAS.Enable {
		oasGenerator := dependency.GetInstance(dependency.OasGeneratorDependency).(oas.OasGenerator)
		oasGenerator.Start()
	}
	if config.Config.ServiceMap {
		serviceMapGenerator := dependency.GetInstance(dependency.ServiceMapGeneratorDependency).(servicemap.ServiceMap)
		serviceMapGenerator.Enable()
	}
}

func initializeDependencies() {
	dependency.RegisterGenerator(dependency.ServiceMapGeneratorDependency, func() interface{} { return servicemap.GetDefaultServiceMapInstance() })
	dependency.RegisterGenerator(dependency.OasGeneratorDependency, func() interface{} { return oas.GetDefaultOasGeneratorInstance(config.Config.OAS.MaxExampleLen) })
	dependency.RegisterGenerator(dependency.EntriesInserter, func() interface{} { return api.GetBasenineEntryInserterInstance() })
	dependency.RegisterGenerator(dependency.EntriesProvider, func() interface{} { return &entries.BasenineEntriesProvider{} })
	dependency.RegisterGenerator(dependency.EntriesSocketStreamer, func() interface{} { return &api.BasenineEntryStreamer{} })
	dependency.RegisterGenerator(dependency.EntryStreamerSocketConnector, func() interface{} { return &api.DefaultEntryStreamerSocketConnector{} })
}
