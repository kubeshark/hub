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
	workerHost := c.Param("worker")
	id := c.Param("id")
	query := c.Query("q")
	field := c.Query("field")

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/item/%s", workerHost, id), nil)
	if err != nil {
		log.Error().Err(err).Str("pcap", id).Msg("Worker fetch item build request:")
	}

	q := req.URL.Query()
	q.Add("q", query)
	q.Add("worker", workerHost)

	req.URL.RawQuery = q.Encode()

	res, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Str("pcap", id).Msg("Worker fetch item do request:")
		handleError(c, err)
		return
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Error().Err(err).Str("pcap", id).Msg("Worker fetch item body read:")
		handleError(c, err)
		return
	}

	if res.StatusCode != 200 {
		c.JSON(res.StatusCode, body)
		return
	}

	var payload map[string]interface{}
	err = json.Unmarshal(body, &payload)
	if err != nil {
		log.Error().Err(err).Str("pcap", id).Msg("Worker fetch item JSON unmarshal:")
		handleError(c, err)
		return
	}

	var data interface{}
	var ok bool
	if field == "" {
		data = payload
	} else {
		data, ok = payload[field]
		if !ok {
			c.JSON(http.StatusBadRequest, payload)
			return
		}
	}

	c.JSON(http.StatusOK, data)
}
