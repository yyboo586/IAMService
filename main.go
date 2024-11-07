package main

import (
	"ServiceA/dbaccess"
	"ServiceA/driveradapters"
	"ServiceA/interfaces"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/yyboo586/utils/db"
	myjwt "github.com/yyboo586/utils/myJWT"
	"github.com/yyboo586/utils/rest"
)

var (
	err        error
	privateKey *rsa.PrivateKey
	casb       *casbin.Enforcer
)

func init() {
	// 解码 PEM 格式的私钥
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		log.Fatal("failed to decode PEM block containing the private key")
	}

	// 解析私钥
	privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		log.Fatal("failed to parse private key:", err)
	}
}

type Server struct {
	userHandler interfaces.RESTHandler
}

func (s *Server) Start() {
	go func() {
		engine := gin.Default()
		engine.Use(permissionCheck())

		s.userHandler.RegisterPublic(engine)

		if err := engine.Run(":10001"); err != nil {
			panic(err)
		}
	}()
}

func main() {
	dbPool, err := db.NewDB("root", "12345678", "localhost", 3306, "ServiceA")
	if err != nil {
		panic(err)
	}

	casb, err = casbin.NewEnforcer("model.conf", "policy.csv")
	if err != nil {
		panic(err)
	}

	policies, _ := casb.GetPolicy()
	log.Println("All Policies: ", policies)

	// 依赖注入
	dbaccess.SetDBPool(dbPool)

	s := &Server{
		userHandler: driveradapters.NewUserHandler(),
	}

	s.Start()

	select {}
}

func permissionCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method
		needAuth := false

		// 检查策略，以确定当前URI是否需要鉴权
		policies, err := casb.GetPolicy()
		if err != nil {
			rest.ReplyError(c, err)
			c.Abort()
			return
		}
		var matchedPolicies [][]string
		for _, policy := range policies {
			if policy[3] == "true" && policy[2] == method && matchKey2(policy[1], path) {
				needAuth = true
				matchedPolicies = append(matchedPolicies, policy)
			}
		}
		log.Printf("DEBUG: Matched Policies: %v\n", matchedPolicies)

		// 如果当前URI不需要鉴权，可以访问
		if !needAuth {
			c.Next()
			return
		}

		tokenInfo := strings.Split(c.GetHeader("Authorization"), " ")
		if len(tokenInfo) < 2 {
			rest.ReplyError(c, rest.NewHTTPError(http.StatusUnauthorized, "token is invalid", nil))
			c.Abort()
			return
		}
		extClaims, err := myjwt.Verify(tokenInfo[1], privateKey)
		if err != nil {
			rest.ReplyError(c, rest.NewHTTPError(http.StatusUnauthorized, "token is invalid", nil))
			c.Abort()
			return
		}
		// log.Printf("DEBUG: ExtClaims: %v, err: %v", extClaims, err)

		// 检查角色权限
		allowed, err := casb.Enforce(extClaims["name"].(string), path, method, needAuth)
		if err != nil {
			rest.ReplyError(c, rest.NewHTTPError(http.StatusInternalServerError, err.Error(), nil))
			c.Abort()
			return
		}
		if !allowed {
			rest.ReplyError(c, rest.NewHTTPError(http.StatusForbidden, "no permission", nil))
			c.Abort()
			return
		}

		// 将令牌信息放入上下文
		ctx := context.WithValue(context.WithValue(c.Request.Context(), interfaces.ClaimsKey, extClaims), interfaces.TokenKey, tokenInfo[1])
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func matchKey2(pattern, url string) bool {
	pArray := strings.Split(pattern, "/")
	uArray := strings.Split(url, "/")

	if len(pArray) != len(uArray) {
		return false
	}

	// 遍历每个部分进行比较
	for i := range pArray {
		if pArray[i] != uArray[i] && !strings.HasPrefix(pArray[i], ":") {
			fmt.Println(pArray[i], uArray[i])
			return false // 直接不匹配且不是动态参数
		}
	}

	return true
}

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
-----END RSA PRIVATE KEY-----`
