package rest

import "github.com/gin-gonic/gin"

// writeError ends the request with a JSON error body.
func writeError(c *gin.Context, status int, msg string) {
	c.AbortWithStatusJSON(status, gin.H{"error": msg})
}
