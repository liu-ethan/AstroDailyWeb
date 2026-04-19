package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secret []byte
	issuer string
	ttl    time.Duration
}

// NewJWTManager 创建 JWT 管理器。
// 参数：secret - 签名密钥；issuer - 签发者；ttl - 令牌有效期。
// 返回：*JWTManager - JWT 管理器实例。
func NewJWTManager(secret, issuer string, ttl time.Duration) *JWTManager {
	return &JWTManager{secret: []byte(secret), issuer: issuer, ttl: ttl}
}

// Generate 生成指定用户的 JWT。
// 参数：userID - 用户 ID。
// 返回：string - JWT 字符串；time.Time - 过期时间；error - 生成失败错误。
func (m *JWTManager) Generate(userID int64) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(m.ttl)
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, exp, nil
}

// Parse 校验并解析 JWT。
// 参数：tokenString - 待解析的 JWT 字符串。
// 返回：*Claims - 解析后的声明；error - 校验或解析失败错误。
func (m *JWTManager) Parse(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
