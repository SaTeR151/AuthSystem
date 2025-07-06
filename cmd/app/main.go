package main

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	docs "github.com/sater-151/AuthSystem/docs"
	"github.com/sater-151/AuthSystem/internal/config"
	"github.com/sater-151/AuthSystem/internal/controller/rest"
	"github.com/sater-151/AuthSystem/internal/controller/rest/middleware"
	"github.com/sater-151/AuthSystem/internal/database/postgresql"
	"github.com/sater-151/AuthSystem/internal/pkg/webhooks"
	authsystem "github.com/sater-151/AuthSystem/internal/services/authSystem"
	"github.com/sirupsen/logrus"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

//	@title			AuthSystem
//	@version		0.9.0
//	@description	Сервис для авторизации, обновления и деавторизации пользователя, а также получение его guid

// @host		localhost:8080
// @BasePath	/api

// @securitydefinitions.apikey AccessToken
// @in header
// @name at

// @securitydefinitions.apikey RefreshToken
// @in header
// @name rt
func main() {
	if err := godotenv.Load(); err != nil {
		logrus.Error(err)
		return
	}

	config.InitLoggerConfig()

	psqlConfig := config.GetPostresqlConfig()
	db, close, err := postgresql.Open(psqlConfig)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer close()
	if err := db.MigrationUp(); err != nil {
		logrus.Error(err)
		return
	}
	serverConfig := config.GetServerConfig()

	wh := webhooks.NewClient()
	authsystem := authsystem.New(db, wh)

	router := gin.Default()

	docs.SwaggerInfo.BasePath = "/"
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// Для удобства доступа к swagger
	router.GET("/swagger", func(c *gin.Context) {
		location := url.URL{Path: "/swagger/index.html"}
		c.Redirect(http.StatusFound, location.RequestURI())
	})

	router.POST("/api/login", rest.Login(authsystem))
	router.POST("/api/refresh", rest.Refresh(authsystem))
	{
		authGroup := router.Group("/api/auth", middleware.CheckAuthorization(authsystem))
		authGroup.POST("/logout", rest.Deauthorization(authsystem))
		authGroup.GET("/guid", rest.GetGUID(authsystem))
	}

	if err := router.Run(":" + serverConfig.Port); err != nil {
		logrus.Error(err)
		return
	}
}
