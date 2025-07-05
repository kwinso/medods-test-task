package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kwinso/medods-test-task/internal/api"
	"github.com/kwinso/medods-test-task/internal/config"
	"github.com/kwinso/medods-test-task/internal/services"
)

type AuthHandler struct {
	Config      config.Config
	authService services.AuthService
}

func NewAuthHandler(cfg config.Config, authService services.AuthService) AuthHandler {
	return AuthHandler{
		Config:      cfg,
		authService: authService,
	}
}

func (h *AuthHandler) SetupRoutes(router *gin.Engine) {
	router.POST("/login", h.Login)
}

// @BasePath /api/v1

// PingExample godoc
// @Summary ping example
// @Schemes
// @Description do ping
// @Tags example
// @Accept json
// @Produce json
// @Success 200 {string} Helloworld
// @Router /example/helloworld [get]
func (h *AuthHandler) Login(c *gin.Context) {
	var req api.LoginRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.authService.AuthorizeByGUID(req.GUID)
}
