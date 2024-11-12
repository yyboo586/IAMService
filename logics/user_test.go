package logics

import (
	"UserManagement/interfaces"
	"UserManagement/interfaces/mock"
	jwtUtils "UserManagement/utils/jwt"
	"UserManagement/utils/rest"
	errUtils "UserManagement/utils/rest/errors"

	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

const privateKeyPEM = `-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEAqi4Ej6S+ePKh580Q4WKYajWwC54dRGGmu9i2dLb5QIvgyL0Z
kIBgBJnIqzhgEtdCwU3W6IWW+PkKu8mI08KHfcf+pCT6kRGYG+Xm+ZIVq0HcvFzI
lxSn0Mvc/N6kk2swdqktUXV/Pwau2atj10tlSmn/HKO/W47FWxBwC3O/gK3c7fOE
x9WsTS/Pt+d02gKvLQ2Bialfca9z9yG3YK28FJu3j6bA2BuFMTqQl/T1jAeyYZTb
tG7p3K2cgzrCRfi/PB7e9S0Lqc/yBs2cd3dV6f3+Ve7c2Y5/8znwFxzgMZ8b7VGW
hR5vgSLSSzyAU5WmgUt5m98qCc0oPikuQzey3wIDAQABAoIBAAvUnySNQ2CNHYxL
yTyh6g6YJODp4Qb78udkLWr3vWQrVTkfTEOraQFo33ZnuOYWaOGfU61efBxa09Ay
NnziLSElYiJvH6wuGPD3jpMTAMajEYFWwese2Hu/cGFz6OUGspvNLwVWsb3j7Qvc
ylgROb1umPmYuJjY2Ad4oRFqvolnb8avLEtXfrlgNf1YCWuhWb9pka2KDAfhmXH6
bXCv34CxgRj9DyGmplPWFuzH4cqICHsztfHdbbe2f1zfoOQUkmF9SrYgVEHPl1Fb
3yQ12lw43CNHqzlh7Bd9OQ7SdPGzCMA+4oWLeGqHOQS3sVj7B4Izq344iXz6Hq6F
sR0ZcyECgYEAyaJcbHGxITffB7tjge8O8aiIxAwiGnpTMuDBYVB8ni6VoC7OJVN8
ogFsP3CMFBG8k8r7AjV+vKZOBf8KfRoRy0mlZ61yQSmggfgZ7fTGmmN+MsHrSZ6B
khxqDPfIPQwLJYBT/V/PueQR5OdxMNJut6jlhzY7QsWphg1I92GaeF0CgYEA2BCJ
qrzehv0Z+cuhAx59qiVCYAp3rD8xBqyFrcTS7qrbD9I6XufdJZNZvKbqMnrdfidV
ru2N+wbv71TSONN9ZyHqu6O1tmCAvKWPd6oxJmTUSMLGTNJ5LgpuwsEvHaGlWyh1
sEWPjfGYe3DKS6VRKP0UfTELiI7H84R3DD6+tGsCgYEAyCOTr8SN+BX4GDmlRMSg
RbhuwIH2m+eNi7PR3yFAANbmh8/NqPkcfcYBx1qUgBs23lAdJI0q1mAQlB0aMSDe
RrU8LBPak9mYy0kTm8FaHMbi7cjUHgfqPrhbf7G3HPlGWxvswlQG4VIDfP1Juhc1
9LD92182JUoDwd6P7ZUA+bUCgYEAoPfJKGtnOZgslv4OqZ04r97sUVLbD3dQlhFH
0krVfrvJUkMj+3qwNgNOEo8j4ZHJm+fAHP+cDE2ByYMezvk47vHEyCBSC1pf7qtF
dDhWP61UvhRl2evgHd3l4LA94sx/vacp7rYUGgLIwAYqoCq8iVXqws4cMpN1AcZJ
TtUcDJsCgYEAnc8aEhq7VxH667JWS7gf9TjpGcnKsfRULPF1CqPSJ0UFZk0/GUmN
EfJeqKPImwBnKOsnYuQ4rYnKcpoXGd3If9JetRr+VHU9JJHDeaR7QUmFXvVKWqDT
ye4qXPbwsqFoz5DkI8rUvIFw/L+efvczC3v3sq1CQ3Jdlj6Vo3xd8j0=
-----END RSA PRIVATE KEY-----
`

func newUser(dbUser interfaces.DBUser) *user {
	// 解码 PEM 格式的私钥
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		log.Fatal("failed to decode PEM block containing the private key")
	}

	// 解析私钥
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		log.Fatal("failed to parse private key:", err)
	}

	return &user{
		pwdRegex:   regexp.MustCompile(`^[a-zA-Z0-9]{6,12}$`),
		nameRegex:  regexp.MustCompile(`^[\p{Han}a-zA-Z0-9]{1,6}$`),
		privateKey: privateKey,
		dbUser:     dbUser,
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
				{"", errUtils.NewHTTPError(http.StatusBadRequest, "invalid name", nil)},
				{"test user", errUtils.NewHTTPError(http.StatusBadRequest, "invalid name", nil)},
				{"test&user", errUtils.NewHTTPError(http.StatusBadRequest, "invalid name", nil)},
				{"testuser", errUtils.NewHTTPError(http.StatusBadRequest, "invalid name", nil)},
				{"七个中文字符啊", errUtils.NewHTTPError(http.StatusBadRequest, "invalid name", nil)},
			}

			for _, tc := range testCases {
				Convey(fmt.Sprintf("创建失败，用户名: %s", tc.name), func() {
					invalidUser := &interfaces.User{
						Name:     tc.name,
						Password: "TestPass123",
					}

					err := user.Create(context.Background(), invalidUser)

					assert.Equal(t, tc.expectedErr, err)
				})
			}
		})

		Convey("密码校验", func() {
			testCases := []struct {
				password    string
				expectedErr error
			}{
				{"", errUtils.NewHTTPError(http.StatusBadRequest, "invalid password", nil)},
				{"TestPass!@#", errUtils.NewHTTPError(http.StatusBadRequest, "invalid password", nil)},
				{"12345", errUtils.NewHTTPError(http.StatusBadRequest, "invalid password", nil)},
				{"1234567890123", errUtils.NewHTTPError(http.StatusBadRequest, "invalid password", nil)},
			}

			for _, tc := range testCases {
				Convey(fmt.Sprintf("创建失败，用户密码: %s", tc.password), func() {
					invalidUser := &interfaces.User{
						Name:     "tom",
						Password: tc.password,
					}

					err := user.Create(context.Background(), invalidUser)

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

			err := user.Create(context.Background(), validUser)

			assert.Equal(t, errUtils.NewHTTPError(http.StatusInternalServerError, "database error", nil), err)
		})

		Convey("用户名已存在", func() {
			validUser := &interfaces.User{
				Name:     "tom",
				Password: "123456",
			}

			dbUser.EXPECT().FetchByName(gomock.Any()).Return(nil, true, nil)

			err := user.Create(context.Background(), validUser)

			assert.Equal(t, errUtils.NewHTTPError(http.StatusConflict, "name already exists", nil), err)
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

				err := user.Create(context.Background(), validUser)

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
				{"", "", errUtils.NewHTTPError(http.StatusBadRequest, "invalid name or password", nil)},
				{"tom", "", errUtils.NewHTTPError(http.StatusBadRequest, "invalid name or password", nil)},
				{"", "123456", errUtils.NewHTTPError(http.StatusBadRequest, "invalid name or password", nil)},
			}

			for _, tc := range testCases {
				Convey(fmt.Sprintf("登录失败，用户名: %s, 密码: %s", tc.name, tc.passwd), func() {
					id, _, err := user.Login(tc.name, tc.passwd)

					assert.Equal(t, "", id)
					assert.Equal(t, tc.expectedErr, err)
				})
			}
		})

		Convey("数据库错误, FetchByName failed", func() {
			dbUser.EXPECT().FetchByName(gomock.Any()).Return(nil, false, errors.New("database error"))

			id, _, err := user.Login("tom", "123456")

			assert.Equal(t, "", id)
			assert.Equal(t, errUtils.NewHTTPError(http.StatusInternalServerError, "database error", nil), err)
		})

		Convey("用户名不存在", func() {
			dbUser.EXPECT().FetchByName(gomock.Any()).Return(nil, false, nil)

			id, _, err := user.Login("nonexistent", "123456")

			assert.Equal(t, "", id)
			assert.Equal(t, errUtils.NewHTTPError(http.StatusBadRequest, "invalid name or password", nil), err)
		})

		Convey("密码错误", func() {
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)

			dbUser.EXPECT().FetchByName(gomock.Any()).Return(&interfaces.User{ID: "1", Name: "tom", Password: string(hashedPassword)}, true, nil)

			id, _, err := user.Login("tom", "wrongpassword")

			assert.Equal(t, "", id)
			assert.Equal(t, errUtils.NewHTTPError(http.StatusBadRequest, "invalid name or password", nil), err)
		})

		Convey("数据库错误, UpdateLoginTime failed", func() {
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)

			dbUser.EXPECT().FetchByName(gomock.Any()).Return(&interfaces.User{ID: "1", Name: "tom", Password: string(hashedPassword)}, true, nil)
			dbUser.EXPECT().UpdateLoginTime(gomock.Any()).Return(errors.New("database error"))

			id, _, err := user.Login("tom", "123456")

			assert.Equal(t, "", id)
			assert.Equal(t, errUtils.NewHTTPError(http.StatusInternalServerError, "database error", nil), err)
		})

		Convey("登录成功", func() {
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
			dbUser.EXPECT().FetchByName(gomock.Any()).Return(&interfaces.User{ID: "1", Name: "tom", Password: string(hashedPassword)}, true, nil)
			dbUser.EXPECT().UpdateLoginTime(gomock.Any()).Return(nil)

			id, _, err := user.Login("tom", "123456")
			assert.Equal(t, "1", id)
			assert.Equal(t, nil, err)
		})
	})
}

func TestGetUserInfo(t *testing.T) {
	Convey("GetUserInfo", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		dbUser := mock.NewMockDBUser(ctrl)
		user := newUser(dbUser)

		claims := map[string]interface{}{"id": "id"}
		jwtTokenStr, _ := jwtUtils.Sign("id", claims, user.privateKey)
		ctx := context.WithValue(context.WithValue(context.Background(), rest.TokenKey, jwtTokenStr), rest.ClaimsKey, claims)
		Convey("数据库错误", func() {
			dbUser.EXPECT().GetUserInfoByID(gomock.Any()).Return(nil, false, errors.New("database error"))

			_, err := user.GetUserInfo(ctx, "id")

			assert.Equal(t, errUtils.NewHTTPError(http.StatusInternalServerError, "database error", nil), err)
		})

		Convey("用户不存在", func() {
			dbUser.EXPECT().GetUserInfoByID(gomock.Any()).Return(nil, false, nil)

			_, err := user.GetUserInfo(ctx, "id")

			assert.Equal(t, errUtils.NewHTTPError(http.StatusNotFound, "user not found", nil), err)
		})

		Convey("获取用户信息成功", func() {
			dbUser.EXPECT().GetUserInfoByID(gomock.Any()).Return(&interfaces.User{ID: "1", Name: "tom"}, true, nil)

			userInfo, err := user.GetUserInfo(ctx, "id")

			assert.Equal(t, nil, err)
			assert.Equal(t, "1", userInfo.ID)
			assert.Equal(t, "tom", userInfo.Name)
		})
	})
}
