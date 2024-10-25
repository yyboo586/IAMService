package interfaces

import "github.com/gin-gonic/gin"

type RESTHandler interface {
	RegisterPublic(*gin.Engine)
}
