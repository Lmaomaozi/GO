package controller

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo/options"

    "roleplay/internal/model"
    "roleplay/internal/repository"
)

// FollowUser 关注用户（去重、禁止自关注）
func FollowUser(c *gin.Context) {
    follower := c.GetString("userId")
    target := c.Param("user_id")
    if follower == target {
        respond(c, http.StatusBadRequest, "不能关注自己", nil)
        return
    }
    now := time.Now()
    // 插入关注关系（利用唯一索引去重）
    _, err := repository.DB().Collection("follow_edges").InsertOne(c, model.FollowEdge{FollowerId: follower, FollowingId: target, CreatedAt: now, UpdatedAt: now})
    if err != nil {
        // 可能是重复关注
    }
    // 计数自增
    _, _ = repository.DB().Collection("user_stats").UpdateOne(c, bson.M{"userId": follower}, bson.M{"$inc": bson.M{"followingCount": 1}}, options.Update().SetUpsert(true))
    _, _ = repository.DB().Collection("user_stats").UpdateOne(c, bson.M{"userId": target}, bson.M{"$inc": bson.M{"followersCount": 1}}, options.Update().SetUpsert(true))
    // 活动记录
    _, _ = repository.DB().Collection("user_activities").InsertOne(c, model.UserActivity{UserId: follower, ActivityType: "follow", TargetType: "user", TargetId: target, Title: "关注了用户", CreatedAt: now})
    respond(c, http.StatusOK, "关注成功", nil)
}

// UnfollowUser 取消关注
func UnfollowUser(c *gin.Context) {
    follower := c.GetString("userId")
    target := c.Param("user_id")
    _, _ = repository.DB().Collection("follow_edges").DeleteOne(c, bson.M{"followerId": follower, "followingId": target})
    // 计数自减
    _, _ = repository.DB().Collection("user_stats").UpdateOne(c, bson.M{"userId": follower}, bson.M{"$inc": bson.M{"followingCount": -1}})
    _, _ = repository.DB().Collection("user_stats").UpdateOne(c, bson.M{"userId": target}, bson.M{"$inc": bson.M{"followersCount": -1}})
    respond(c, http.StatusOK, "已取消关注", nil)
}

// GetFollowStatus 关注状态
func GetFollowStatus(c *gin.Context) {
    follower := c.GetString("userId")
    target := c.Param("user_id")
    cnt, _ := repository.DB().Collection("follow_edges").CountDocuments(c, bson.M{"followerId": follower, "followingId": target})
    respond(c, http.StatusOK, "success", gin.H{"is_following": cnt > 0})
}

// ListFollowers 粉丝列表（简化，仅返回用户ID）
func ListFollowers(c *gin.Context) {
    userId := c.GetString("userId")
    cur, _ := repository.DB().Collection("follow_edges").Find(c, bson.M{"followingId": userId})
    var list []model.FollowEdge
    _ = cur.All(c, &list)
    ids := make([]string, 0, len(list))
    for _, e := range list { ids = append(ids, e.FollowerId) }
    respond(c, http.StatusOK, "success", gin.H{"followers": ids})
}

// ListFollowing 关注列表
func ListFollowing(c *gin.Context) {
    userId := c.GetString("userId")
    cur, _ := repository.DB().Collection("follow_edges").Find(c, bson.M{"followerId": userId})
    var list []model.FollowEdge
    _ = cur.All(c, &list)
    ids := make([]string, 0, len(list))
    for _, e := range list { ids = append(ids, e.FollowingId) }
    respond(c, http.StatusOK, "success", gin.H{"following": ids})
}

// remove custom options types; we use official mongo options above

