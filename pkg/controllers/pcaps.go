package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/kubeshark/hub/pkg/misc"
	"github.com/kubeshark/hub/pkg/worker"
	"github.com/rs/zerolog/log"
)

type totalTcpStreamsResponse struct {
	Total int64 `json:"total"`
}

func GetTotalTcpStreams(c *gin.Context) {
	var counter int64

	worker.RangeHosts(func(workerHost, v interface{}) bool {
		client := &http.Client{}
		getTotalTcpStreamsUrl := fmt.Sprintf("http://%s/pcaps/total-tcp-streams", workerHost)
		log.Debug().Str("url", getTotalTcpStreamsUrl).Msg("Doing get total TCP streams request:")
		res, err := client.Get(getTotalTcpStreamsUrl)
		if err != nil {
			log.Error().Err(err).Str("url", getTotalTcpStreamsUrl).Msg("Get total TCP streams request:")
			return true
		}

		if res.Body != nil {
			defer res.Body.Close()
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			log.Error().Err(err).Str("url", getTotalTcpStreamsUrl).Msg("Can't read body:")
			return true
		}

		resObj := totalTcpStreamsResponse{}
		err = json.Unmarshal(body, &resObj)
		if err != nil {
			log.Error().Err(err).Str("url", getTotalTcpStreamsUrl).Msg("Can't unmarshal:")
			return true
		}

		counter += resObj.Total

		return true
	})

	c.JSON(http.StatusOK, gin.H{
		"total": counter,
	})
}

func GetMerge(c *gin.Context) {
	dir, err := os.MkdirTemp(misc.GetDataDir(), "mergecap")
	if err != nil {
		log.Error().Err(err).Msg("Failed to create temp directory!")
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	defer os.RemoveAll(dir)

	worker.RangeHosts(func(workerHost, v interface{}) bool {
		client := &http.Client{}
		w := workerHost.(string)

		err := misc.FetchMergedPcapFile(client, dir, w)
		if err != nil {
			return true
		}

		err = misc.FetchNameResolutionHistory(client, dir, w)
		if err != nil {
			return true
		}

		return true
	})

	zipName, zipPath, err := misc.ZipIt(dir)
	if err != nil {
		log.Error().Str("dir", dir).Err(err).Msg("Couldn't ZIP the directory!")
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	defer os.Remove(zipPath)

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+zipName)
	c.Header("Content-Type", "application/octet-stream")
	c.File(zipPath)
}

func GetReplay(c *gin.Context) {
	workerHost := c.Param("worker")
	id := c.Param("id")

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/pcaps/replay/%s", workerHost, id), nil)
	if err != nil {
		log.Error().Err(err).Str("pcap", id).Msg("Worker replay PCAP build request:")
	}

	res, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Str("pcap", id).Msg("Worker replay PCAP do request:")
		handleError(c, err)
		return
	}

	c.JSON(res.StatusCode, gin.H{
		"status": res.Status,
	})
}