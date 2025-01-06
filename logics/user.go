package logics

import (
	"context"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/yyboo586/IAMService/dbaccess"
	"github.com/yyboo586/IAMService/interfaces"
	"github.com/yyboo586/common/logUtils"

	"github.com/yyboo586/common/rest"

	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	uOnce sync.Once
	u     *user
)

type user struct {
	pwdRegex  *regexp.Regexp
	nameRegex *regexp.Regexp
	logger    *logUtils.Logger
	mailer    interfaces.LogicsMailer
	jwt       interfaces.LogicsJWT
	dbUser    interfaces.DBUser
}

func NewUser() interfaces.LogicsUser {
	uOnce.Do(func() {
		u = &user{
			pwdRegex:  regexp.MustCompile(`^[a-zA-Z0-9]{6,12}$`),
			nameRegex: regexp.MustCompile(`^[\p{Han}a-zA-Z0-9]{1,10}$`),
			logger:    loggerInstance,
			mailer:    NewLogicsMailer(),
			jwt:       NewLogicsJWT(),
			dbUser:    dbaccess.NewUser(),
		}
	})
	return u
}

func (u *user) Create(ctx context.Context, userInfo *interfaces.User) (err error) {
	// 校验注册信息是否符合规范
	if err = u.validateUserInfo(userInfo); err != nil {
		return err
	}

	// 检查用户名是否已存在
	_, exists, err := u.dbUser.GetUserInfoByName(userInfo.Name)
	if err != nil {
		u.logger.Errorf("failed to get user info by name: %v", err)
		return rest.NewHTTPError(http.StatusInternalServerError, "服务器内部错误，请联系管理员", nil)
	}
	if exists {
		return rest.NewHTTPError(http.StatusConflict, "用户名已存在", nil)
	}

	cipherText, err := bcrypt.GenerateFromPassword([]byte(userInfo.Password), bcrypt.DefaultCost)
	if err != nil {
		u.logger.Errorf("failed to generate password: %v", err)
		return rest.NewHTTPError(http.StatusInternalServerError, "服务器内部错误，请联系管理员", nil)
	}
	userInfo.Password = string(cipherText)
	userInfo.ID = uuid.Must(uuid.NewV4()).String()

	if err = u.dbUser.Create(userInfo); err != nil {
		u.logger.Errorf("failed to create user: %v", err)
		return rest.NewHTTPError(http.StatusInternalServerError, "服务器内部错误，请联系管理员", nil)
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
		return "", "", rest.NewHTTPError(http.StatusBadRequest, "invalid name or password", nil)
	}

	// 引入固定时间延迟
	// 密码比较的时间复杂度不一致：
	// - 当用户名不存在或密码错误时，返回的错误信息相同，这有助于防止攻击者通过响应时间来判断用户名是否存在。
	// - 但是，bcrypt.CompareHashAndPassword 的执行时间是固定的，而 FetchByName 和 UpdateLoginTime 的执行时间可能因数据库状态而异。
	// - 为了进一步提高安全性，可以在密码验证之前引入一个固定时间的延迟，以确保所有路径的执行时间大致相同。
	// defer func() {
	// 	time.Sleep(100 * time.Millisecond)
	// }()

	user, exists, err := u.dbUser.GetUserInfoByName(name)
	if err != nil {
		u.logger.Debugf("failed to get user info by name: %v", err)
		return "", "", rest.InternalServerError
	}
	if !exists {
		return "", "", rest.NewHTTPError(http.StatusBadRequest, "invalid name or password", nil)
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(passwd)); err != nil {
		u.logger.Debugf("failed to compare hash and password: %v", err)
		return "", "", rest.NewHTTPError(http.StatusBadRequest, "invalid name or password", nil)
	}

	if err = u.dbUser.UpdateLoginTime(user.ID); err != nil {
		u.logger.Debugf("failed to update login time: %v", err)
		return "", "", rest.InternalServerError
	}

	claims := map[string]interface{}{
		"id":   user.ID,
		"name": user.Name,
	}
	if jwtTokenStr, err = u.jwt.Sign(user.ID, claims, "ac_token", "RS256"); err != nil {
		u.logger.Debugf("failed to sign jwt: %v", err)
		return "", "", rest.InternalServerError
	}

	return user.ID, jwtTokenStr, nil
}

func (u *user) GetUserInfo(ctx context.Context, id string) (userInfo *interfaces.User, err error) {
	u.logger.Debugf("ExtClaims: %v, jwtToken: %v\n", ctx.Value(interfaces.ClaimsKey).(map[string]interface{}), ctx.Value(interfaces.TokenKey).(string))
	// 只能获取自己的数据
	if id != ctx.Value(interfaces.ClaimsKey).(map[string]interface{})["id"].(string) {
		return nil, rest.Forbidden
	}

	var exists bool
	userInfo, exists, err = u.dbUser.GetUserInfoByID(id)
	if err != nil {
		u.logger.Debugf("failed to get user info by id: %v", err)
		return nil, rest.InternalServerError
	}
	if !exists {
		u.logger.Debugf("user not found")
		return nil, rest.NotFound
	}

	return userInfo, nil
}

func (u *user) validateUserInfo(userInfo *interfaces.User) error {
	userInfo.Name = strings.Trim(userInfo.Name, " ")

	if !u.nameRegex.MatchString(userInfo.Name) {
		return rest.NewHTTPError(http.StatusBadRequest, "invalid name", nil)
	}

	if !u.pwdRegex.MatchString(userInfo.Password) {
		return rest.NewHTTPError(http.StatusBadRequest, "invalid password", nil)
	}

	return nil
}
