package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

var (
	StartTime int64 // global
)

// StartServer starts the server with a graceful shutdown
func StartServer(app *gin.Engine, port int) {
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

func CheckErr(err error) {
	if err != nil {
		log.Error().Err(err).Send()
	}
}

func ReadJsonFile(filePath string, value interface{}) error {
	if content, err := os.ReadFile(filePath); err != nil {
		return err
	} else {
		if err = json.Unmarshal(content, value); err != nil {
			return err
		}
	}

	return nil
}

func SaveJsonFile(filePath string, value interface{}) error {
	if data, err := json.Marshal(value); err != nil {
		return err
	} else {
		if err = os.WriteFile(filePath, data, 0644); err != nil {
			return err
		}
	}

	return nil
}

func UniqueStringSlice(s []string) []string {
	uniqueSlice := make([]string, 0)
	uniqueMap := map[string]bool{}

	for _, val := range s {
		if uniqueMap[val] {
			continue
		}
		uniqueMap[val] = true
		uniqueSlice = append(uniqueSlice, val)
	}

	return uniqueSlice
}
