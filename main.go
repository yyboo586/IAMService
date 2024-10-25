package main

import (
	"ServiceA/dbaccess"
	"ServiceA/driveradapters"
	"ServiceA/interfaces"

	"github.com/gin-gonic/gin"
	"github.com/yyboo586/utils/db"
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
	dbPool, err := db.NewDB("root", "12345678", "localhost", 3306, "ServiceA")
	if err != nil {
		panic(err)
	}

	dbaccess.SetDBPool(dbPool)

	s := &Server{
		userHandler: driveradapters.NewUserHandler(),
	}

	s.Start()

	select {}
}
