package main

import (
	"UserManagement/dbaccess"
	"UserManagement/driveradapters"
	"UserManagement/interfaces"
	"UserManagement/logics"

	"github.com/casbin/casbin/v2"

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

	privateKey, _ := rsaUtils.LoadPrivateKey()

	e, err := casbin.NewEnforcer("model.conf", "policy.csv")
	if err != nil {
		panic(err)
	}

	// 依赖注入
	dbaccess.SetDBPool(dbPool)

	logics.SetPrivateKey(privateKey)
	logics.SetDBPool(dbPool)

	s := &Server{
		userHandler: driveradapters.NewUserHandler(e),
	}

	s.Start()

	select {}
}
