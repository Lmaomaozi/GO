package controller

import (
    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

// respond 以统一外层结构返回 JSON。
func respond(c *gin.Context, code int, message string, data any) {
    c.JSON(code, gin.H{
        "code":    code,
        "message": message,
        "data":    data,
    })
}

func parseObjectID(hex string) (primitive.ObjectID, error) { return primitive.ObjectIDFromHex(hex) }

