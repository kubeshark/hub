package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kubeshark/base/pkg/languages/kfl"
)

type ValidateResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
}

func GetValidate(c *gin.Context) {
	query := c.Query("q")
	valid := true
	message := ""

	err := kfl.Validate(query)
	if err != nil {
		valid = false
		message = err.Error()
	}

	c.JSON(http.StatusOK, ValidateResponse{
		Valid:   valid,
		Message: message,
	})
}
