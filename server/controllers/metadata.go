package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kubeshark/hub/pkg/version"
)

type VersionResponse struct {
	Ver string `json:"ver"`
}

func GetVersion(c *gin.Context) {
	resp := VersionResponse{Ver: version.Ver}
	c.JSON(http.StatusOK, resp)
}
