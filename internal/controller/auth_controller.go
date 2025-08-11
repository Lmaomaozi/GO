package controller

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"

	//"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"

	"roleplay/internal/auth"
	"roleplay/internal/config"
	"roleplay/internal/model"
	"roleplay/internal/repository"
)

var validate = validator.New()

type sendCodeReq struct {
	Phone string `json:"phone" validate:"required"`
}

// SendCode 下发登录验证码（根据配置可为Mock或真实通道）。
func SendCode(c *gin.Context) {
	var req sendCodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, http.StatusBadRequest, "invalid request", nil)
		return
	}
	if err := validate.Struct(&req); err != nil {
		respond(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	code := config.C.SMS.MockCode
	now := time.Now()
	ac := model.AuthCode{
		Phone:     req.Phone,
		Code:      code,
		Scene:     "login",
		CreatedAt: now,
		ExpireAt:  now.Add(10 * time.Minute),
	}
	_, err := repository.DB().Collection("auth_codes").InsertOne(c, ac)
	if err != nil {
		zap.L().Error("insert auth code", zap.Error(err))
		respond(c, http.StatusInternalServerError, "server error", nil)
		return
	}
	respond(c, http.StatusOK, "success", gin.H{"mock_code": code})
}

type loginReq struct {
	Phone string `json:"phone" validate:"required"`
	Code  string `json:"code" validate:"required"`
}

// Login 使用手机号+验证码登录；用户不存在则创建并登录，同时签发令牌。
func Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, http.StatusBadRequest, "invalid request", nil)
		return
	}
	if err := validate.Struct(&req); err != nil {
		respond(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	// Verify code (exists and not expired)
	var ac model.AuthCode
	err := repository.DB().Collection("auth_codes").FindOne(c, bson.M{
		"phone":    req.Phone,
		"scene":    "login",
		"code":     req.Code,
		"expireAt": bson.M{"$gt": time.Now()},
	}).Decode(&ac)
	if err != nil {
		respond(c, http.StatusUnauthorized, "invalid code", nil)
		return
	}

	// Upsert user by phone
	users := repository.DB().Collection("users")
	now := time.Now()
	userId := fmt.Sprintf("u_%s", req.Phone)
	userOpenId := strings.ToUpper(req.Phone)
	update := bson.M{
		"$setOnInsert": bson.M{
			"phone":      req.Phone,
			"userId":     userId,
			"userOpenId": userOpenId,
			"nickname":   "用户" + req.Phone[max(0, len(req.Phone)-4):],
			"avatar":     "",
			"createdAt":  now,
		},
		"$set": bson.M{
			"updatedAt": now,
		},
	}
	upsert := true
	after := options.After
	var u model.User
	err = users.FindOneAndUpdate(c, bson.M{"phone": req.Phone}, update, &options.FindOneAndUpdateOptions{Upsert: &upsert, ReturnDocument: &after}).Decode(&u)
	if err != nil {
		respond(c, http.StatusInternalServerError, "server error", nil)
		return
	}

	access, refresh, err := auth.GenerateTokens(u.UserId)
	if err != nil {
		respond(c, http.StatusInternalServerError, "token error", nil)
		return
	}
	respond(c, http.StatusOK, "success", gin.H{"accessToken": access, "refreshToken": refresh})
}

// RefreshToken 使用刷新令牌换取新的访问令牌。
func RefreshToken(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.RefreshToken == "" {
		respond(c, http.StatusBadRequest, "invalid request", nil)
		return
	}
	claims, err := auth.ParseToken(body.RefreshToken)
	if err != nil {
		respond(c, http.StatusUnauthorized, "invalid refresh token", nil)
		return
	}
	// 过期校验由 ParseToken 完成。此处仅旋转新对
	access, refresh, err := auth.GenerateTokens(claims.UserId)
	if err != nil {
		respond(c, http.StatusInternalServerError, "token error", nil)
		return
	}
	respond(c, http.StatusOK, "success", gin.H{"accessToken": access, "refreshToken": refresh})
}

type oneClickLoginReq struct {
	Phone    string `json:"phone" validate:"required"`
	DeviceId string `json:"device_id" validate:"required"`
	Platform string `json:"platform" validate:"required,oneof=android ios web"`
}

// OneClickLogin 一键登录（模拟运营商验证，实际项目中可接入真实运营商API）
func OneClickLogin(c *gin.Context) {
	var req oneClickLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, http.StatusBadRequest, "invalid request", nil)
		return
	}
	if err := validate.Struct(&req); err != nil {
		respond(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// 模拟运营商验证：检查设备ID和手机号是否匹配
	// 实际项目中这里应该调用运营商API验证
	if req.DeviceId == "" || req.Phone == "" {
		respond(c, http.StatusBadRequest, "invalid device or phone", nil)
		return
	}

	// 查找或创建用户（类似登录逻辑）
	users := repository.DB().Collection("users")
	now := time.Now()
	userId := fmt.Sprintf("u_%s", req.Phone)
	userOpenId := strings.ToUpper(req.Phone)

	update := bson.M{
		"$setOnInsert": bson.M{
			"phone":      req.Phone,
			"userId":     userId,
			"userOpenId": userOpenId,
			"nickname":   "用户" + req.Phone[max(0, len(req.Phone)-4):],
			"avatar":     "",
			"createdAt":  now,
		},
		"$set": bson.M{
			"updatedAt": now,
		},
	}

	upsert := true
	after := options.After
	var u model.User
	err := users.FindOneAndUpdate(c, bson.M{"phone": req.Phone}, update, &options.FindOneAndUpdateOptions{Upsert: &upsert, ReturnDocument: &after}).Decode(&u)
	if err != nil {
		respond(c, http.StatusInternalServerError, "server error", nil)
		return
	}

	// 生成令牌
	access, refresh, err := auth.GenerateTokens(u.UserId)
	if err != nil {
		respond(c, http.StatusInternalServerError, "token error", nil)
		return
	}

	respond(c, http.StatusOK, "success", gin.H{"accessToken": access, "refreshToken": refresh})
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
