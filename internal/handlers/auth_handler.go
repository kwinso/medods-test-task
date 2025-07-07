package handlers

import (
	"encoding/base64"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kwinso/medods-test-task/internal/api"
	"github.com/kwinso/medods-test-task/internal/config"
	"github.com/kwinso/medods-test-task/internal/handlers/middleware"
	"github.com/kwinso/medods-test-task/internal/services"
	"github.com/kwinso/medods-test-task/internal/tokens"
	"log"
	"net/http"
	"net/netip"
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

func (h *AuthHandler) SetupRoutes(router *gin.Engine, auth middleware.Middleware) {
	router.POST("/login", h.Login)
	router.PUT("/refresh", h.RefreshTokens)

	authorized := router.Group("/")
	authorized.Use(auth.Handle)
	{
		authorized.GET("/me", h.GetMe)
		authorized.DELETE("/logout", h.Logout)
	}
}

// Login handles generating a pair of tokens for a requested GUID
// @Summary	Generate a token pair from guid
// @Param		request	body	api.LoginRequest	true	"login request"
// @Accept		json
// @Produce	json
// @Success	200	{object}	api.TokenPair
// @Failure	400	{object}	api.ErrorResponse	"Bad Request"
// @Failure	500 {object}	api.ErrorResponse	"Internal Server Error"
// @Router		/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req api.LoginRequest
	if err := c.ShouldBind(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, api.ErrorResponse{Error: err.Error()})
		return
	}

	ua := c.Request.UserAgent()
	ipString := c.ClientIP()
	inet, err := netip.ParseAddr(ipString)
	if err != nil {
		h.logger.Printf("Failed to parse IP address: %v\n", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, api.InternalServerErrorResponse)
		return
	}

	tokenPair, err := h.authService.AuthorizeByGUID(c.Request.Context(), req.GUID, ua, inet)

	if err != nil {
		h.logger.Printf("Failed to authorize user: %v\n", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, api.InternalServerErrorResponse)
		return
	}

	c.JSON(http.StatusOK, api.TokenPair{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokens.EncodeRefreshTokenToBase64(tokenPair.RefreshToken),
	})
}

// GetMe handles getting authorized user GUID
// @Summary			Get the GUID for the authenticated user
// @Description	Returns the GUID for the authenticated user
// @Security		BearerAuth
// @Produce			json
// @Success			200	{object}	api.GetMeResponse
// @Failure			401	{object}	api.ErrorResponse	"Unauthorized"
// @Failure			500 {object}	api.ErrorResponse	"Internal Server Error"
// @Router			/me [get]
func (h *AuthHandler) GetMe(c *gin.Context) {
	guid := c.GetString("user_guid")

	c.JSON(http.StatusOK, api.GetMeResponse{
		Guid: guid,
	})
}

// RefreshTokens is a route for handling tokens refresh
// @Summary			Refresh the access token for the authenticated user
// @Description	Refresh the access token for the authenticated user
// @Param			request	body	api.RefreshRequest	true	"refresh request"
// @Accept			json
// @Produce			json
// @Success			200	{object}	api.TokenPair
// @Failure			400	{object}	api.ErrorResponse	"Bad Request"
// @Failure			401	{object}	api.ErrorResponse	"Unauthorized"
// @Failure			500 {object}	api.ErrorResponse	"Internal Server Error"
// @Router			/refresh [put]
func (h *AuthHandler) RefreshTokens(c *gin.Context) {
	var req api.RefreshRequest
	if err := c.ShouldBind(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, api.ErrorResponse{Error: err.Error()})
		return
	}

	ipString := c.ClientIP()
	inet, err := netip.ParseAddr(ipString)
	if err != nil {
		h.logger.Printf("Failed to parse IP address: %v\n", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, api.InternalServerErrorResponse)
		return
	}

	// decode token from base64
	token, err := base64.StdEncoding.DecodeString(req.RefreshToken)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, api.BadRequestResponse)
		return
	}

	tokenPair, err := h.authService.RefreshAuth(c.Request.Context(), string(token), c.GetHeader("User-Agent"), inet)
	if err != nil {
		if errors.Is(err, services.ErrUserAgentMismatch) ||
			errors.Is(err, services.ErrInvalidTokenFormat) ||
			errors.Is(err, services.ErrAuthExpired) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, api.UnauthorizedResponse)
		} else {
			h.logger.Printf("Failed to refresh auth: %v\n", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, api.InternalServerErrorResponse)
		}

		return
	}

	c.JSON(http.StatusOK, api.TokenPair{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokens.EncodeRefreshTokenToBase64(tokenPair.RefreshToken),
	})
}

// Logout handles logout logic
// @Summary			Logout the authenticated user
// @Description	Deletes the auth for the authenticated user
// @Security		BearerAuth
// @Success			204 "Successfully logged out"
// @Failure			401	{object}	api.ErrorResponse	"Unauthorized"
// @Failure			500 {object}	api.ErrorResponse	"Internal Server Error"
// @Router			/logout [delete]
func (h *AuthHandler) Logout(c *gin.Context) {
	authId := c.MustGet("auth_id").(uuid.UUID)

	err := h.authService.DeleteAuthById(c.Request.Context(), authId)
	if err != nil {
		h.logger.Printf("Failed to delete auth id %d: %v\n", authId, err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, api.InternalServerErrorResponse)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
