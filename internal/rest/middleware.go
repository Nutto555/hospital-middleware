package rest

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Nutto555/hospital-middleware/internal/auth"
)

// TokenVerifier turns a bearer token into the principal it was issued to.
type TokenVerifier interface {
	Verify(token string) (auth.Principal, error)
}

const principalKey = "principal"

// requireAuth rejects requests without a valid bearer token and stores the principal for handlers.
func requireAuth(v TokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")
		if !ok || strings.TrimSpace(token) == "" {
			writeError(c, http.StatusUnauthorized, auth.ErrInvalidToken.Error())
			return
		}
		p, err := v.Verify(strings.TrimSpace(token))
		if err != nil {
			writeError(c, http.StatusUnauthorized, auth.ErrInvalidToken.Error())
			return
		}
		c.Set(principalKey, p)
		c.Next()
	}
}

// principalFrom returns the principal stored by requireAuth.
func principalFrom(c *gin.Context) (auth.Principal, bool) {
	v, ok := c.Get(principalKey)
	if !ok {
		return auth.Principal{}, false
	}
	p, ok := v.(auth.Principal)
	return p, ok
}
