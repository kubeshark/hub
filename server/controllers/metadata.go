package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kubeshark/base/pkg/models"
	"github.com/kubeshark/hub/pkg/version"
)

func GetVersion(c *gin.Context) {
	resp := models.VersionResponse{Ver: version.Ver}
	c.JSON(http.StatusOK, resp)
}
