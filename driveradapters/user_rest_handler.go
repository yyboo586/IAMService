package driveradapters

import (
	"context"
	"net/http"
	"strings"
	"sync"

	"github.com/yyboo586/IAMService/interfaces"
	"github.com/yyboo586/IAMService/logics"
	"github.com/yyboo586/common/logUtils"

	"github.com/casbin/casbin/v2"
	"github.com/xeipuuv/gojsonschema"

	"github.com/yyboo586/common/rest"

	"github.com/gin-gonic/gin"
)

var (
	_ interfaces.RESTHandler = (*UserHandler)(nil)
)

var (
	uOnce sync.Once
	u     *UserHandler
)

type UserHandler struct {
	logicsUser       interfaces.LogicsUser
	logicsJWT        interfaces.LogicsJWT
	e                *casbin.Enforcer
	logger           *logUtils.Logger
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
			logicsJWT:        logics.NewLogicsJWT(),
			e:                enforcer,
			logger:           loggerInstance,
			userCreateSchema: userCreateSchema,
			userLoginSchema:  userLoginSchema,
		}
	})
	return u
}

func (u *UserHandler) RegisterPublic(engine *gin.Engine) {
	engine.GET("/api/v1/IAMService/users/:id", u.getUserInfo)
	engine.POST("/api/v1/IAMService/users", u.create)

	engine.Handle(http.MethodGet, "/api/v1/IAMService/ready", u.ready)
	engine.Handle(http.MethodGet, "/api/v1/IAMService/health", u.health)

	engine.Handle(http.MethodPost, "/api/v1/IAMService/user-login", u.login)
}

func (u *UserHandler) authRequired(c *gin.Context) (ctx context.Context, err error) {
	tokenInfos := strings.Split(c.GetHeader("Authorization"), " ")
	if len(tokenInfos) < 2 {
		return c.Request.Context(), rest.NewHTTPError(http.StatusUnauthorized, "token is invalid", nil)
	}

	claims, err := u.logicsJWT.Verify(tokenInfos[1])
	if err != nil {
		return c.Request.Context(), rest.NewHTTPError(http.StatusUnauthorized, "token is invalid", nil)
	}

	// 将令牌信息放入上下文
	ctx = context.WithValue(context.WithValue(c.Request.Context(), interfaces.ClaimsKey, claims.ExtClaims), interfaces.TokenKey, tokenInfos[1])
	return ctx, nil
}

func (u *UserHandler) create(c *gin.Context) {
	ctx, err := u.authRequired(c)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}

	body, err := Validate(c, u.userCreateSchema)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}

	userInfo := &interfaces.User{}
	userInfo.Name = body.(map[string]interface{})["name"].(string)
	userInfo.Password = body.(map[string]interface{})["password"].(string)
	if email, ok := body.(map[string]interface{})["email"]; ok {
		userInfo.Email = email.(string)
	}

	err = u.logicsUser.Create(ctx, userInfo)
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
	ctx, err := u.authRequired(c)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}

	id := c.Param("id")
	userInfo, err := u.logicsUser.GetUserInfo(ctx, id)
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
