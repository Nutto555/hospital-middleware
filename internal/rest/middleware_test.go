package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Nutto555/hospital-middleware/internal/auth"
)

func TestRequireAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokens := auth.NewJWT("secret", time.Hour)
	good, _, err := tokens.Issue(auth.Principal{StaffID: "staff-1", HospitalCode: "hospital-a"})
	if err != nil {
		t.Fatal(err)
	}
	foreign, _, err := auth.NewJWT("other", time.Hour).Issue(auth.Principal{StaffID: "staff-1", HospitalCode: "hospital-a"})
	if err != nil {
		t.Fatal(err)
	}

	r := gin.New()
	r.GET("/protected", requireAuth(tokens), func(c *gin.Context) {
		p, ok := principalFrom(c)
		if !ok {
			c.String(http.StatusInternalServerError, "no principal")
			return
		}
		c.String(http.StatusOK, p.StaffID+"@"+p.HospitalCode)
	})

	cases := []struct {
		name   string
		header string
		status int
		body   string
	}{
		{"valid", "Bearer " + good, 200, "staff-1@hospital-a"},
		{"missing", "", 401, `{"error":"invalid or missing token"}`},
		{"wrong scheme", "Basic " + good, 401, `{"error":"invalid or missing token"}`},
		{"empty bearer", "Bearer ", 401, `{"error":"invalid or missing token"}`},
		{"garbage", "Bearer nope", 401, `{"error":"invalid or missing token"}`},
		{"other secret", "Bearer " + foreign, 401, `{"error":"invalid or missing token"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			if rec.Code != tc.status || rec.Body.String() != tc.body {
				t.Fatalf("got %d %s, want %d %s", rec.Code, rec.Body, tc.status, tc.body)
			}
		})
	}
}

func TestPrincipalFromWithoutMiddleware(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	if _, ok := principalFrom(c); ok {
		t.Fatal("no principal must be reported as absent")
	}
}
