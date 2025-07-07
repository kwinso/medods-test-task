package internal

import (
	"fmt"
	"github.com/kwinso/medods-test-task/internal/handlers/middleware"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/kwinso/medods-test-task/docs"
	"github.com/kwinso/medods-test-task/internal/api"
	"github.com/kwinso/medods-test-task/internal/config"
	"github.com/kwinso/medods-test-task/internal/db"
	"github.com/kwinso/medods-test-task/internal/db/repositories"
	"github.com/kwinso/medods-test-task/internal/handlers"
	"github.com/kwinso/medods-test-task/internal/services"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func newRouter(cfg config.Config, db db.DBTX, logger *log.Logger) *gin.Engine {
	router := gin.Default()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		api.RegisterCustomValidators(v)
	}

	authRepo := repositories.NewPgxAuthRepository(db)

	reportsService := services.NewWebhookReportsService(cfg.WebhookURL)

	authService := services.NewAuthService(authRepo, &reportsService, logger, cfg.JwtKey, cfg.TokenTTL, cfg.AuthTTL)
	authHandler := handlers.NewAuthHandler(cfg, authService, logger)

	authMiddleware := middleware.NewAuthMiddleware(authService, logger)
	authHandler.SetupRoutes(router, authMiddleware)

	docs.SwaggerInfo.BasePath = "/"
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	return router
}

// ServeWithConfig bootstraps and app using the app config and db connection
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				Authorization header using the Bearer scheme
func ServeWithConfig(cfg config.Config, db db.DBTX, logger *log.Logger) error {
	return http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), newRouter(cfg, db, logger))
}
