package logics

import (
	"database/sql"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/yyboo586/IAMService/interfaces"
	"github.com/yyboo586/IAMService/interfaces/mock"
	"github.com/yyboo586/common/logUtils"
	"github.com/yyboo586/common/rest"
	"golang.org/x/crypto/bcrypt"

	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func setupUserTest(dbPool *sql.DB, jwt interfaces.LogicsJWT, outbox interfaces.LogicsOutbox, dbUser interfaces.DBUser) interfaces.LogicsUser {
	loggerInstance, _ := logUtils.NewLogger("debug")

	return &user{
		pwdRegex:  regexp.MustCompile(`^[a-zA-Z0-9]{6,12}$`),
		nameRegex: regexp.MustCompile(`^[\p{Han}a-zA-Z0-9]{1,6}$`),
		dbPool:    dbPool,
		logger:    loggerInstance,
		jwt:       jwt,
		outbox:    outbox,
		dbUser:    dbUser,
	}
}

func TestCreate(t *testing.T) {
	convey.Convey("Create", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		db, dbmock, err := sqlmock.New()
		assert.Equal(t, nil, err)
		defer db.Close()

		jwt := mock.NewMockLogicsJWT(ctrl)
		outbox := mock.NewMockLogicsOutbox(ctrl)
		dbUser := mock.NewMockDBUser(ctrl)
		user := setupUserTest(db, jwt, outbox, dbUser)

		ctx := context.Background()
		validUser := &interfaces.User{
			Name:     "tom",
			Password: "123456",
		}
		convey.Convey("用户名校验", func() {
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
				convey.Convey(fmt.Sprintf("创建失败，用户名: %s", tc.name), func() {
					invalidUser := &interfaces.User{
						Name:     tc.name,
						Password: "TestPass123",
					}

					err := user.Create(ctx, invalidUser)

					assert.Equal(t, tc.expectedErr, err)
				})
			}
		})
		convey.Convey("密码校验", func() {
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
				convey.Convey(fmt.Sprintf("创建失败，用户密码: %s", tc.password), func() {
					invalidUser := &interfaces.User{
						Name:     "tom",
						Password: tc.password,
					}

					err := user.Create(ctx, invalidUser)

					assert.Equal(t, tc.expectedErr, err)
				})
			}
		})
		convey.Convey("数据库错误1-GetUserInfoByName", func() {
			dbUser.EXPECT().GetUserInfoByName(gomock.Any()).Return(nil, false, errors.New("database error"))

			err := user.Create(ctx, validUser)
			assert.Equal(t, rest.NewHTTPError(http.StatusInternalServerError, "服务器内部错误，请联系管理员", nil), err)
		})
		convey.Convey("用户名已存在", func() {
			dbUser.EXPECT().GetUserInfoByName(gomock.Any()).Return(nil, true, nil)

			err := user.Create(ctx, validUser)

			assert.Equal(t, rest.NewHTTPError(http.StatusConflict, "用户名已存在", nil), err)
		})
		convey.Convey("数据库错误2-Create", func() {
			dbUser.EXPECT().GetUserInfoByName(gomock.Any()).Return(nil, false, nil)
			dbmock.ExpectBegin()
			dbUser.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errors.New("database error"))
			dbmock.ExpectRollback()

			err := user.Create(ctx, validUser)

			assert.Equal(t, rest.NewHTTPError(http.StatusInternalServerError, "服务器内部错误，请联系管理员", nil), err)
		})
		convey.Convey("创建成功", func() {
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
				validUser = &interfaces.User{
					Name:     tc.name,
					Password: tc.password,
				}
				dbUser.EXPECT().GetUserInfoByName(gomock.Any()).Return(nil, false, nil)
				dbmock.ExpectBegin()
				dbUser.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
				outbox.EXPECT().AddMessage(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				dbmock.ExpectCommit()

				err := user.Create(ctx, validUser)

				assert.Equal(t, nil, err)
			}
		})
	})
}

// TestLogin 测试登录功能
func TestLogin(t *testing.T) {
	convey.Convey("Login", t, func() {
		// 初始化测试环境
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		db, _, err := sqlmock.New()
		assert.Equal(t, nil, err)
		defer db.Close()

		jwt := mock.NewMockLogicsJWT(ctrl)
		dbUser := mock.NewMockDBUser(ctrl)
		user := setupUserTest(db, jwt, nil, dbUser)
		// 测试参数校验
		convey.Convey("参数校验", func() {
			testCases := []struct {
				name        string
				passwd      string
				expectedErr error
			}{
				{"", "", rest.NewHTTPError(http.StatusBadRequest, "invalid name or password", nil)},
				{"tom", "", rest.NewHTTPError(http.StatusBadRequest, "invalid name or password", nil)},
				{"", "123456", rest.NewHTTPError(http.StatusBadRequest, "invalid name or password", nil)},
			}

			// 遍历测试用例
			for _, tc := range testCases {
				convey.Convey(fmt.Sprintf("登录失败，用户名: %s, 密码: %s", tc.name, tc.passwd), func() {
					id, _, err := user.Login(tc.name, tc.passwd)

					assert.Equal(t, "", id)
					assert.Equal(t, tc.expectedErr, err)
				})
			}
		})
		convey.Convey("数据库错误, GetUserInfoByName failed", func() {
			dbUser.EXPECT().GetUserInfoByName(gomock.Any()).Return(nil, false, errors.New("database error"))

			id, _, err := user.Login("tom", "123456")

			assert.Equal(t, "", id)
			assert.Equal(t, rest.InternalServerError, err)
		})
		convey.Convey("用户名不存在", func() {
			dbUser.EXPECT().GetUserInfoByName(gomock.Any()).Return(nil, false, nil)

			id, _, err := user.Login("nonexistent", "123456")

			assert.Equal(t, "", id)
			assert.Equal(t, rest.NewHTTPError(http.StatusBadRequest, "invalid name or password", nil), err)
		})
		convey.Convey("密码错误", func() {
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)

			dbUser.EXPECT().GetUserInfoByName(gomock.Any()).Return(&interfaces.User{ID: "1", Name: "tom", Password: string(hashedPassword)}, true, nil)

			id, _, err := user.Login("tom", "wrongpassword")

			assert.Equal(t, "", id)
			assert.Equal(t, rest.NewHTTPError(http.StatusBadRequest, "invalid name or password", nil), err)
		})
		convey.Convey("数据库错误, UpdateLoginTime failed", func() {
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
			dbUser.EXPECT().GetUserInfoByName(gomock.Any()).Return(&interfaces.User{ID: "1", Name: "tom", Password: string(hashedPassword)}, true, nil)
			dbUser.EXPECT().UpdateLoginTime(gomock.Any()).Return(errors.New("database error"))

			id, _, err := user.Login("tom", "123456")

			assert.Equal(t, "", id)
			assert.Equal(t, rest.InternalServerError, err)
		})
		convey.Convey("jwt错误, Sign failed", func() {
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
			dbUser.EXPECT().GetUserInfoByName(gomock.Any()).Return(&interfaces.User{ID: "1", Name: "tom", Password: string(hashedPassword)}, true, nil)
			dbUser.EXPECT().UpdateLoginTime(gomock.Any()).Return(nil)
			jwt.EXPECT().Sign(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", errors.New("jwt error"))

			_, _, err := user.Login("tom", "123456")

			assert.Equal(t, rest.InternalServerError, err)
		})
		convey.Convey("登录成功", func() {
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
			dbUser.EXPECT().GetUserInfoByName(gomock.Any()).Return(&interfaces.User{ID: "1", Name: "tom", Password: string(hashedPassword)}, true, nil)
			dbUser.EXPECT().UpdateLoginTime(gomock.Any()).Return(nil)
			jwt.EXPECT().Sign(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("jwtToken", nil)

			id, jwtToken, err := user.Login("tom", "123456")

			assert.Equal(t, "1", id)
			assert.Equal(t, nil, err)
			assert.Equal(t, "jwtToken", jwtToken)
		})
	})
}

func TestGetUserInfo(t *testing.T) {
	convey.Convey("GetUserInfo", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		db, _, err := sqlmock.New()
		assert.Equal(t, nil, err)
		defer db.Close()

		jwt := mock.NewMockLogicsJWT(ctrl)
		dbUser := mock.NewMockDBUser(ctrl)
		user := setupUserTest(db, jwt, nil, dbUser)

		claims := map[string]interface{}{"id": "id"}
		ctx := context.WithValue(context.WithValue(context.Background(), interfaces.TokenKey, "jwtToken"), interfaces.ClaimsKey, claims)
		convey.Convey("数据库错误", func() {
			dbUser.EXPECT().GetUserInfoByID(gomock.Any()).Return(nil, false, errors.New("database error"))

			_, err := user.GetUserInfo(ctx, "id")

			assert.Equal(t, rest.InternalServerError, err)
		})
		convey.Convey("用户不存在", func() {
			dbUser.EXPECT().GetUserInfoByID(gomock.Any()).Return(nil, false, nil)

			_, err := user.GetUserInfo(ctx, "id")

			assert.Equal(t, rest.NotFound, err)
		})
		convey.Convey("获取用户信息成功", func() {
			dbUser.EXPECT().GetUserInfoByID(gomock.Any()).Return(&interfaces.User{ID: "1", Name: "tom"}, true, nil)

			userInfo, err := user.GetUserInfo(ctx, "id")

			assert.Equal(t, nil, err)
			assert.Equal(t, "1", userInfo.ID)
			assert.Equal(t, "tom", userInfo.Name)
		})
	})
}
