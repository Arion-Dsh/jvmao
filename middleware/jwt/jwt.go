package jwt

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/arion-dsh/jvmao"
	"github.com/dgrijalva/jwt-go"
)

var theJWTManager *JWTManager

type JWTMiddlewareOption struct {
	Secret        string
	TokenDuration time.Duration
}

func NewJwtMiddleware(opt JWTMiddlewareOption) jvmao.MiddlewareFunc {
	theJWTManager = NewJWTManager(opt.Secret, opt.TokenDuration)
	return JWTMiddleware
}

func JWTMiddleware(h jvmao.HandlerFunc) jvmao.HandlerFunc {
	return func(ctx jvmao.Context) error {

		token := ctx.HanderValue("authorization")
		if token == "" {
			return ctx.Error(401, errors.New("Unauthorized"))
		}
		token = strings.TrimPrefix(token, "Bearer ")

		claims, err := theJWTManager.Verify(token)
		if err != nil {
			return ctx.Error(401, errors.New("Unauthorized"))
		}

		ctx.Set("jwt", claims)

		return h(ctx)
	}
}

type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
}

func NewJWTManager(secretKey string, tokenDuration time.Duration) *JWTManager {
	jwt := &JWTManager{secretKey, tokenDuration}
	if theJWTManager == nil {
		theJWTManager = &JWTManager{secretKey, tokenDuration}
	}
	return jwt

}

type defaultClaims struct {
	jwt.StandardClaims
	Data interface{} `json:"data"`
}

func Generate(data interface{}) (string, error) {
	if theJWTManager == nil {
		return "", errors.New("jwt manager not initialized")
	}
	return theJWTManager.Generate(data)
}
func GenerateWithSecret(data interface{}, s string) (string, error) {
	if theJWTManager == nil {
		return "", errors.New("jwt manager not initialized")
	}
	return theJWTManager.GenerateWithSecret(data, s)
}

func (m *JWTManager) GenerateWithSecret(data interface{}, s string) (string, error) {
	return m.generateWithSecret(data, s)
}

func (m *JWTManager) Generate(data interface{}) (string, error) {
	return m.generateWithSecret(data, m.secretKey)
}

func (m *JWTManager) generateWithSecret(data interface{}, s string) (string, error) {
	claims := defaultClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(theJWTManager.tokenDuration).Unix(),
		},
		Data: data,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s))
}

func Verify(accessToken string) (interface{}, error) {
	if theJWTManager == nil {
		return nil, errors.New("jwt manager not initialized")
	}
	return theJWTManager.Verify(accessToken)
}

func VerifyWithSecret(tk, s string) (interface{}, error) {
	if theJWTManager == nil {
		return nil, errors.New("jwt manager not initialized")
	}
	return theJWTManager.verifyWithSecret(tk, s)

}

func (m *JWTManager) Verify(accessToken string) (interface{}, error) {
	return m.verifyWithSecret(accessToken, m.secretKey)
}

func (m *JWTManager) verifyWithSecret(tk, s string) (interface{}, error) {
	token, err := jwt.ParseWithClaims(
		tk,
		&defaultClaims{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, fmt.Errorf("unexpected token signing method")
			}
			return []byte(s), nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)

	}

	claims, ok := token.Claims.(*defaultClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")

	}

	return claims.Data, nil

}
