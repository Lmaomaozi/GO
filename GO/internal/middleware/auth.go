package middleware

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"

    "roleplay/internal/auth"
)

// AuthMiddleware validates JWT from the Authorization header and injects user ID into context.
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        header := c.GetHeader("Authorization")
        if header == "" || !strings.HasPrefix(header, "Bearer ") {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "missing or invalid Authorization"})
            return
        }
        token := strings.TrimPrefix(header, "Bearer ")
        claims, err := auth.ParseToken(token)
        if err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "invalid token"})
            return
        }
        c.Set("userId", claims.UserId)
        c.Next()
    }
}

