package logics

import (
	"context"
	"crypto/rsa"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/yyboo586/IAMService/dbaccess"
	"github.com/yyboo586/IAMService/interfaces"
	"github.com/yyboo586/common/logUtils"

	"github.com/yyboo586/IAMService/utils/rest"
	errUtils "github.com/yyboo586/IAMService/utils/rest/errors"

	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	uOnce sync.Once
	u     *user
)

type user struct {
	pwdRegex   *regexp.Regexp
	nameRegex  *regexp.Regexp
	privateKey *rsa.PrivateKey
	logger     *logUtils.Logger
	mailer     interfaces.LogicsMailer
	loJWT      interfaces.LogicsJWT
	dbUser     interfaces.DBUser
}

func NewUser() *user {
	uOnce.Do(func() {
		u = &user{
			pwdRegex:   regexp.MustCompile(`^[a-zA-Z0-9]{6,12}$`),
			nameRegex:  regexp.MustCompile(`^[\p{Han}a-zA-Z0-9]{1,10}$`),
			privateKey: privateKey,
			logger:     loggerInstance,
			mailer:     NewLogicsMailer(),
			loJWT:      NewLogicsJWT(),
			dbUser:     dbaccess.NewUser(),
		}
	})
	return u
}

func (u *user) Create(ctx context.Context, userInfo *interfaces.User) (err error) {
	if err = u.validate(userInfo); err != nil {
		return err
	}

	_, exists, err := u.dbUser.FetchByName(userInfo.Name)
	if err != nil {
		return errUtils.NewHTTPError(http.StatusInternalServerError, err.Error(), nil)
	}
	if exists {
		return errUtils.NewHTTPError(http.StatusConflict, "name already exists", nil)
	}

	cipherText, _ := bcrypt.GenerateFromPassword([]byte(userInfo.Password), bcrypt.DefaultCost)
	userInfo.Password = string(cipherText)

	userInfo.ID = uuid.Must(uuid.NewV4()).String()

	err = u.dbUser.Create(userInfo)
	if err != nil {
		return errUtils.NewHTTPError(http.StatusInternalServerError, err.Error(), nil)
	}

	// 不保证可以发送成功
	if userInfo.Email != "" {
		go func() {
			msg := &interfaces.MailMessage{
				ID: userInfo.ID,
				To: userInfo.Email,
			}

			if err = u.mailer.SendMail(ctx, interfaces.UserWelcome, msg); err != nil {
				u.logger.Errorf("failed to send email: %v", err)
			} else {
				u.logger.Infof("send email successfully")
			}
		}()
	}

	return nil
}

func (u *user) Login(name, passwd string) (id string, jwtTokenStr string, err error) {
	if name == "" || passwd == "" {
		err = errUtils.NewHTTPError(http.StatusBadRequest, "invalid name or password", nil)
		return
	}

	// 引入固定时间延迟
	// 密码比较的时间复杂度不一致：
	// - 当用户名不存在或密码错误时，返回的错误信息相同，这有助于防止攻击者通过响应时间来判断用户名是否存在。
	// - 但是，bcrypt.CompareHashAndPassword 的执行时间是固定的，而 FetchByName 和 UpdateLoginTime 的执行时间可能因数据库状态而异。
	// - 为了进一步提高安全性，可以在密码验证之前引入一个固定时间的延迟，以确保所有路径的执行时间大致相同。
	// defer func() {
	// 	time.Sleep(100 * time.Millisecond)
	// }()

	user, exists, err := u.dbUser.FetchByName(name)
	if err != nil {
		err = errUtils.NewHTTPError(http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if !exists {
		err = errUtils.NewHTTPError(http.StatusBadRequest, "invalid name or password", nil)
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(passwd)); err != nil {
		err = errUtils.NewHTTPError(http.StatusBadRequest, "invalid name or password", nil)
		return
	}

	if err = u.dbUser.UpdateLoginTime(user.ID); err != nil {
		err = errUtils.NewHTTPError(http.StatusInternalServerError, err.Error(), nil)
		return
	}

	claims := map[string]interface{}{
		"id":   user.ID,
		"name": user.Name,
	}
	if jwtTokenStr, err = u.loJWT.Sign(user.ID, claims, "id_token", "HS256"); err != nil {
		err = errUtils.NewHTTPError(http.StatusInternalServerError, err.Error(), nil)
		return
	}

	return user.ID, jwtTokenStr, nil
}

func (u *user) GetUserInfo(ctx context.Context, id string) (userInfo *interfaces.User, err error) {
	// log.Printf("DEBUG: ExtClaims: %v, jwtToken: %v\n", ctx.Value(interfaces.ClaimsKey).(map[string]interface{}), ctx.Value(interfaces.TokenKey).(string))
	// 只能获取自己的数据
	if id != ctx.Value(rest.ClaimsKey).(map[string]interface{})["id"].(string) {
		err = errUtils.NewHTTPError(http.StatusForbidden, "no permission", nil)
		return
	}

	var exists bool
	userInfo, exists, err = u.dbUser.GetUserInfoByID(id)
	if err != nil {
		err = errUtils.NewHTTPError(http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if !exists {
		err = errUtils.NewHTTPError(http.StatusNotFound, "user not found", nil)
		return
	}

	return
}

func (u *user) validate(userInfo *interfaces.User) error {
	userInfo.Name = strings.Trim(userInfo.Name, " ")

	if !u.nameRegex.MatchString(userInfo.Name) {
		return errUtils.NewHTTPError(http.StatusBadRequest, "invalid name", nil)
	}

	if !u.pwdRegex.MatchString(userInfo.Password) {
		return errUtils.NewHTTPError(http.StatusBadRequest, "invalid password", nil)
	}

	return nil
}
