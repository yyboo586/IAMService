package main

import (
	"github.com/yyboo586/IAMService/dbaccess"
	"github.com/yyboo586/IAMService/drivenadapters"
	"github.com/yyboo586/IAMService/driveradapters"
	"github.com/yyboo586/IAMService/interfaces"
	"github.com/yyboo586/IAMService/logics"

	"github.com/casbin/casbin/v2"
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

func (s *Server) Start() {
	go func() {
		engine := gin.Default()

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

	mailDialer := mail.NewDialer(config.Mailer.Host, config.Mailer.Port, config.Mailer.User, config.Mailer.Pass)

	e, err := casbin.NewEnforcer("model.conf", "policy.csv")
	if err != nil {
		panic(err)
	}

	logger, err := logUtils.NewLogger(config.Logger.Level)
	if err != nil {
		panic(err)
	}

	// 依赖注入
	dbaccess.SetDBPool(dbPool)
	dbaccess.SetLogger(logger)

	drivenadapters.SetMailDialer(mailDialer)

	logics.SetLogger(logger)

	driveradapters.SetEnforcer(e)
	driveradapters.SetLogger(logger)

	s := &Server{
		userHandler: driveradapters.NewUserHandler(),
		oidcHandler: driveradapters.NewOIDCHandler(),
		config:      config,
	}

	s.Start()

	select {}
}
