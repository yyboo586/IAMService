package driveradapters

import (
	"encoding/json"
	"io"
	"net/http"

	_ "embed"

	errUtils "github.com/yyboo586/IAMService/utils/rest/errors"
	"github.com/yyboo586/common/logUtils"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/xeipuuv/gojsonschema"
)

var (
	//go:embed jsonschema/user_create.json
	userCreateString string
	//go:embed jsonschema/user_login.json
	userLoginString string
)

var (
	enforcer       *casbin.Enforcer
	loggerInstance *logUtils.Logger
)

func SetEnforcer(e *casbin.Enforcer) {
	enforcer = e
}

func SetLogger(i *logUtils.Logger) {
	loggerInstance = i
}

func Validate(c *gin.Context, schema *gojsonschema.Schema) (i interface{}, err error) {
	data, _ := io.ReadAll(c.Request.Body)
	_ = json.Unmarshal(data, &i)

	// return schema.Validate()
	results, err := schema.Validate(gojsonschema.NewGoLoader(i))
	if err != nil {
		return nil, err
	}
	if !results.Valid() {
		details := make(map[string]interface{})
		for _, tmpErr := range results.Errors() {
			// log.Println(tmpErr)
			details[tmpErr.Field()] = tmpErr.Description()
		}
		return nil, errUtils.NewHTTPError(http.StatusBadRequest, "invalid request body", details)
	}

	return i, nil
}
