package app

import (
	"fmt"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	_ "github.com/SkycoinPro/skywire-services-uptime/docs" // Needed for swagger doc

	"github.com/SkycoinPro/skywire-services-uptime/src/api"

	"github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

type Server struct {
	Engine *gin.Engine
}

func NewServer(ctrls ...api.Controller) *Server {
	if viper.GetBool("server.release-mode") {
		gin.SetMode(gin.ReleaseMode)
	}

	server := &Server{
		Engine: gin.Default(),
	}
	server.initCors()
	server.initRoutes(ctrls...)
	return server
}

func (s *Server) Run() {
	s.Engine.Run(serverAddress())
}

func (s *Server) initCors() {
	s.Engine.Use(cors.New(cors.Config{
		AllowHeaders:    viper.GetStringSlice("c0rs.allowed-headers"),
		AllowMethods:    viper.GetStringSlice("c0rs.allowed-methods"),
		AllowAllOrigins: true,
		MaxAge:          viper.GetDuration("c0rs.max-age"),
	}))
}

func (s *Server) initRoutes(ctrls ...api.Controller) {
	publicAPIGroup := s.Engine.Group("/api/v1")
	closedAPIGroup := publicAPIGroup.Group("")

	// use ginSwagger middleware to
	publicAPIGroup.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	for _, controller := range ctrls {
		controller.RegisterAPIs(publicAPIGroup, closedAPIGroup)
	}
}

func serverAddress() string {
	return fmt.Sprintf("%s:%s", viper.GetString("server.ip"), viper.GetString("server.port"))
}
