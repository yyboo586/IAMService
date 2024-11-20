package main

import (
	"UserManagement/dbaccess"
	"UserManagement/drivenadapters"
	"UserManagement/driveradapters"
	"UserManagement/interfaces"
	"UserManagement/logics"

	"github.com/casbin/casbin/v2"
	"github.com/go-mail/mail/v2"

	dbUtils "UserManagement/utils/db"
	rsaUtils "UserManagement/utils/rsa"

	"github.com/gin-gonic/gin"
)

type Server struct {
	userHandler interfaces.RESTHandler
}

func (s *Server) Start() {
	go func() {
		engine := gin.Default()

		s.userHandler.RegisterPublic(engine)

		if err := engine.Run(":10001"); err != nil {
			panic(err)
		}
	}()
}

func main() {
	dbPool, err := dbUtils.NewDB("root", "12345678", "localhost", 3306, "ServiceA")
	if err != nil {
		panic(err)
	}

	mailDialer := mail.NewDialer("sandbox.smtp.mailtrap.io", 25, "8df5de08b5f13f", "fcb5034938135d")

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
	}

	s.Start()

	select {}
}
