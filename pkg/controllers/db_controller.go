package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kubeshark/hub/pkg/app"
	"github.com/kubeshark/hub/pkg/config"
	"github.com/kubeshark/hub/pkg/db"
	basenine "github.com/up9inc/basenine/client/go"
)

func Flush(c *gin.Context) {
	if err := basenine.Flush(db.BasenineHost, db.BaseninePort); err != nil {
		c.JSON(http.StatusBadRequest, err)
	} else {
		c.JSON(http.StatusOK, "Flushed.")
	}
}

func Reset(c *gin.Context) {
	if err := basenine.Reset(db.BasenineHost, db.BaseninePort); err != nil {
		c.JSON(http.StatusBadRequest, err)
	} else {
		app.ConfigureBasenineServer(db.BasenineHost, db.BaseninePort, config.Config.MaxDBSizeBytes, config.Config.LogLevel, config.Config.InsertionFilter)
		c.JSON(http.StatusOK, "Resetted.")
	}
}
