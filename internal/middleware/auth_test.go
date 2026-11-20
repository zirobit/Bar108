package middleware

import (
	"bar108/config"
	"bar108/internal/db"
	jwtpkg "bar108/internal/jwt"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type testBlacklist struct{}

func (testBlacklist) IsTokenBlacklisted(context.Context, string) (bool, error) {
	return false, nil
}

type testUserStore struct{ user db.User }

func (s testUserStore) GetUserByID(context.Context, int32) (db.User, error) {
	return s.user, nil
}

func TestAuthUsesCurrentAccountState(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := jwtpkg.New(config.JWTConfig{Secret: "a-secret-for-tests", ExpiryHours: 1})
	token, err := manager.Generate(1, "admin")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name       string
		user       db.User
		wantStatus int
	}{
		{"deactivated", db.User{ID: 1, Role: "admin", IsActive: false}, http.StatusUnauthorized},
		{"role downgraded", db.User{ID: 1, Role: "customer", IsActive: true}, http.StatusForbidden},
		{"active admin", db.User{ID: 1, Role: "admin", IsActive: true}, http.StatusOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			r.Use(AuthMiddleware(manager, testBlacklist{}, testUserStore{tc.user}), AdminMiddleware())
			r.GET("/admin", func(c *gin.Context) { c.Status(http.StatusOK) })
			req := httptest.NewRequest(http.MethodGet, "/admin", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tc.wantStatus {
				t.Fatalf("got %d, want %d", w.Code, tc.wantStatus)
			}
		})
	}
}
