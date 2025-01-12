package main

import (
	"net/http"
	"strings"

	"github.com/nsqio/go-nsq"
	"github.com/yyboo586/IAMService/dbaccess"
	"github.com/yyboo586/IAMService/drivenadapters"
	"github.com/yyboo586/IAMService/driveradapters"
	"github.com/yyboo586/IAMService/interfaces"
	"github.com/yyboo586/IAMService/logics"

	"github.com/go-mail/mail/v2"

	configUtils "github.com/yyboo586/IAMService/utils/config"

	"github.com/yyboo586/common/dbUtils"
	"github.com/yyboo586/common/logUtils"

	"github.com/gin-gonic/gin"
)

type Server struct {
	userHandler interfaces.RESTHandler
	oidcHandler interfaces.RESTHandler
	config      *configUtils.Config
}

var (
	// allowOrigins = []string{"http://127.0.0.1:5000", "http://127.0.0.1:5001", "http://127.0.0.1:5501"}
	allowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	allowHeaders = []string{"Content-Type", "Authorization"}
)

func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}

		c.Writer.Header().Set("Access-Control-Allow-Methods", strings.Join(allowMethods, ", "))
		c.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(allowHeaders, ", "))
		c.Writer.Header().Set("Access-Control-Max-Age", "3600")
		if c.Request.Method == "OPTIONS" {
			c.Writer.WriteHeader(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func (s *Server) Start() {
	go func() {
		gin.SetMode(gin.ReleaseMode)
		engine := gin.Default()
		engine.Use(cors())

		s.userHandler.RegisterPublic(engine)
		s.oidcHandler.RegisterPublic(engine)

		if err := engine.Run(s.config.Server.Addr); err != nil {
			panic(err)
		}
	}()
}

func main() {
	config := configUtils.Default()
	dbPool, err := dbUtils.NewDB(&config.DBConfig)
	if err != nil {
		panic(err)
	}
	logger, err := logUtils.NewLogger(config.Logger.Level)
	if err != nil {
		panic(err)
	}
	mailDialer := mail.NewDialer(config.Mailer.Host, config.Mailer.Port, config.Mailer.User, config.Mailer.Pass)

	producer, err := nsq.NewProducer("127.0.0.1:4150", nsq.NewConfig())
	if err != nil {
		panic(err)
	}

	// 依赖注入
	dbaccess.SetDBPool(dbPool)
	dbaccess.SetLogger(logger)

	drivenadapters.SetMailDialer(mailDialer)
	drivenadapters.SetMQProducer(producer)
	drivenadapters.SetLogger(logger)

	logics.SetDB(dbPool)
	logics.SetLogger(logger)

	driveradapters.SetLogger(logger)

	s := &Server{
		userHandler: driveradapters.NewUserHandler(),
		oidcHandler: driveradapters.NewOIDCHandler(),
		config:      config,
	}

	s.Start()

	select {}
}
