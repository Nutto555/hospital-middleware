package rest

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"

	"github.com/Nutto555/hospital-middleware/internal/domain"
	"github.com/Nutto555/hospital-middleware/internal/service"
)

// StaffService is what the staff endpoints need from the service layer.
type StaffService interface {
	Create(ctx context.Context, username, password, hospital string) (domain.Staff, error)
	Login(ctx context.Context, username, password, hospital string) (service.Session, error)
}

var usernameRe = regexp.MustCompile(`^[A-Za-z0-9._-]{3,64}$`)

type credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Hospital string `json:"hospital"`
}

// validateForCreate applies the account rules; login only needs the fields to be present.
func (c credentials) validateForCreate() error {
	switch {
	case !usernameRe.MatchString(c.Username):
		return errors.New("username must be 3-64 characters of letters, digits, '.', '_' or '-'")
	case utf8.RuneCountInString(c.Password) < 8:
		return errors.New("password must be at least 8 characters")
	case strings.TrimSpace(c.Hospital) == "":
		return errors.New("hospital is required")
	}
	return nil
}

func (c credentials) validateForLogin() error {
	switch {
	case c.Username == "":
		return errors.New("username is required")
	case c.Password == "":
		return errors.New("password is required")
	case strings.TrimSpace(c.Hospital) == "":
		return errors.New("hospital is required")
	}
	return nil
}

type staffResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Hospital string `json:"hospital"`
}

type loginResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}

func createStaff(svc StaffService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var in credentials
		if err := c.ShouldBindJSON(&in); err != nil {
			writeError(c, http.StatusBadRequest, "invalid request body")
			return
		}
		if err := in.validateForCreate(); err != nil {
			writeError(c, http.StatusBadRequest, err.Error())
			return
		}
		st, err := svc.Create(c.Request.Context(), in.Username, in.Password, strings.TrimSpace(in.Hospital))
		if err != nil {
			writeServiceError(c, err)
			return
		}
		c.JSON(http.StatusCreated, staffResponse{ID: st.ID, Username: st.Username, Hospital: st.HospitalCode})
	}
}

func login(svc StaffService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var in credentials
		if err := c.ShouldBindJSON(&in); err != nil {
			writeError(c, http.StatusBadRequest, "invalid request body")
			return
		}
		if err := in.validateForLogin(); err != nil {
			writeError(c, http.StatusBadRequest, err.Error())
			return
		}
		sess, err := svc.Login(c.Request.Context(), in.Username, in.Password, strings.TrimSpace(in.Hospital))
		if err != nil {
			writeServiceError(c, err)
			return
		}
		c.JSON(http.StatusOK, loginResponse{Token: sess.Token, ExpiresAt: sess.ExpiresAt.UTC().Format(time.RFC3339)})
	}
}
