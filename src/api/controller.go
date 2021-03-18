package api

import (
	"github.com/gin-gonic/gin"
)

type Controller interface {
	RegisterAPIs(public *gin.RouterGroup, closed *gin.RouterGroup)
}

type ErrorResponse struct {
	Error string `json:"message"`
}
