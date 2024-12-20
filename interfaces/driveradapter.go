package interfaces

import "github.com/gin-gonic/gin"

type RESTHandler interface {
	RegisterPublic(*gin.Engine)
}

type contextKey string

const (
	TokenKey  contextKey = "token"
	ClaimsKey contextKey = "claims"
	URIKey    contextKey = "uri"
	MethodKey contextKey = "method"
)
