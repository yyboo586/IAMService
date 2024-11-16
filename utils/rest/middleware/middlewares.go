package middleware

import (
	"context"
	"crypto/rsa"
	"log"
	"net/http"
	"strings"

	jwtUtils "UserManagement/utils/jwt"
	"UserManagement/utils/rest"
	errUtils "UserManagement/utils/rest/errors"
	rsaUtils "UserManagement/utils/rsa"

	"github.com/casbin/casbin/v2"

	"github.com/gin-gonic/gin"
)

var (
	err        error
	privateKey *rsa.PrivateKey
)

func init() {
	privateKey, err = rsaUtils.LoadPrivateKey()
	if err != nil {
		panic(err)
	}
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenInfos := strings.Split(c.GetHeader("Authorization"), " ")
		if len(tokenInfos) < 2 {
			rest.ReplyError(c, errUtils.NewHTTPError(http.StatusUnauthorized, "token is invalid", nil))
			c.Abort()
			return
		}

		extClaims, err := jwtUtils.Verify(tokenInfos[1], privateKey)
		if err != nil {
			rest.ReplyError(c, errUtils.NewHTTPError(http.StatusUnauthorized, "token is invalid", nil))
			c.Abort()
			return
		}

		// 将令牌信息放入上下文
		ctx := context.WithValue(context.WithValue(context.WithValue(c.Request.Context(), rest.ClaimsKey, extClaims), rest.TokenKey, tokenInfos[1]), rest.URIKey, c.Request.RequestURI)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func PermissionRequired(e *casbin.Enforcer) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := c.Request.Context().Value(rest.ClaimsKey)
		sub := claims.(map[string]interface{})["name"]
		obj := c.Request.URL.Path
		act := c.Request.Method

		log.Println(sub, obj, act)

		ok, err := e.Enforce(sub, obj, act)
		if err != nil {
			rest.ReplyError(c, errUtils.NewHTTPError(http.StatusInternalServerError, err.Error(), nil))
			c.Abort()
			return
		}
		if !ok {
			rest.ReplyError(c, errUtils.NewHTTPError(http.StatusForbidden, "no permission", nil))
			c.Abort()
			return
		}

		c.Next()
	}
}
