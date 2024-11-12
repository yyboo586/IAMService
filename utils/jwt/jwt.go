package jwt

import (
	"crypto/rsa"
	"time"

	gojose "github.com/go-jose/go-jose/v4"
	gojoseJwt "github.com/go-jose/go-jose/v4/jwt"
)

type CustomClaims struct {
	gojoseJwt.Claims

	ExtClaims map[string]interface{}
}

func Sign(userID string, claims map[string]interface{}, privateKey *rsa.PrivateKey) (jwtTokenStr string, err error) {
	signer, err := gojose.NewSigner(gojose.SigningKey{Key: privateKey, Algorithm: gojose.RS256}, (&gojose.SignerOptions{}).WithType("JWT"))
	if err != nil {
		return
	}

	if claims == nil {
		claims = make(map[string]interface{})
	}

	cclaims := CustomClaims{
		gojoseJwt.Claims{
			Issuer:    "example.com",
			Subject:   userID,
			Audience:  []string{"UserManagement"},
			Expiry:    gojoseJwt.NewNumericDate(time.Now().Add(time.Hour * 1)),
			IssuedAt:  gojoseJwt.NewNumericDate(time.Now()),
			NotBefore: gojoseJwt.NewNumericDate(time.Now()),
		},
		claims,
	}
	jwtTokenStr, err = gojoseJwt.Signed(signer).Claims(cclaims).Serialize()
	if err != nil {
		return
	}

	return jwtTokenStr, nil
}

func Verify(jwtTokenStr string, privateKey *rsa.PrivateKey) (extClaims map[string]interface{}, err error) {
	jwtToken, err := gojoseJwt.ParseSigned(jwtTokenStr, []gojose.SignatureAlgorithm{gojose.RS256})
	if err != nil {
		return
	}

	var cclaims CustomClaims
	if err = jwtToken.Claims(&privateKey.PublicKey, &cclaims); err != nil {
		return
	}
	expected := gojoseJwt.Expected{
		Issuer:      "example.com",
		AnyAudience: gojoseJwt.Audience{"UserManagement"},
		Time:        time.Time{},
	}
	if err = cclaims.Claims.Validate(expected); err != nil {
		return
	}

	return cclaims.ExtClaims, err
}
