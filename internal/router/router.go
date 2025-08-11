package router

import (
	"github.com/gin-gonic/gin"

	"roleplay/internal/controller"
	"roleplay/internal/middleware"
)

// New 返回一个注册好全部路由的 gin.Engine。
func New() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	// 静态资源：头像等
	r.Static("/static", "./uploads")

	// Health 健康检查
	r.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	// Public APIs 公共接口（无需鉴权）
	r.POST("/api/user/send_code", controller.SendCode)
	r.POST("/api/user/login", controller.Login)
	r.POST("/api/user/oneclick_login", controller.OneClickLogin)
	r.POST("/api/auth/refresh", controller.RefreshToken)

	// Protected group 需鉴权接口
	auth := r.Group("/api", middleware.AuthMiddleware())

	// User 用户模块
	auth.GET("/user/me", controller.GetMe)
	auth.PUT("/user/me", controller.UpdateMe)

	// Relation 好友与黑名单
	auth.POST("/relation/friend/request", controller.CreateFriendRequest)
	auth.POST("/relation/friend/respond", controller.RespondFriendRequest)
	auth.GET("/relation/friend/requests", controller.ListFriendRequests)
	auth.GET("/relation/friends", controller.ListFriends)
	auth.DELETE("/relation/friend/:user_id", controller.DeleteFriend)

	auth.POST("/relation/block/:user_id", controller.BlockUser)
	auth.DELETE("/relation/block/:user_id", controller.UnblockUser)
	auth.GET("/relation/blocks", controller.ListBlocks)

	// Groups 群组模块
	auth.POST("/group", controller.CreateGroup)
	auth.POST("/group/:group_id/members", controller.AddGroupMembers)
	auth.DELETE("/group/:group_id/members/:user_id", controller.RemoveGroupMember)
	auth.GET("/group/my", controller.ListMyGroups)
	auth.GET("/group/:group_id", controller.GetGroup)

	// Messaging 消息模块
	auth.POST("/message/send", controller.SendMessage)
	auth.GET("/message/history", controller.GetMessageHistory)

	// Room 演绎房间
	auth.POST("/room/join", controller.JoinRoom)
	auth.GET("/room/:id/messages", controller.GetRoomMessages)
	auth.POST("/room/:id/message", controller.SendRoomMessage)

	// File 文件上传
	auth.POST("/file/avatar", controller.UploadAvatar)

	// Follow 关注系统
	auth.POST("/relation/follow/:user_id", controller.FollowUser)
	auth.DELETE("/relation/follow/:user_id", controller.UnfollowUser)
	auth.GET("/relation/follow/status/:user_id", controller.GetFollowStatus)
	auth.GET("/relation/followers", controller.ListFollowers)
	auth.GET("/relation/following", controller.ListFollowing)

	// User 扩展：用户主页与心跳
	auth.GET("/user/profile/:user_id", controller.GetUserProfile)
	auth.GET("/user/activities/:user_id", controller.GetUserActivities)
	auth.POST("/user/heartbeat", controller.UserHeartbeat)

	return r
}
