package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kubeshark/hub/pkg/worker"
	v1 "k8s.io/api/core/v1"
)

func PostWorker(c *gin.Context) {
	var pod v1.Pod
	if err := c.Bind(&pod); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	if !worker.RemovedDefaultHost {
		worker.RemoveHost(worker.HostWithPort(worker.DefaultWorkerHost))
		worker.RemovedDefaultHost = true
	}

	var msg string
	host := worker.HostWithPort(pod.Status.PodIP)
	if host == "" {
		msg = "Pod IP is empty. Did nothing."
		c.JSON(http.StatusOK, gin.H{
			"msg":  msg,
			"host": host,
		})
		return
	}

	if pod.Status.Phase == v1.PodRunning && pod.Status.ContainerStatuses[0].Ready {
		worker.AddHost(host)
		msg = "Added a new worker host."
	} else {
		worker.RemoveHost(host)
		msg = "Removed a worker host."
	}

	err := worker.PostTargettedPodsToWorkers()
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"msg":  msg,
		"host": host,
	})
}

type Target struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

func GetTargetted(c *gin.Context) {
	targets := make([]Target, 0)
	pods := worker.GetTarggetedPods()
	for _, pod := range pods {
		targets = append(targets, Target{
			Name:      pod.GetObjectMeta().GetName(),
			Namespace: pod.GetNamespace(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"targets": targets,
	})
}

func PostTargetted(c *gin.Context) {
	var pods []v1.Pod
	if err := c.Bind(&pods); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	worker.SetTargettedPods(pods)
	err := worker.PostTargettedPodsToWorkers()
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pods": len(pods),
	})
}
