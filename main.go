package main

import (
	"UserManagement/dbaccess"
	"UserManagement/drivenadapters"
	"UserManagement/driveradapters"
	"UserManagement/interfaces"
	"UserManagement/logics"

	"github.com/casbin/casbin/v2"
	"github.com/go-mail/mail/v2"

	configUtils "UserManagement/utils/config"
	dbUtils "UserManagement/utils/db"
	rsaUtils "UserManagement/utils/rsa"

	"github.com/gin-gonic/gin"
)

type Server struct {
	userHandler interfaces.RESTHandler
	config      *configUtils.Config
}

func (s *Server) Start() {
	go func() {
		engine := gin.Default()

		s.userHandler.RegisterPublic(engine)

		if err := engine.Run(s.config.Server.Addr); err != nil {
			panic(err)
		}
	}()
}

func main() {
	config := configUtils.Default()

	dbPool, err := dbUtils.NewDB(config.DB.User, config.DB.Pass, config.DB.Host, config.DB.Port, config.DB.DBName)
	if err != nil {
		panic(err)
	}

	mailDialer := mail.NewDialer(config.Mailer.Host, config.Mailer.Port, config.Mailer.User, config.Mailer.Pass)

	privateKey, _ := rsaUtils.LoadPrivateKey()

	e, err := casbin.NewEnforcer("model.conf", "policy.csv")
	if err != nil {
		panic(err)
	}

	// 依赖注入
	dbaccess.SetDBPool(dbPool)

	drivenadapters.SetMailDialer(mailDialer)

	logics.SetPrivateKey(privateKey)
	logics.SetDBPool(dbPool)

	driveradapters.SetEnforcer(e)

	s := &Server{
		userHandler: driveradapters.NewUserHandler(),
		config:      config,
	}

	s.Start()

	select {}
}
