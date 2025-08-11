package controller

import "github.com/gin-gonic/gin"

// respond sends a JSON response with uniform envelope.
func respond(c *gin.Context, code int, message string, data any) {
    c.JSON(code, gin.H{
        "code":    code,
        "message": message,
        "data":    data,
    })
}

