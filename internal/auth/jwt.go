package auth

import (
    "time"

    "github.com/golang-jwt/jwt/v5"
    "go.uber.org/zap"

    "roleplay/internal/config"
)

// Claims 定义 JWT 自定义负载，包含用户ID。
type Claims struct {
    UserId string `json:"user_id"`
    jwt.RegisteredClaims
}

// GenerateTokens 为给定用户签发访问令牌与刷新令牌。
func GenerateTokens(userId string) (accessToken string, refreshToken string, err error) {
    now := time.Now()
    access := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
        UserId: userId,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(now.Add(config.AccessTTL())),
            IssuedAt:  jwt.NewNumericDate(now),
        },
    })
    refresh := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
        UserId: userId,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(now.Add(config.RefreshTTL())),
            IssuedAt:  jwt.NewNumericDate(now),
        },
    })
    a, err := access.SignedString([]byte(config.C.JWT.Secret))
    if err != nil {
        return "", "", err
    }
    r, err := refresh.SignedString([]byte(config.C.JWT.Secret))
    if err != nil {
        return "", "", err
    }
    return a, r, nil
}

// ParseToken 校验JWT并返回负载。
func ParseToken(token string) (*Claims, error) {
    t, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(config.C.JWT.Secret), nil
    })
    if err != nil {
        return nil, err
    }
    if claims, ok := t.Claims.(*Claims); ok && t.Valid {
        return claims, nil
    }
    zap.L().Warn("invalid token claims")
    return nil, jwt.ErrTokenInvalidClaims
}

