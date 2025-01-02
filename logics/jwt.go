package logics

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/gofrs/uuid"
	"github.com/yyboo586/IAMService/dbaccess"
	"github.com/yyboo586/IAMService/interfaces"
	"github.com/yyboo586/common/jwtUtils"
	"github.com/yyboo586/common/logUtils"
	"github.com/yyboo586/common/rest"
)

var (
	logicsJWTOnce sync.Once
	lJWT          interfaces.LogicsJWT
)

type logicsJWT struct {
	dbJWT interfaces.DBJWT

	cacheLock sync.RWMutex
	keysCache map[string]*jose.JSONWebKeySet

	logger *logUtils.Logger
}

func NewLogicsJWT() interfaces.LogicsJWT {
	logicsJWTOnce.Do(func() {
		lJWT = &logicsJWT{
			dbJWT:     dbaccess.NewDBJWT(),
			keysCache: make(map[string]*jose.JSONWebKeySet),
			logger:    loggerInstance,
		}
	})

	return lJWT
}

func (j *logicsJWT) Sign(userID string, claims map[string]interface{}, setID, alg string) (jwtTokenStr string, err error) {
	key, err := j.loadORGenerateKeys(setID, alg)
	if err != nil {
		j.logger.Error(err)
		return "", err
	}

	signer, err := jose.NewSigner(jose.SigningKey{Key: key.Key, Algorithm: jose.SignatureAlgorithm(alg)}, (&jose.SignerOptions{}).WithType("JWT").WithHeader("kid", key.KeyID))
	if err != nil {
		j.logger.Error(err)
		return
	}

	if claims == nil {
		claims = make(map[string]interface{})
	}

	cclaims := interfaces.CustomClaims{
		Claims: jwt.Claims{
			Issuer:    "IAMService.com",
			Subject:   userID,
			Audience:  []string{"IAMService"},
			Expiry:    jwt.NewNumericDate(time.Now().Add(time.Hour * 1)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ID:        uuid.Must(uuid.NewV4()).String(),
		},
		ExtClaims: claims,
	}
	jwtTokenStr, err = jwt.Signed(signer).Claims(cclaims).Serialize()
	if err != nil {
		j.logger.Error(err)
		return
	}

	return jwtTokenStr, nil
}

func (j *logicsJWT) Verify(jwtTokenStr string) (claims *interfaces.CustomClaims, err error) {
	kid, err := getKid(jwtTokenStr)
	if err != nil {
		j.logger.Error(err)
		return
	}
	key, err := j.getKey(kid)
	if err != nil {
		j.logger.Error(err)
		return nil, err
	}

	jwtToken, err := jwt.ParseSigned(jwtTokenStr, []jose.SignatureAlgorithm{jose.SignatureAlgorithm(key.Algorithm)})
	if err != nil {
		j.logger.Error(err)
		return nil, rest.NewHTTPError(http.StatusUnauthorized, "token invalid", nil)
	}

	claims = &interfaces.CustomClaims{}
	switch key.Algorithm {
	case "HS256", "HS384", "HS512":
		err = jwtToken.Claims(key.Key, &claims)
	default:
		err = jwtToken.Claims(key.Public(), &claims)
	}
	if err != nil {
		j.logger.Error(err)
		return nil, rest.NewHTTPError(http.StatusUnauthorized, "token invalid", nil)
	}
	expected := jwt.Expected{
		Issuer:      "IAMService.com",
		AnyAudience: jwt.Audience{"IAMService"},
		Time:        time.Time{},
	}
	if err = claims.Claims.Validate(expected); err != nil {
		j.logger.Info(err)
		return nil, rest.NewHTTPError(http.StatusUnauthorized, "token invalid", nil)
	}

	exists, err := j.dbJWT.GetBlacklist(claims.Claims.ID)
	if err != nil {
		j.logger.Error(err)
		return nil, err
	}
	if exists {
		return nil, rest.NewHTTPError(http.StatusUnauthorized, "token revoked", nil)
	}

	return claims, err
}

func (j *logicsJWT) RevokeToken(jwtTokenStr string) (err error) {
	claims, err := j.Verify(jwtTokenStr)
	if err != nil {
		if e, ok := err.(*rest.HTTPError); ok {
			if e.StatusCode() == http.StatusUnauthorized {
				return nil
			}
		}
		return err
	}

	if err = j.dbJWT.AddBlacklist(claims.ID); err != nil {
		j.logger.Error(err)
		return err
	}

	return nil
}

func (j *logicsJWT) loadORGenerateKeys(setID, alg string) (*jose.JSONWebKey, error) {
	j.cacheLock.RLock()
	if kSet, ok := j.keysCache[setID]; ok {
		key := kSet.Keys[0]
		j.cacheLock.RUnlock()
		return &key, nil
	}
	j.cacheLock.RUnlock()

	j.cacheLock.Lock()
	defer j.cacheLock.Unlock()

	if kSet, ok := j.keysCache[setID]; ok {
		return &kSet.Keys[0], nil
	}

	kSet, err := j.dbJWT.GetKeySet(setID)
	if err != nil {
		return nil, fmt.Errorf("get keyset failed: %w", err)
	}
	if len(kSet.Keys) == 0 {
		kid := uuid.Must(uuid.NewV4()).String()
		kSet, err = j.generateAndPersistJWKSet(setID, alg, kid, "sig")
		if err != nil {
			return nil, fmt.Errorf("generateAndPersistJWKSet failed: %w", err)
		}
	}

	j.keysCache[setID] = kSet

	key := kSet.Keys[0]
	return &key, nil
}

func (j *logicsJWT) generateAndPersistJWKSet(setID, alg, kid, use string) (kSet *jose.JSONWebKeySet, err error) {
	if len(kid) == 0 {
		kid = uuid.Must(uuid.NewV4()).String()
	}
	if len(use) == 0 {
		use = "sig"
	}

	var key interface{}
	switch jose.SignatureAlgorithm(alg) {
	case jose.HS256, jose.HS384, jose.HS512:
		if key, err = jwtUtils.NewSymmetricKey(jose.SignatureAlgorithm(alg)); err != nil {
			return nil, err
		}
	default:
		if _, key, err = jwtUtils.NewAsymmetricKey(jose.SignatureAlgorithm(alg)); err != nil {
			return nil, err
		}
	}

	kSet = &jose.JSONWebKeySet{
		Keys: []jose.JSONWebKey{
			{
				Algorithm:                   string(alg),
				Key:                         key,
				KeyID:                       kid,
				Use:                         use,
				Certificates:                []*x509.Certificate{},
				CertificateThumbprintSHA256: []byte{},
				CertificateThumbprintSHA1:   []byte{},
			},
		},
	}

	if err = j.dbJWT.AddKeySet(setID, kSet); err != nil {
		return nil, err
	}
	return kSet, nil
}

func getKid(jwtTokenStr string) (alg string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("GetKid error: %v", r)
		}
	}()

	headerBase64 := strings.Split(jwtTokenStr, ".")[0]

	headerBytes, err := base64.RawURLEncoding.DecodeString(headerBase64)
	if err != nil {
		return "", fmt.Errorf("decode header failed: %w", err)
	}

	var header map[string]interface{}
	if err = json.Unmarshal(headerBytes, &header); err != nil {
		return "", fmt.Errorf("unmarshal header failed: %w", err)
	}

	return header["kid"].(string), nil
}

func (j *logicsJWT) getKey(kid string) (*jose.JSONWebKey, error) {
	j.cacheLock.Lock()
	defer j.cacheLock.Unlock()

	for _, v := range j.keysCache {
		for _, key := range v.Keys {
			if key.KeyID == kid {
				return &key, nil
			}
		}
	}

	key, err := j.dbJWT.GetKey(kid)
	if err != nil {
		return nil, fmt.Errorf("get key failed: %w", err)
	}

	return key, nil
}

func (j *logicsJWT) GetPublicKey(kid string) (*jose.JSONWebKey, error) {
	key, err := j.getKey(kid)
	if err != nil {
		j.logger.Error(err)
		return nil, err
	}

	k, err := getPublic(key)
	if err != nil {
		j.logger.Error(err)
		return nil, err
	}

	return &k, nil
}

func getPublic(k *jose.JSONWebKey) (jose.JSONWebKey, error) {
	ret := *k
	switch key := k.Key.(type) {
	case *ecdsa.PrivateKey:
		ret.Key = key.Public()
	case *rsa.PrivateKey:
		ret.Key = key.Public()
		return ret, nil
	default:
		return jose.JSONWebKey{}, fmt.Errorf("unsupported key type")
	}

	return ret, nil
}
