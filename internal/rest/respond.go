package rest

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Nutto555/hospital-middleware/internal/domain"
)

// writeError ends the request with a JSON error body.
func writeError(c *gin.Context, status int, msg string) {
	c.AbortWithStatusJSON(status, gin.H{"error": msg})
}

// writeServiceError maps the domain sentinels to statuses; anything else is a logged 500.
func writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrUnknownHospital):
		writeError(c, http.StatusBadRequest, domain.ErrUnknownHospital.Error())
	case errors.Is(err, domain.ErrPasswordTooLong):
		writeError(c, http.StatusBadRequest, domain.ErrPasswordTooLong.Error())
	case errors.Is(err, domain.ErrInvalidCredentials):
		writeError(c, http.StatusUnauthorized, domain.ErrInvalidCredentials.Error())
	case errors.Is(err, domain.ErrConflict):
		writeError(c, http.StatusConflict, "username already exists in this hospital")
	case errors.Is(err, domain.ErrHISUnavailable):
		writeError(c, http.StatusBadGateway, domain.ErrHISUnavailable.Error())
	default:
		log.Printf("%s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		writeError(c, http.StatusInternalServerError, "internal error")
	}
}
