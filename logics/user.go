package logics

import (
	"ServiceA/dbaccess"
	"ServiceA/interfaces"
	"crypto/rsa"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	myjwt "github.com/yyboo586/utils/myJWT"
	"github.com/yyboo586/utils/rest"

	"github.com/yyboo586/utils/uuid"
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
	dbUser     interfaces.DBUser
}

func NewUser() *user {
	uOnce.Do(func() {
		u = &user{
			pwdRegex:   regexp.MustCompile(`^[a-zA-Z0-9]{6,12}$`),
			nameRegex:  regexp.MustCompile(`^[\p{Han}a-zA-Z0-9]{1,6}$`),
			privateKey: privateKey,
			dbUser:     dbaccess.NewUser(),
		}
	})
	return u
}

func (u *user) Create(user *interfaces.User) (err error) {
	user.Name = strings.Trim(user.Name, " ")

	if !u.nameRegex.MatchString(user.Name) {
		return rest.NewHTTPError(http.StatusBadRequest, "invalid name", nil)
	}

	if !u.pwdRegex.MatchString(user.Password) {
		return rest.NewHTTPError(http.StatusBadRequest, "invalid password", nil)
	}

	_, exists, err := u.dbUser.FetchByName(user.Name)
	if err != nil {
		return rest.NewHTTPError(http.StatusInternalServerError, err.Error(), nil)
	}
	if exists {
		return rest.NewHTTPError(http.StatusConflict, "name already exists", nil)
	}

	cipherText, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	user.Password = string(cipherText)

	user.ID = uuid.NewV4().String()

	return u.dbUser.Create(user)
}

func (u *user) Login(name, passwd string) (id string, jwtTokenStr string, err error) {
	if name == "" || passwd == "" {
		err = rest.NewHTTPError(http.StatusBadRequest, "invalid name or password", nil)
		return
	}

	// 引入固定时间延迟
	// 密码比较的时间复杂度不一致：
	// - 当用户名不存在或密码错误时，返回的错误信息相同，这有助于防止攻击者通过响应时间来判断用户名是否存在。
	// - 但是，bcrypt.CompareHashAndPassword 的执行时间是固定的，而 FetchByName 和 UpdateLoginTime 的执行时间可能因数据库状态而异。
	// - 为了进一步提高安全性，可以在密码验证之前引入一个固定时间的延迟，以确保所有路径的执行时间大致相同。
	defer func() {
		time.Sleep(100 * time.Millisecond)
	}()

	user, exists, err := u.dbUser.FetchByName(name)
	if err != nil {
		err = rest.NewHTTPError(http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if !exists {
		err = rest.NewHTTPError(http.StatusBadRequest, "invalid name or password", nil)
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(passwd)); err != nil {
		err = rest.NewHTTPError(http.StatusBadRequest, "invalid name or password", nil)
		return
	}

	if err = u.dbUser.UpdateLoginTime(user.ID); err != nil {
		err = rest.NewHTTPError(http.StatusInternalServerError, err.Error(), nil)
		return
	}

	if jwtTokenStr, err = myjwt.Sign(user.ID, nil, u.privateKey); err != nil {
		err = rest.NewHTTPError(http.StatusInternalServerError, err.Error(), nil)
		return
	}

	return user.ID, jwtTokenStr, nil
}
