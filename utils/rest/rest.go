package rest

import (
	"encoding/json"
	"net/http"

	"UserManagement/utils/rest/errors"

	"github.com/gin-gonic/gin"
)

type contextKey string

const (
	TokenKey  contextKey = "token"
	ClaimsKey contextKey = "claims"
	URIKey    contextKey = "uri"
	MethodKey contextKey = "method"
)

func ReplyError(c *gin.Context, err error) {
	var code int
	var body []byte

	switch e := err.(type) {
	case *errors.HTTPError:
		code = e.StatusCode()
		body, _ = json.Marshal(e)
	default:
		code = http.StatusInternalServerError
		body = []byte(err.Error())
	}

	c.Writer.WriteHeader(code)
	c.Writer.Write(body)
}

func ReplyOK(c *gin.Context, statusCode int, data interface{}) {
	var body []byte

	if data != nil {
		body, _ = json.Marshal(data)
	}

	c.Writer.WriteHeader(statusCode)
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Write(body)
}
