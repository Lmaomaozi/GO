package controller

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	"roleplay/internal/auth"
	"roleplay/internal/config"
	"roleplay/internal/model"
	"roleplay/internal/repository"
)

var validate = validator.New()

type sendCodeReq struct {
	Phone string `json:"phone" validate:"required,e164|len=11"`
}

// SendCode 下发登录验证码（根据配置可能为 Mock 或真实通道）。
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

// Login 使用手机号+验证码登录；若用户不存在则创建并签发令牌。
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
	// Generate IDs from phone for demo; production should ensure uniqueness differently
	userId := "u_" + req.Phone
	userOpenId := strings.ToUpper(req.Phone)
	update := bson.M{
		"$setOnInsert": bson.M{
			"phone":      req.Phone,
			"userId":     userId,
			"userOpenId": userOpenId,
			"nickname":   "用户" + req.Phone[len(req.Phone)-4:],
			"avatar":     "",
			"createdAt":  now,
		},
		"$set": bson.M{
			"updatedAt": now,
		},
	}
	opts := &mongo.Options{}
	res := users.FindOneAndUpdate(c, bson.M{"phone": req.Phone}, update, &mongo.FindOneAndUpdateOptions{Upsert: &[]bool{true}[0], ReturnDocument: mongo.After})
	var u model.User
	if err := res.Decode(&u); err != nil {
		// If After not supported fallback: get by phone
		if err := users.FindOne(c, bson.M{"phone": req.Phone}).Decode(&u); err != nil {
			zap.L().Error("find user after upsert", zap.Error(err))
			respond(c, http.StatusInternalServerError, "server error", nil)
			return
		}
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
	access, refresh, err := auth.GenerateTokens(claims.UserId)
	if err != nil {
		respond(c, http.StatusInternalServerError, "token error", nil)
		return
	}
	respond(c, http.StatusOK, "success", gin.H{"accessToken": access, "refreshToken": refresh})
}

// OneClickLogin 一键登录：基于设备信息的模拟校验，无需短信验证码。
func OneClickLogin(c *gin.Context) {
	var req struct {
		Phone    string `json:"phone" validate:"required"`
		DeviceId string `json:"device_id" validate:"required"`
		Platform string `json:"platform" validate:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, http.StatusBadRequest, "invalid request", nil)
		return
	}
	if err := validate.Struct(&req); err != nil {
		respond(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

    // 模拟运营商校验：仅检查参数存在（实际应调用运营商 API）
	if req.Phone == "" || req.DeviceId == "" || req.Platform == "" {
		respond(c, http.StatusBadRequest, "invalid device or phone", nil)
		return
	}

    // 按手机号 upsert 用户（与 Login 保持一致）
	users := repository.DB().Collection("users")
	now := time.Now()
	userId := "u_" + req.Phone
	userOpenId := strings.ToUpper(req.Phone)
	update := bson.M{
		"$setOnInsert": bson.M{
			"phone":      req.Phone,
			"userId":     userId,
			"userOpenId": userOpenId,
			"nickname":   "用户" + req.Phone[len(req.Phone)-4:],
			"avatar":     "",
			"createdAt":  now,
		},
		"$set": bson.M{
			"updatedAt": now,
		},
	}
	res := users.FindOneAndUpdate(c, bson.M{"phone": req.Phone}, update, &mongo.FindOneAndUpdateOptions{Upsert: &[]bool{true}[0], ReturnDocument: mongo.After})
	var u model.User
	if err := res.Decode(&u); err != nil {
		// Fallback get by phone
		if err := users.FindOne(c, bson.M{"phone": req.Phone}).Decode(&u); err != nil {
			zap.L().Error("find user after upsert", zap.Error(err))
			respond(c, http.StatusInternalServerError, "server error", nil)
			return
		}
	}

	access, refresh, err := auth.GenerateTokens(u.UserId)
	if err != nil {
		respond(c, http.StatusInternalServerError, "token error", nil)
		return
	}
	respond(c, http.StatusOK, "success", gin.H{"accessToken": access, "refreshToken": refresh})
}
