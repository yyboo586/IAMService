package driveradapters

// import (
// 	"context"
// 	"log"

// 	"UserManagement/interfaces"
// 	"UserManagement/interfaces/mock"
// 	jwtUtils "UserManagement/utils/jwt"
// 	rsaUtils "UserManagement/utils/rsa"

// 	"bytes"
// 	"encoding/json"
// 	"errors"
// 	"io"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/gin-gonic/gin"
// 	"github.com/golang/mock/gomock"
// 	"github.com/stretchr/testify/assert"

// 	. "github.com/smartystreets/goconvey/convey"
// )

// func newUserHandler(logicsUser interfaces.LogicsUser) *UserHandler {
// 	return &UserHandler{
// 		logicsUser: logicsUser,
// 	}
// }

// func generateJWTToken(userID string, claims map[string]interface{}) (jwtTokenStr string) {
// 	privateKey, err := rsaUtils.LoadPrivateKey()
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	jwtTokenStr, _ = jwtUtils.Sign(userID, claims, privateKey)

// 	return
// }

// func TestCreate(t *testing.T) {
// 	Convey("Create", t, func() {
// 		ctrl := gomock.NewController(t)
// 		defer ctrl.Finish()

// 		mockLogicsUser := mock.NewMockLogicsUser(ctrl)
// 		userHandler := newUserHandler(mockLogicsUser)

// 		gin.SetMode(gin.TestMode)
// 		engine := gin.Default()

// 		userHandler.RegisterPublic(engine)

// 		user := &interfaces.User{
// 			Name:     "tom",
// 			Password: "123456",
// 		}
// 		data, _ := json.Marshal(user)
// 		req, _ := http.NewRequest(http.MethodPost, "/api/v1/user-management/users", bytes.NewBuffer(data))
// 		jwtTokenStr := generateJWTToken("user_id", nil)
// 		req.Header.Set("Content-Type", "application/json")
// 		req.Header.Set("Authorization", "Bearer "+jwtTokenStr)

// 		ctx := context.Background()
// 		Convey("Fail", func() {
// 			mockLogicsUser.EXPECT().Create(ctx, user).Return(errors.New("internal error"))

// 			w := httptest.NewRecorder()

// 			engine.ServeHTTP(w, req)

// 			assert.Equal(t, http.StatusInternalServerError, w.Code)
// 		})

// 		Convey("Success", func() {
// 			mockLogicsUser.EXPECT().Create(ctx, user).Return(nil)

// 			w := httptest.NewRecorder()

// 			engine.ServeHTTP(w, req)

// 			assert.Equal(t, http.StatusCreated, w.Code)
// 		})
// 	})
// }

// func TestLogin(t *testing.T) {
// 	Convey("Login", t, func() {
// 		ctrl := gomock.NewController(t)
// 		defer ctrl.Finish()

// 		mockLogicsUser := mock.NewMockLogicsUser(ctrl)
// 		userHandler := newUserHandler(mockLogicsUser)

// 		gin.SetMode(gin.TestMode)
// 		engine := gin.Default()

// 		userHandler.RegisterPublic(engine)

// 		user := &interfaces.User{
// 			Name:     "tom",
// 			Password: "123456",
// 		}
// 		data, _ := json.Marshal(user)
// 		req, _ := http.NewRequest(http.MethodPost, "/api/v1/user-management/user-login", bytes.NewBuffer(data))
// 		req.Header.Set("Content-Type", "application/json")

// 		Convey("Failed", func() {
// 			mockLogicsUser.EXPECT().Login(user.Name, user.Password).Return("", "", errors.New("internal error"))

// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			data, _ := io.ReadAll(w.Result().Body)

// 			assert.Equal(t, http.StatusInternalServerError, w.Code)
// 			assert.Equal(t, "internal error", string(data))
// 		})

// 		Convey("Success", func() {
// 			mockLogicsUser.EXPECT().Login(user.Name, user.Password).Return("id", "", nil)

// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			var i interface{}
// 			data, _ := io.ReadAll(w.Result().Body)
// 			json.Unmarshal(data, &i)

// 			assert.Equal(t, http.StatusOK, w.Code)
// 			assert.Equal(t, "id", i.(map[string]interface{})["id"].(string))
// 		})
// 	})
// }

// func TestGetUserInfo(t *testing.T) {
// 	Convey("GetUserInfo", t, func() {
// 		ctrl := gomock.NewController(t)
// 		defer ctrl.Finish()

// 		mockLogicsUser := mock.NewMockLogicsUser(ctrl)
// 		userHandler := newUserHandler(mockLogicsUser)

// 		gin.SetMode(gin.TestMode)
// 		engine := gin.Default()

// 		userHandler.RegisterPublic(engine)

// 		req, _ := http.NewRequest(http.MethodGet, "/api/v1/user-management/users/123", nil)
// 		jwtTokenStr := generateJWTToken("user_id", nil)
// 		req.Header.Set("Content-Type", "application/json")
// 		req.Header.Set("Authorization", "Bearer "+jwtTokenStr)

// 		Convey("Failed", func() {
// 			mockLogicsUser.EXPECT().GetUserInfo(gomock.Any(), "123").Return(nil, errors.New("internal error"))

// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			data, _ := io.ReadAll(w.Result().Body)

// 			assert.Equal(t, http.StatusInternalServerError, w.Code)
// 			assert.Equal(t, "internal error", string(data))
// 		})

// 		Convey("Success", func() {
// 			userInfo := &interfaces.User{
// 				ID:   "user_id",
// 				Name: "tom",
// 			}
// 			mockLogicsUser.EXPECT().GetUserInfo(gomock.Any(), "123").Return(userInfo, nil)

// 			w := httptest.NewRecorder()
// 			engine.ServeHTTP(w, req)

// 			var i interface{}
// 			data, _ := io.ReadAll(w.Result().Body)
// 			json.Unmarshal(data, &i)

// 			assert.Equal(t, http.StatusOK, w.Code)
// 			assert.Equal(t, userInfo.ID, i.(map[string]interface{})["id"].(string))
// 			assert.Equal(t, userInfo.Name, i.(map[string]interface{})["name"].(string))
// 		})
// 	})
// }
