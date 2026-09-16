// Package rest exposes the HTTP API with Gin.
package rest

import "github.com/gin-gonic/gin"

// Deps are the collaborators the handlers need.
type Deps struct {
	Staff  StaffService
	Tokens TokenVerifier
}

// NewRouter builds the engine with every route and middleware registered.
func NewRouter(d Deps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.GET("/healthz", health)
	r.POST("/staff/create", createStaff(d.Staff))
	r.POST("/staff/login", login(d.Staff))
	return r
}
