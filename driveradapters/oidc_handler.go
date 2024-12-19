package driveradapters

import (
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/yyboo586/IAMService/interfaces"
	"github.com/yyboo586/IAMService/logics"
	"github.com/yyboo586/IAMService/utils/rest"
	"github.com/yyboo586/IAMService/utils/rest/errors"
)

var (
	oidcOnce    sync.Once
	oidcHandler *OIDCHandler
)

type OIDCHandler struct {
	loJWT interfaces.LogicsJWT
}

func NewOIDCHandler() *OIDCHandler {
	oidcOnce.Do(func() {
		oidcHandler = &OIDCHandler{
			loJWT: logics.NewLogicsJWT(),
		}
	})

	return oidcHandler
}

func (o *OIDCHandler) RegisterPublic(engine *gin.Engine) {
	engine.GET("/api/v1/IAMService/jwk/:id", o.GetPublicKey)
	engine.POST("/api/v1/IAMService/jwt/revoke", o.RevokeToken)
}

func (o *OIDCHandler) GetPublicKey(c *gin.Context) {
	kid := c.Param("id")

	key, err := o.loJWT.GetPublicKey(kid)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}

	rest.ReplyOK(c, http.StatusOK, key)
}

func (o *OIDCHandler) RevokeToken(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		rest.ReplyError(c, errors.NewHTTPError(http.StatusUnauthorized, "token is required", nil))
		return
	}

	err := o.loJWT.RevokeToken(strings.TrimPrefix(token, "Bearer "))
	if err != nil {
		rest.ReplyError(c, err)
		return
	}

	rest.ReplyOK(c, http.StatusOK, nil)
}
