package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func writeBindError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
}
