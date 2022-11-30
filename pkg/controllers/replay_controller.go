package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kubeshark/hub/pkg/replay"
	"github.com/kubeshark/hub/pkg/validation"
	"github.com/rs/zerolog/log"
)

const (
	replayTimeout = 10 * time.Second
)

func ReplayRequest(c *gin.Context) {
	log.Debug().Msg("Starting replay")
	replayDetails := &replay.Details{}
	if err := c.Bind(replayDetails); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	log.Debug().Interface("replay-details", replayDetails).Msg("Validating replay...")
	if err := validation.Validate(replayDetails); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	log.Debug().Interface("replay-details", replayDetails).Msg("Executing replay...")
	result := replay.ExecuteRequest(replayDetails, replayTimeout)
	c.JSON(http.StatusOK, result)
}
