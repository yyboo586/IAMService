package driveradapters

import (
	"net/http"
	"sync"

	"UserManagement/interfaces"
	"UserManagement/logics"

	"github.com/casbin/casbin/v2"
	"github.com/xeipuuv/gojsonschema"

	"UserManagement/utils/rest"
	"UserManagement/utils/rest/middleware"

	"github.com/gin-gonic/gin"
)

var (
	uOnce sync.Once
	u     *UserHandler
)

type UserHandler struct {
	logicsUser       interfaces.LogicsUser
	e                *casbin.Enforcer
	userCreateSchema *gojsonschema.Schema
	userLoginSchema  *gojsonschema.Schema
}

func NewUserHandler() *UserHandler {
	uOnce.Do(func() {
		userCreateSchema, err := gojsonschema.NewSchema(gojsonschema.NewStringLoader(userCreateString))
		if err != nil {
			panic(err)
		}
		userLoginSchema, err := gojsonschema.NewSchema(gojsonschema.NewStringLoader(userLoginString))
		if err != nil {
			panic(err)
		}
		u = &UserHandler{
			logicsUser:       logics.NewUser(),
			e:                enforcer,
			userCreateSchema: userCreateSchema,
			userLoginSchema:  userLoginSchema,
		}
	})
	return u
}

func (u *UserHandler) RegisterPublic(engine *gin.Engine) {
	checkRequired := engine.Group("/", middleware.AuthRequired(), middleware.PermissionRequired(u.e))
	{
		checkRequired.GET("/api/v1/user-management/users/:id", u.getUserInfo)

		checkRequired.POST("/api/v1/user-management/users", u.create)
	}

	engine.Handle(http.MethodGet, "/api/v1/user-management/ready", u.ready)
	engine.Handle(http.MethodGet, "/api/v1/user-management/health", u.health)

	engine.Handle(http.MethodPost, "/api/v1/user-management/user-login", u.login)
}

func (u *UserHandler) create(c *gin.Context) {
	userInfo := interfaces.NewUser()
	defer interfaces.FreeUser(userInfo)

	body, err := Validate(c, u.userCreateSchema)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}

	userInfo.Name = body.(map[string]interface{})["name"].(string)
	userInfo.Password = body.(map[string]interface{})["password"].(string)

	err = u.logicsUser.Create(c.Request.Context(), userInfo)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}

	rest.ReplyOK(c, http.StatusCreated, nil)
}

func (u *UserHandler) login(c *gin.Context) {
	body, err := Validate(c, u.userLoginSchema)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}

	name := body.(map[string]interface{})["name"].(string)
	password := body.(map[string]interface{})["password"].(string)

	id, jwtTokenStr, err := u.logicsUser.Login(name, password)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}

	data := map[string]interface{}{
		"id":    id,
		"token": jwtTokenStr,
	}
	rest.ReplyOK(c, http.StatusOK, data)
}

func (u *UserHandler) getUserInfo(c *gin.Context) {
	id := c.Param("id")

	userInfo, err := u.logicsUser.GetUserInfo(c.Request.Context(), id)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}

	data := map[string]interface{}{
		"id":   userInfo.ID,
		"name": userInfo.Name,
	}
	rest.ReplyOK(c, http.StatusOK, data)
}

func (u *UserHandler) health(c *gin.Context) {
	rest.ReplyOK(c, http.StatusOK, nil)
}

func (u *UserHandler) ready(c *gin.Context) {
	rest.ReplyOK(c, http.StatusOK, nil)
}
