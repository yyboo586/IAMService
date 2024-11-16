package driveradapters

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	errUtils "UserManagement/utils/rest/errors"
	_ "embed"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/xeipuuv/gojsonschema"
)

func validate(c *gin.Context) (i interface{}, err error) {
	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}

	return i, json.Unmarshal(data, &i)
}

var (
	//go:embed jsonschema/user_create.json
	userCreateString string
	//go:embed jsonschema/user_login.json
	userLoginString string
)

var (
	enforcer *casbin.Enforcer
)

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
			log.Println(tmpErr)
			details[tmpErr.Field()] = tmpErr.Description()
		}
		return nil, errUtils.NewHTTPError(http.StatusBadRequest, "invalid request body", details)
	}

	return i, nil
}

func SetEnforcer(e *casbin.Enforcer) {
	enforcer = e
}
