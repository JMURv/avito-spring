package auth

import (
	"context"
	"github.com/JMURv/avito-spring/internal/config"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"time"
)

const issuer = "avito-spring"
const tokenDuration = time.Hour * 2

type Core interface {
	Hash(val string) (string, error)
	ComparePasswords(hashed, pswd []byte) error
	NewToken(uid uuid.UUID, role string) (string, error)
	ParseClaims(ctx context.Context, tokenStr string) (Claims, error)
}

type Claims struct {
	UID  uuid.UUID `json:"uid"`
	Role string    `json:"roles"`
	jwt.RegisteredClaims
}

type Auth struct {
	secret []byte
}

func New(conf config.Config) *Auth {
	return &Auth{
		secret: []byte(conf.Secret),
	}
}

func (a *Auth) Hash(val string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(val), bcrypt.MinCost)
	if err != nil {
		zap.L().Error(
			"Failed to generate hash",
			zap.String("val", val),
			zap.Error(err),
		)
		return "", err
	}
	return string(bytes), nil
}

func (a *Auth) ComparePasswords(hashed, pswd []byte) error {
	if err := bcrypt.CompareHashAndPassword(hashed, pswd); err != nil {
		return ErrInvalidCredentials
	}
	return nil
}

func (a *Auth) NewToken(uid uuid.UUID, role string) (string, error) {
	signed, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256, &Claims{
			UID:  uid,
			Role: role,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenDuration)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Issuer:    issuer,
			},
		},
	).SignedString(a.secret)

	if err != nil {
		zap.L().Error(
			ErrWhileCreatingToken.Error(),
			zap.String("role", role),
			zap.Error(err),
		)
		return "", ErrWhileCreatingToken
	}

	return signed, nil
}

func (a *Auth) ParseClaims(_ context.Context, tokenStr string) (Claims, error) {
	claims := Claims{}
	token, err := jwt.ParseWithClaims(
		tokenStr, &claims, func(token *jwt.Token) (any, error) {
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, ErrUnexpectedSignMethod
			}
			return a.secret, nil
		},
	)

	if err != nil {
		zap.L().Error(
			"Failed to parse claims",
			zap.String("token", tokenStr),
			zap.String("alg", token.Method.Alg()),
			zap.Error(err),
		)
		return claims, err
	}

	if !token.Valid {
		zap.L().Debug(
			"Token is invalid",
			zap.String("token", tokenStr),
		)
		return claims, ErrInvalidToken
	}

	return claims, nil
}
