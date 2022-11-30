package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kubeshark/base/pkg/models"
	"github.com/kubeshark/hub/pkg/api"
	"github.com/kubeshark/hub/pkg/holder"
	"github.com/kubeshark/hub/pkg/kubernetes"
	"github.com/kubeshark/hub/pkg/providers"
	"github.com/kubeshark/hub/pkg/providers/targettedPods"
	"github.com/kubeshark/hub/pkg/providers/workers"
	"github.com/kubeshark/hub/pkg/validation"
	"github.com/rs/zerolog/log"
	core "k8s.io/api/core/v1"
)

func HealthCheck(c *gin.Context) {
	workersStatus := make([]*models.WorkerStatus, 0)
	for _, value := range workers.GetStatus() {
		workersStatus = append(workersStatus, value)
	}

	response := models.HealthResponse{
		TargettedPods:         targettedPods.Get(),
		ConnectedWorkersCount: workers.GetConnectedCount(),
		WorkersStatus:         workersStatus,
	}
	c.JSON(http.StatusOK, response)
}

func PostTargettedPods(c *gin.Context) {
	var requestTargettedPods []core.Pod
	if err := c.Bind(&requestTargettedPods); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	podInfos := kubernetes.GetPodInfosForPods(requestTargettedPods)

	log.Info().Int("targetted-pods-count", len(requestTargettedPods)).Msg("POST request:")
	targettedPods.Set(podInfos)
	api.BroadcastTargettedPodsStatus()

	nodeToTargettedPodMap := kubernetes.GetNodeHostToTargettedPodsMap(requestTargettedPods)
	targettedPods.SetNodeToTargettedPodMap(nodeToTargettedPodMap)
	api.BroadcastTargettedPodsToWorkers(nodeToTargettedPodMap)
}

func PostWorkerStatus(c *gin.Context) {
	workerStatus := &models.WorkerStatus{}
	if err := c.Bind(workerStatus); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	if err := validation.Validate(workerStatus); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	log.Info().Interface("worker-status", workerStatus).Msg("POST request:")
	workers.SetStatus(workerStatus)
	api.BroadcastTargettedPodsStatus()
}

func GetConnectedWorkersCount(c *gin.Context) {
	c.JSON(http.StatusOK, workers.GetConnectedCount())
}

func GetTargettingStatus(c *gin.Context) {
	targettedPodsStatus := targettedPods.GetTargettedPodsStatus()
	c.JSON(http.StatusOK, targettedPodsStatus)
}

func GetGeneralStats(c *gin.Context) {
	c.JSON(http.StatusOK, providers.GetGeneralStats())
}

func GetTrafficStats(c *gin.Context) {
	startTime, endTime, err := getStartEndTime(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, providers.GetTrafficStats(startTime, endTime))
}

func getStartEndTime(c *gin.Context) (time.Time, time.Time, error) {
	startTimeValue, err := strconv.Atoi(c.Query("startTimeMs"))
	if err != nil {
		return time.UnixMilli(0), time.UnixMilli(0), fmt.Errorf("invalid start time: %v", err)
	}
	endTimeValue, err := strconv.Atoi(c.Query("endTimeMs"))
	if err != nil {
		return time.UnixMilli(0), time.UnixMilli(0), fmt.Errorf("invalid end time: %v", err)
	}
	return time.UnixMilli(int64(startTimeValue)), time.UnixMilli(int64(endTimeValue)), nil
}

func GetCurrentResolvingInformation(c *gin.Context) {
	c.JSON(http.StatusOK, holder.GetResolver().GetMap())
}
