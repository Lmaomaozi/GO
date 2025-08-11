package controller

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"

    "roleplay/internal/model"
    "roleplay/internal/repository"
)

// GetMe 获取当前登录用户的资料。
func GetMe(c *gin.Context) {
    userId := c.GetString("userId")
    var u model.User
    if err := repository.DB().Collection("users").FindOne(c, bson.M{"userId": userId}).Decode(&u); err != nil {
        respond(c, http.StatusNotFound, "user not found", nil)
        return
    }
    respond(c, http.StatusOK, "success", u)
}

// UpdateMe 更新当前登录用户的资料。
func UpdateMe(c *gin.Context) {
    userId := c.GetString("userId")
    var body struct {
        Nickname string `json:"nickname"`
        Avatar   string `json:"avatar"`
        Gender   string `json:"gender"`
        Bio      string `json:"bio"`
    }
    if err := c.ShouldBindJSON(&body); err != nil {
        respond(c, http.StatusBadRequest, "invalid request", nil)
        return
    }
    update := bson.M{
        "$set": bson.M{
            "nickname": body.Nickname,
            "avatar":   body.Avatar,
            "gender":   body.Gender,
            "bio":      body.Bio,
            "updatedAt": time.Now(),
        },
    }
    _, err := repository.DB().Collection("users").UpdateOne(c, bson.M{"userId": userId}, update)
    if err != nil {
        respond(c, http.StatusInternalServerError, "server error", nil)
        return
    }
    GetMe(c)
}

