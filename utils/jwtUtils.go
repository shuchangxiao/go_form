package utils

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"strings"
	"time"
	"veripTest/config"
)

func CreateJWT(id uint) (string, error) {
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  id,
		"exp": time.Now().Add(time.Duration(config.Cf.Jwt.Exptime) * time.Hour).Unix(),
	})
	signedString, err := claims.SignedString([]byte(config.Cf.Jwt.Secret))
	return signedString, err
}

// ValidJWT 验证JWT令牌的有效性。
// 该函数接收一个令牌字符串作为输入，如果令牌有效，返回用户ID和一个表示验证结果的布尔值。
// 如果令牌无效或格式不正确，返回-1和false。
func ValidJWT(tokenString string) (int, bool) {
	// 检查令牌字符串长度和前缀，确保它以"Bearer "开始。
	if len(tokenString) <= 7 || !strings.HasPrefix(tokenString, "Bearer ") {
		return -1, false
	}
	// 去除"Bearer "前缀。
	tokenString = tokenString[7:]
	// 解析JWT令牌。
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 检查签名方法是否为HMAC，如果不是，返回错误。
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected login method %v", token.Header["alg"])
		}
		// 返回用于验证JWT的密钥。
		return []byte(config.Cf.Jwt.Secret), nil
	})
	// 处理解析过程中可能发生的错误。
	if err != nil {
		return -1, false
	}
	// 检查令牌的声明部分，并验证令牌的有效性。
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		// 如果令牌有效，提取并返回用户ID。
		id := claims["id"].(int)
		return id, true
	}
	// 如果令牌无效或验证失败，返回-1和false。
	return -1, false
}
