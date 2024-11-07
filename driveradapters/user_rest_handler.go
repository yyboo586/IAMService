package driveradapters

import (
	"ServiceA/interfaces"
	"ServiceA/logics"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/yyboo586/utils/rest"
)

var (
	uOnce sync.Once
	u     *UserHandler
)

type UserHandler struct {
	logicsUser interfaces.LogicsUser
}

func NewUserHandler() *UserHandler {
	uOnce.Do(func() {
		u = &UserHandler{
			logicsUser: logics.NewUser(),
		}
	})
	return u
}

func (u *UserHandler) RegisterPublic(engine *gin.Engine) {
	engine.Handle(http.MethodGet, "/api/v1/ServiceA/users/:id", u.getUserInfo)

	engine.Handle(http.MethodPost, "/api/v1/ServiceA/users", u.create)
	engine.Handle(http.MethodPost, "/api/v1/ServiceA/user-login", u.login)
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

	err = u.logicsUser.Create(user)
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
