package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/kokkoniemi/texinroistot/config"
	"github.com/kokkoniemi/texinroistot/crypt"
)

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

type JWTClaims struct {
	JWTUserClaims
	Key   string `json:"key,omitempty"`
	KeyIv string `json:"iv,omitempty"`
	jwt.RegisteredClaims
}

type JWTUserClaims struct {
	UserID    string `json:"uid,omitempty"`
	EmailHash string `json:"e,omitempty"`
	EmailIv   string `json:"eiv,omitempty"`
}

func (a *AuthService) CreateAccessToken(sharedToken string, email string) (string, error) {
	accessShared, err := crypt.Encrypt(sharedToken)
	if err != nil {
		return "", err
	}

	signingKey := []byte(config.CookieAccessSecret)
	userHash := crypt.Hash(email)
	encryptedEmail, err := crypt.Encrypt(email)
	if err != nil {
		return "", err
	}

	claims := JWTClaims{
		JWTUserClaims{
			userHash,
			encryptedEmail.GetContent(),
			encryptedEmail.GetIv(),
		},
		accessShared.GetContent(),
		accessShared.GetIv(),
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(signingKey)
}

func (a *AuthService) CreateRefreshToken(sharedKey string) (string, error) {
	refreshShared, err := crypt.Encrypt(sharedKey)
	if err != nil {
		return "", err
	}

	signingKey := []byte(config.CookieRefreshSecret)

	claims := JWTClaims{
		JWTUserClaims{},
		refreshShared.GetContent(),
		refreshShared.GetIv(),
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * 7 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(signingKey)
}

func (a AuthService) VerifyRefreshToken(token string, jwtOpts ...JWTVerifyOptionOverride) (*JWTClaims, error) {
	return a.verifyToken(config.CookieRefreshSecret, token, jwtOpts...)
}

func (a AuthService) VerifyAccessToken(token string, jwtOpts ...JWTVerifyOptionOverride) (*JWTClaims, error) {
	return a.verifyToken(config.CookieAccessSecret, token, jwtOpts...)
}

type JWTVerifyOption struct {
	key interface{}
}

type JWTVerifyOptionOverride func(o *JWTVerifyOption)

var (
	unexpectedSigningMethodError = "unexpected signing method: %v"
)

func (a AuthService) verifyToken(secret string, token string, jwtOpts ...JWTVerifyOptionOverride) (*JWTClaims, error) {
	options := &JWTVerifyOption{[]byte(secret)}

	for _, opt := range jwtOpts {
		opt(options)
	}

	jwtToken, err := jwt.ParseWithClaims(token, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf(unexpectedSigningMethodError, token.Header["alg"])
		}
		return options.key, nil
	})

	if err != nil {
		fmt.Errorf("error occured while parsing jwt token")
		return nil, err
	}

	if claims, ok := jwtToken.Claims.(*JWTClaims); ok && jwtToken.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid jwt token")
}
