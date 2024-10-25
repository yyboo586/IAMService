package logics

import (
	"ServiceA/interfaces"
	"ServiceA/interfaces/mock"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"github.com/yyboo586/utils/rest"
	"golang.org/x/crypto/bcrypt"
)

func newUser(dbUser interfaces.DBUser) *user {
	return &user{
		pwdRegex:  regexp.MustCompile(`^[a-zA-Z0-9]{6,12}$`),
		nameRegex: regexp.MustCompile(`^[\p{Han}a-zA-Z0-9]{1,6}$`),
		dbUser:    dbUser,
	}
}

func TestCreate(t *testing.T) {
	Convey("Create", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		dbUser := mock.NewMockDBUser(ctrl)
		user := newUser(dbUser)

		Convey("用户名校验", func() {
			testCases := []struct {
				name        string
				expectedErr error
			}{
				{"", rest.NewHTTPError(http.StatusBadRequest, "invalid name", nil)},
				{"test user", rest.NewHTTPError(http.StatusBadRequest, "invalid name", nil)},
				{"test&user", rest.NewHTTPError(http.StatusBadRequest, "invalid name", nil)},
				{"testuser", rest.NewHTTPError(http.StatusBadRequest, "invalid name", nil)},
				{"七个中文字符啊", rest.NewHTTPError(http.StatusBadRequest, "invalid name", nil)},
			}

			for _, tc := range testCases {
				Convey(fmt.Sprintf("创建失败，用户名: %s", tc.name), func() {
					invalidUser := &interfaces.User{
						Name:     tc.name,
						Password: "TestPass123",
					}

					err := user.Create(invalidUser)

					assert.Equal(t, tc.expectedErr, err)
				})
			}
		})

		Convey("密码校验", func() {
			testCases := []struct {
				password    string
				expectedErr error
			}{
				{"", rest.NewHTTPError(http.StatusBadRequest, "invalid password", nil)},
				{"TestPass!@#", rest.NewHTTPError(http.StatusBadRequest, "invalid password", nil)},
				{"12345", rest.NewHTTPError(http.StatusBadRequest, "invalid password", nil)},
				{"1234567890123", rest.NewHTTPError(http.StatusBadRequest, "invalid password", nil)},
			}

			for _, tc := range testCases {
				Convey(fmt.Sprintf("创建失败，用户密码: %s", tc.password), func() {
					invalidUser := &interfaces.User{
						Name:     "tom",
						Password: tc.password,
					}

					err := user.Create(invalidUser)

					assert.Equal(t, tc.expectedErr, err)
				})
			}
		})

		Convey("数据库错误", func() {
			validUser := &interfaces.User{
				Name:     "tom",
				Password: "123456",
			}

			dbUser.EXPECT().FetchByName(gomock.Any()).Return(nil, false, errors.New("database error"))

			err := user.Create(validUser)

			assert.Equal(t, rest.NewHTTPError(http.StatusInternalServerError, "database error", nil), err)
		})

		Convey("用户名已存在", func() {
			validUser := &interfaces.User{
				Name:     "tom",
				Password: "123456",
			}

			dbUser.EXPECT().FetchByName(gomock.Any()).Return(nil, true, nil)

			err := user.Create(validUser)

			assert.Equal(t, rest.NewHTTPError(http.StatusConflict, "name already exists", nil), err)
		})

		Convey("创建成功", func() {
			testCases := []struct {
				name     string
				password string
			}{
				{"t", "TestPass123"},
				{"tttttt", "TestPass123"},
				{"一", "TestPass123"},
				{"六个中文字符", "TestPass123"},
				{"六个中文字符", "TestPa"},
				{"六个中文字符", "TestPaTestPa"},
			}
			for _, tc := range testCases {
				validUser := &interfaces.User{
					Name:     tc.name,
					Password: tc.password,
				}

				dbUser.EXPECT().FetchByName(gomock.Any()).Return(nil, false, nil)
				dbUser.EXPECT().Create(gomock.Any()).Return(nil)

				err := user.Create(validUser)

				assert.Equal(t, nil, err)
			}
		})
	})
}

func TestLogin(t *testing.T) {
	Convey("Login", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		dbUser := mock.NewMockDBUser(ctrl)
		user := newUser(dbUser)

		Convey("参数校验", func() {
			testCases := []struct {
				name        string
				passwd      string
				expectedErr error
			}{
				{"", "", rest.NewHTTPError(http.StatusBadRequest, "invalid name or password", nil)},
				{"tom", "", rest.NewHTTPError(http.StatusBadRequest, "invalid name or password", nil)},
				{"", "123456", rest.NewHTTPError(http.StatusBadRequest, "invalid name or password", nil)},
			}

			for _, tc := range testCases {
				Convey(fmt.Sprintf("登录失败，用户名: %s, 密码: %s", tc.name, tc.passwd), func() {
					id, err := user.Login(tc.name, tc.passwd)

					assert.Equal(t, "", id)
					assert.Equal(t, tc.expectedErr, err)
				})
			}
		})

		Convey("数据库错误, FetchByName failed", func() {
			dbUser.EXPECT().FetchByName(gomock.Any()).Return(nil, false, errors.New("database error"))

			id, err := user.Login("tom", "123456")

			assert.Equal(t, "", id)
			assert.Equal(t, rest.NewHTTPError(http.StatusInternalServerError, "database error", nil), err)
		})

		Convey("用户名不存在", func() {
			dbUser.EXPECT().FetchByName(gomock.Any()).Return(nil, false, nil)

			id, err := user.Login("nonexistent", "123456")

			assert.Equal(t, "", id)
			assert.Equal(t, rest.NewHTTPError(http.StatusBadRequest, "invalid name or password", nil), err)
		})

		Convey("密码错误", func() {
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)

			dbUser.EXPECT().FetchByName(gomock.Any()).Return(&interfaces.User{ID: "1", Name: "tom", Password: string(hashedPassword)}, true, nil)

			id, err := user.Login("tom", "wrongpassword")

			assert.Equal(t, "", id)
			assert.Equal(t, rest.NewHTTPError(http.StatusBadRequest, "invalid name or password", nil), err)
		})

		Convey("数据库错误, UpdateLoginTime failed", func() {
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)

			dbUser.EXPECT().FetchByName(gomock.Any()).Return(&interfaces.User{ID: "1", Name: "tom", Password: string(hashedPassword)}, true, nil)
			dbUser.EXPECT().UpdateLoginTime(gomock.Any()).Return(errors.New("database error"))

			id, err := user.Login("tom", "123456")

			assert.Equal(t, "", id)
			assert.Equal(t, rest.NewHTTPError(http.StatusInternalServerError, "database error", nil), err)
		})

		Convey("登录成功", func() {
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
			dbUser.EXPECT().FetchByName(gomock.Any()).Return(&interfaces.User{ID: "1", Name: "tom", Password: string(hashedPassword)}, true, nil)
			dbUser.EXPECT().UpdateLoginTime(gomock.Any()).Return(nil)

			id, err := user.Login("tom", "123456")
			assert.Equal(t, "1", id)
			assert.Equal(t, nil, err)
		})
	})
}
