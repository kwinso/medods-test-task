package handlers

import (
	"errors"
	"log"
	"net/http"
	"net/netip"
	"strings"

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

	authorized := router.Group("/")
	authorized.Use(h.AuthorizeMiddleware)
	{
		authorized.GET("/me", h.GetMe)
		authorized.DELETE("/logout", h.Logout)
	}
}

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
		RefreshToken: tokenPair.RefreshToken,
	})
}

func (h *AuthHandler) AuthorizeMiddleware(c *gin.Context) {
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

	auth, err := h.authService.GetAuthByToken(c.Request.Context(), token)
	if err != nil {
		if errors.Is(err, services.ErrAuthExpired) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, api.UnauthorizedResponse)
			return
		}

		h.logger.Printf("Failed to authorize user: %v\n", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, api.InternalServerErrorResponse)
		return
	}

	c.Set("user_guid", auth.Guid)
	// Note: This conversion is needed becase for some reason get
	c.Set("auth_id", auth.ID)

	c.Next()
}

// @Summary			Get the GUID for the authenticated user
// @Description	Returns the GUID for the authenticated user
// @Security		BearerAuth
// @Produce			json
// @Success			200	{object}	api.GetMeResposne
// @Failure			401	{object}	api.ErrorResponse	"Unauthorized"
// @Failure			500 {object}	api.ErrorResponse	"Internal Server Error"
// @Router			/me [get]
func (h *AuthHandler) GetMe(c *gin.Context) {
	guid := c.GetString("user_guid")

	c.JSON(http.StatusOK, api.GetMeResposne{
		Guid: guid,
	})
}

// @Summary			Logout the authenticated user
// @Description	Deletes the auth for the authenticated user
// @Security		BearerAuth
// @Success			204 "Successfully logged out"
// @Failure			401	{object}	api.ErrorResponse	"Unauthorized"
// @Failure			500 {object}	api.ErrorResponse	"Internal Server Error"
// @Router			/logout [delete]
func (h *AuthHandler) Logout(c *gin.Context) {
	authId, _ := c.MustGet("auth_id").(int32)

	err := h.authService.DeleteAuthById(c.Request.Context(), authId)
	if err != nil {
		h.logger.Printf("Failed to delete auth id %d: %v\n", authId, err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, api.InternalServerErrorResponse)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
