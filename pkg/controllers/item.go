package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func handleError(c *gin.Context, err error) {
	_ = c.Error(err)
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"msg": err.Error(),
	})
}

func GetItem(c *gin.Context) {
	id := c.Param("id")

	res, err := http.Get(fmt.Sprintf("http://localhost:8897/item/%s", id))
	if err != nil {
		log.Error().Err(err).Str("pcap", id).Msg("Worker fetch item:")
		handleError(c, err)
		return
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Error().Err(err).Str("pcap", id).Msg("Worker fetch item body read:")
		handleError(c, err)
		return
	}

	var payload map[string]interface{}
	err = json.Unmarshal(body, &payload)
	if err != nil {
		log.Error().Err(err).Str("pcap", id).Msg("Worker fetch item JSON unmarshal:")
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, payload)
}
