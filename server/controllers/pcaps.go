package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

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

func GetDownloadPcap(c *gin.Context) {
	workerHost := c.Param("worker")
	id := c.Param("id")

	dir, err := os.MkdirTemp(misc.GetDataDir(), "singlecap")
	if err != nil {
		log.Error().Err(err).Msg("Failed to create temp directory!")
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	defer os.RemoveAll(dir)

	client := &http.Client{}

	err = misc.FetchPcapFile(client, dir, workerHost, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	err = misc.FetchNameResolutionHistory(client, dir, workerHost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	tarName, tarPath, err := misc.TarIt(dir)
	if err != nil {
		log.Error().Str("dir", dir).Err(err).Msg("Couldn't TAR the directory!")
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	defer os.Remove(tarPath)

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+tarName)
	c.Header("Content-Type", "application/octet-stream")
	c.File(tarPath)
}

type postMergeRequest struct {
	Query string              `json:"query"`
	Pcaps map[string][]string `json:"pcaps"`
}

func PostMerge(c *gin.Context) {
	var req postMergeRequest
	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	dir, err := os.MkdirTemp(misc.GetDataDir(), "tarcap")
	if err != nil {
		log.Error().Err(err).Msg("Failed to create temp directory!")
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	defer os.RemoveAll(dir)

	for workerHost, pcaps := range req.Pcaps {
		client := &http.Client{}

		node := fmt.Sprintf("%s_%s", worker.GetHostName(workerHost), misc.RemovePortFromWorkerHost(workerHost))
		subdir := filepath.Join(dir, node)
		err := os.Mkdir(subdir, os.ModePerm)
		if err != nil {
			log.Error().Err(err).Str("subdir", subdir).Msg("Failed to create subdirectory!")
			c.JSON(http.StatusInternalServerError, err)
			return
		}

		err = misc.FetchMergedPcapFile(client, subdir, req.Query, pcaps, workerHost)
		if err != nil {
			continue
		}

		err = misc.FetchNameResolutionHistory(client, subdir, workerHost)
		if err != nil {
			continue
		}
	}

	tarName, tarPath, err := misc.TarIt(dir)
	if err != nil {
		log.Error().Str("dir", dir).Err(err).Msg("Couldn't TAR the directory!")
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	defer os.Remove(tarPath)

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+tarName)
	c.Header("Content-Type", "application/octet-stream")
	c.File(tarPath)
}

func GetReplay(c *gin.Context) {
	workerHost := c.Param("worker")
	id := c.Param("id")

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/pcaps/replay/%s", workerHost, id), nil)
	if err != nil {
		log.Error().Err(err).Str("pcap", id).Msg("Worker replay PCAP build request:")
	}

	q := req.URL.Query()
	q.Add("count", c.Query("count"))
	q.Add("delay", c.Query("delay"))
	q.Add("host", c.Query("host"))
	q.Add("port", c.Query("port"))

	req.URL.RawQuery = q.Encode()

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

func PostStorageLimit(c *gin.Context) {
	var req worker.PostStorageLimit
	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	err := worker.PostStorageLimitToWorkers(req)
	if err != nil {
		log.Error().Err(err).Msg("Workers set storage limit request:")
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"limit": req.Limit,
	})
}
