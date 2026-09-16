package rest_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/Nutto555/hospital-middleware/internal/rest"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

func TestHealthz(t *testing.T) {
	r := rest.NewRouter(rest.Deps{})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := rec.Body.String(); got != `{"status":"ok"}` {
		t.Fatalf("body = %s", got)
	}
}

func TestWrongMethodIsRejectedNotUnknown(t *testing.T) {
	r := rest.NewRouter(rest.Deps{})
	for _, path := range []string{"/staff/create", "/staff/login", "/patient/search"} {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("GET %s: status = %d, want 405", path, rec.Code)
		}
	}
}
