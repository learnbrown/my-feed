package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	fallbackSecretOnce sync.Once
	fallbackSecret     []byte
)

// [x] 密钥放入配置文件中
func jwtSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret != "" {
		return []byte(secret)
	}

	fallbackSecretOnce.Do(func() {
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			log.Printf("FATAL: cannot generate JWT secret: %v", err)
			fallbackSecret = []byte("fallback-unsafe-key-change-me")
			return
		}
		fallbackSecret = []byte(hex.EncodeToString(b))
		log.Printf("WARNING: JWT_SECRET not set, generated random key. All tokens invalid on restart.")
	})

	return fallbackSecret
}

type Claims struct {
	AccountID uint   `json:"account_id"`
	Username  string `json:"username"`
	jwt.RegisteredClaims
}

func GenerateToken(accountID uint, username string) (string, error) {
	tokenIDBytes := make([]byte, 16)
	if _, err := rand.Read(tokenIDBytes); err != nil {
		return "", err
	}

	// 设置token的声明信息
	claims := Claims{
		AccountID: accountID,
		Username:  username,
		RegisteredClaims: jwt.RegisteredClaims{
			// 保证同一用户连续登录时也会生成不同 token，使旧 token 能被服务端撤销。
			ID: hex.EncodeToString(tokenIDBytes),
			// 设置2小时后过期
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),

			// 签发时间
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}

	// 使用HS256算法创建token对象
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 使用使用密钥签名并获得完整字符串token
	return token.SignedString(jwtSecret())
}

// 解析token
func ParseToken(tokenString string) (claims *Claims, err error) {
	claims = &Claims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			// [x] 检查签名算法
			/*
				原因是防止“算法混淆”：服务端必须确认 token 用的是自己预期的签名算法，
				不能让攻击者构造一个奇怪算法的 token 来绕验证。
				更严格一点可以直接判断 token.Method == jwt.SigningMethodHS256。
			*/
			if token.Method == nil || token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, errors.New("unexpected signing method")
			}

			return jwtSecret(), nil
		},
	)

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
