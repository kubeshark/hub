package controllers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kubeshark/hub/pkg/replay"
	"github.com/kubeshark/hub/pkg/validation"
)

const (
	replayTimeout = 10 * time.Second
)

func ReplayRequest(c *gin.Context) {
	log.Print("Starting replay")
	replayDetails := &replay.Details{}
	if err := c.Bind(replayDetails); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	log.Printf("Validating replay, %v", replayDetails)
	if err := validation.Validate(replayDetails); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	log.Print("Executing replay, %v", replayDetails)
	result := replay.ExecuteRequest(replayDetails, replayTimeout)
	c.JSON(http.StatusOK, result)
}
