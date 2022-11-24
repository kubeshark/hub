package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kubeshark/hub/pkg/version"
	"github.com/kubeshark/worker/models"
)

func GetVersion(c *gin.Context) {
	resp := models.VersionResponse{Ver: version.Ver}
	c.JSON(http.StatusOK, resp)
}
