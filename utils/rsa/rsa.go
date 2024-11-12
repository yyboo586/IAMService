package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"
	"path/filepath"
)

func GenerateAndSavePrivateKey(filePath string) (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	// 将私钥转换为 PEM 格式
	privateDER := x509.MarshalPKCS1PrivateKey(privateKey)
	privateBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateDER,
	}

	// 将 PEM 格式的私钥写入文件
	err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
	if err != nil {
		return nil, err
	}

	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	err = pem.Encode(file, privateBlock)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func LoadPrivateKey() (*rsa.PrivateKey, error) {
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

	return privateKey, nil
}

const privateKeyPEM = `
-----BEGIN RSA PRIVATE KEY-----
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
