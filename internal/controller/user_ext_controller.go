package controller

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo/options"

    "roleplay/internal/model"
    "roleplay/internal/repository"
)

// UserHeartbeat 心跳更新最近在线时间
func UserHeartbeat(c *gin.Context) {
    userId := c.GetString("userId")
    _, _ = repository.DB().Collection("users").UpdateOne(c, bson.M{"userId": userId}, bson.M{"$set": bson.M{"lastSeenAt": time.Now()}})
    respond(c, http.StatusOK, "ok", nil)
}

// GetUserProfile 用户主页聚合
func GetUserProfile(c *gin.Context) {
    targetId := c.Param("user_id")
    currentId := c.GetString("userId")

    var u model.User
    if err := repository.DB().Collection("users").FindOne(c, bson.M{"userId": targetId}).Decode(&u); err != nil {
        respond(c, http.StatusNotFound, "用户不存在", nil)
        return
    }

    // 统计
    var stats model.UserStats
    _ = repository.DB().Collection("user_stats").FindOne(c, bson.M{"userId": targetId}).Decode(&stats)

    // 是否关注
    isFollowing := false
    if currentId != "" && currentId != targetId {
        cnt, _ := repository.DB().Collection("follow_edges").CountDocuments(c, bson.M{"followerId": currentId, "followingId": targetId})
        isFollowing = cnt > 0
    }

    // 在线状态：5分钟内心跳视为在线
    online := time.Since(u.LastSeenAt) <= 5*time.Minute

    respond(c, http.StatusOK, "success", gin.H{
        "profile": gin.H{
            "user_id":         u.UserId,
            "nickname":        u.Nickname,
            "avatar":          u.Avatar,
            "bio":             u.Bio,
            "gender":          u.Gender,
            "online":          online,
            "last_seen_at":    u.LastSeenAt,
            "followers_count": stats.FollowersCount,
            "following_count": stats.FollowingCount,
            "created_at":      u.CreatedAt,
            "updated_at":      u.UpdatedAt,
        },
        "is_following": isFollowing,
        "can_follow":   currentId != "" && currentId != targetId,
    })
}

// GetUserActivities 用户最近动态游标分页
func GetUserActivities(c *gin.Context) {
    targetId := c.Param("user_id")
    lastId := c.Query("last_id")
    limit := int64(20)

    filter := bson.M{"userId": targetId}
    if lastId != "" {
        if oid, err := primitive.ObjectIDFromHex(lastId); err == nil {
            filter["_id"] = bson.M{"$lt": oid}
        }
    }
    cur, err := repository.DB().Collection("user_activities").Find(c, filter, options.Find().SetSort(bson.M{"_id": -1}).SetLimit(limit))
    if err != nil {
        respond(c, http.StatusInternalServerError, "查询失败", nil)
        return
    }
    var list []model.UserActivity
    _ = cur.All(c, &list)
    next := ""
    if len(list) > 0 {
        next = list[len(list)-1].ID.Hex()
    }
    respond(c, http.StatusOK, "success", gin.H{"activities": list, "next_cursor": next})
}

