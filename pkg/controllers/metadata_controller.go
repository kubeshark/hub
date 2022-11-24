package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kubeshark/hub/pkg/version"
	"github.com/kubeshark/kubeshark/shared"
)

func GetVersion(c *gin.Context) {
	resp := shared.VersionResponse{Ver: version.Ver}
	c.JSON(http.StatusOK, resp)
}
