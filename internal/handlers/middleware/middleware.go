package middleware

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/kwinso/medods-test-task/internal/api"
	"github.com/kwinso/medods-test-task/internal/services"
	"log"
	"net/http"
	"strings"
)

type Middleware interface {
	Handle(gin *gin.Context)
}

// AuthMiddleware parses a bearer JWT token from the request and checks if there's a valid session present for the ID.
// If it fails to do so, it aborts the connection with 401 error.
//
// If the token is parsed successfully, it will set following context values:
//   - `user_guid` - GUID of the authorized user
//   - `auth_id` - id of the auth session
type AuthMiddleware struct {
	authService services.AuthService
	logger      *log.Logger
}

// NewAuthMiddleware creates new AuthMiddleware
func NewAuthMiddleware(authService services.AuthService, logger *log.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		logger:      logger,
	}
}

func (m *AuthMiddleware) Handle(c *gin.Context) {
	bearerToken := c.GetHeader("Authorization")
	if bearerToken == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, api.UnauthorizedResponse)
		return
	}

	token := strings.TrimPrefix(bearerToken, "Bearer ")
	if token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, api.UnauthorizedResponse)
		return
	}

	auth, err := m.authService.GetAuthByAccessToken(c.Request.Context(), token)
	if err != nil {
		if errors.Is(err, services.ErrAuthExpired) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, api.UnauthorizedResponse)
			return
		}

		m.logger.Printf("Failed to authorize user: %v\n", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, api.InternalServerErrorResponse)
		return
	}

	c.Set("user_guid", auth.Guid)
	c.Set("auth_id", auth.ID)

	c.Next()
}
