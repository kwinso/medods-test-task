package handlers

import (
	"log"
	"net/http"
	"net/netip"

	"github.com/gin-gonic/gin"
	"github.com/kwinso/medods-test-task/internal/api"
	"github.com/kwinso/medods-test-task/internal/config"
	"github.com/kwinso/medods-test-task/internal/services"
)

type AuthHandler struct {
	Config      config.Config
	authService services.AuthService
	logger      *log.Logger
}

func NewAuthHandler(cfg config.Config, authService services.AuthService, logger *log.Logger) AuthHandler {
	return AuthHandler{
		Config:      cfg,
		authService: authService,
		logger:      logger,
	}
}

func (h *AuthHandler) SetupRoutes(router *gin.Engine) {
	router.POST("/login", h.Login)
}

// @Summary	Generate a token pair from guid
// @Schemes
// @Description	do ping
// @Param			request body api.LoginRequest true "login request"
// @Accept			json
// @Produce		json
// @Success		200	{object}	api.TokenPair
// @Router			/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req api.LoginRequest
	if err := c.ShouldBind(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ua := c.Request.UserAgent()
	ipString := c.ClientIP()
	inet, err := netip.ParseAddr(ipString)
	if err != nil {
		h.logger.Printf("Failed to parse IP address: %v\n", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	tokenPair, err := h.authService.AuthorizeByGUID(c.Request.Context(), req.GUID, ua, inet)

	if err != nil {
		h.logger.Printf("Failed to authorize user: %v\n", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, api.TokenPair{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	})
}
