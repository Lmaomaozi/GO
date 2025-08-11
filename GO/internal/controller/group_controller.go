package controller

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"

    "roleplay/internal/model"
    "roleplay/internal/repository"
)

// CreateGroup creates a new group owned by current user.
func CreateGroup(c *gin.Context) {
    userId := c.GetString("userId")
    var body struct { Name string `json:"name"`; Avatar string `json:"avatar"` }
    if err := c.ShouldBindJSON(&body); err != nil || body.Name == "" {
        respond(c, http.StatusBadRequest, "invalid request", nil)
        return
    }
    g := model.Group{Name: body.Name, Avatar: body.Avatar, OwnerId: userId, CreatedAt: time.Now(), UpdatedAt: time.Now()}
    res, err := repository.DB().Collection("groups").InsertOne(c, g)
    if err != nil { respond(c, http.StatusInternalServerError, "server error", nil); return }
    gid := res.InsertedID.(primitive.ObjectID)
    _, _ = repository.DB().Collection("group_members").InsertOne(c, model.GroupMember{GroupId: gid, UserId: userId, Role: "owner", JoinedAt: time.Now()})
    respond(c, http.StatusOK, "success", gin.H{"group_id": gid.Hex()})
}

// AddGroupMembers adds members to a group (owner or admin only; simple owner check here).
func AddGroupMembers(c *gin.Context) {
    userId := c.GetString("userId")
    gidHex := c.Param("group_id")
    gid, err := primitive.ObjectIDFromHex(gidHex)
    if err != nil { respond(c, http.StatusBadRequest, "invalid id", nil); return }
    // Owner check
    var g model.Group
    if err := repository.DB().Collection("groups").FindOne(c, bson.M{"_id": gid}).Decode(&g); err != nil || g.OwnerId != userId {
        respond(c, http.StatusForbidden, "forbidden", nil)
        return
    }
    var body struct{ UserIds []string `json:"user_ids"` }
    if err := c.ShouldBindJSON(&body); err != nil || len(body.UserIds) == 0 {
        respond(c, http.StatusBadRequest, "invalid request", nil)
        return
    }
    for _, uid := range body.UserIds {
        _, _ = repository.DB().Collection("group_members").InsertOne(c, model.GroupMember{GroupId: gid, UserId: uid, Role: "member", JoinedAt: time.Now()})
    }
    respond(c, http.StatusOK, "success", nil)
}

// RemoveGroupMember removes a member or self leaves.
func RemoveGroupMember(c *gin.Context) {
    userId := c.GetString("userId")
    gid, err := primitive.ObjectIDFromHex(c.Param("group_id"))
    if err != nil { respond(c, http.StatusBadRequest, "invalid id", nil); return }
    target := c.Param("user_id")
    // Only owner can remove others; allow self remove
    var g model.Group
    if err := repository.DB().Collection("groups").FindOne(c, bson.M{"_id": gid}).Decode(&g); err != nil { respond(c, http.StatusNotFound, "group not found", nil); return }
    if target != userId && g.OwnerId != userId { respond(c, http.StatusForbidden, "forbidden", nil); return }
    _, _ = repository.DB().Collection("group_members").DeleteOne(c, bson.M{"groupId": gid, "userId": target})
    respond(c, http.StatusOK, "success", nil)
}

// ListMyGroups returns groups I joined.
func ListMyGroups(c *gin.Context) {
    userId := c.GetString("userId")
    cur, err := repository.DB().Collection("group_members").Find(c, bson.M{"userId": userId})
    if err != nil { respond(c, http.StatusInternalServerError, "server error", nil); return }
    var m []model.GroupMember
    _ = cur.All(c, &m)
    respond(c, http.StatusOK, "success", gin.H{"memberships": m})
}

// GetGroup returns group detail including members.
func GetGroup(c *gin.Context) {
    gid, err := primitive.ObjectIDFromHex(c.Param("group_id"))
    if err != nil { respond(c, http.StatusBadRequest, "invalid id", nil); return }
    var g model.Group
    if err := repository.DB().Collection("groups").FindOne(c, bson.M{"_id": gid}).Decode(&g); err != nil { respond(c, http.StatusNotFound, "group not found", nil); return }
    cur, _ := repository.DB().Collection("group_members").Find(c, bson.M{"groupId": gid})
    var members []model.GroupMember
    _ = cur.All(c, &members)
    respond(c, http.StatusOK, "success", gin.H{"group": g, "members": members})
}

