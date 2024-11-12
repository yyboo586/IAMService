package driveradapters

import (
	"net/http"
	"sync"

	"UserManagement/interfaces"
	"UserManagement/logics"

	"github.com/casbin/casbin/v2"

	"UserManagement/utils/rest"
	"UserManagement/utils/rest/middleware"

	"github.com/gin-gonic/gin"
)

var (
	uOnce sync.Once
	u     *UserHandler
)

type UserHandler struct {
	logicsUser interfaces.LogicsUser
	e          *casbin.Enforcer
}

func NewUserHandler(e *casbin.Enforcer) *UserHandler {
	uOnce.Do(func() {
		u = &UserHandler{
			logicsUser: logics.NewUser(),
			e:          e,
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
	i, err := validate(c)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}

	user := &interfaces.User{}
	user.Name = i.(map[string]interface{})["name"].(string)
	user.Password = i.(map[string]interface{})["password"].(string)

	err = u.logicsUser.Create(c.Request.Context(), user)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}

	rest.ReplyOK(c, http.StatusCreated, nil)
}

func (u *UserHandler) login(c *gin.Context) {
	i, err := validate(c)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}

	name := i.(map[string]interface{})["name"].(string)
	password := i.(map[string]interface{})["password"].(string)

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
