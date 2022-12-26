package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kubeshark/hub/pkg/middlewares"
	"github.com/kubeshark/hub/pkg/routes"
	"github.com/rs/zerolog/log"
)

func Build() *gin.Engine {
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

// Start starts the server with a graceful shutdown
func Start(app *gin.Engine, port int) {
	signals := make(chan os.Signal, 2)
	signal.Notify(signals,
		os.Interrupt,    // this catch ctrl + c
		syscall.SIGTSTP, // this catch ctrl + z
	)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: app,
	}

	go func() {
		<-signals
		log.Warn().Msg("Shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := srv.Shutdown(ctx)
		if err != nil {
			log.Error().Err(err).Send()
		}
		os.Exit(0)
	}()

	// Run server.
	log.Info().Msg("Starting the server...")
	if err := app.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Error().Err(err).Msg("Server is not running!")
	}
}
