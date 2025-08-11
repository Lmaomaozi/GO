package controller

import (
    "net/http"
    "sort"
    "time"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"

    "roleplay/internal/model"
    "roleplay/internal/repository"
)

// CreateFriendRequest 发起好友申请。
func CreateFriendRequest(c *gin.Context) {
    userId := c.GetString("userId")
    var body struct {
        UserId   string `json:"user_id"`
        Greeting string `json:"greeting"`
    }
    if err := c.ShouldBindJSON(&body); err != nil || body.UserId == "" {
        respond(c, http.StatusBadRequest, "invalid request", nil)
        return
    }
    if body.UserId == userId {
        respond(c, http.StatusBadRequest, "cannot add self", nil)
        return
    }
    fr := model.FriendRequest{
        RequesterId: userId,
        RecipientId: body.UserId,
        Greeting:    body.Greeting,
        Status:      "pending",
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }
    _, err := repository.DB().Collection("friend_requests").InsertOne(c, fr)
    if err != nil {
        respond(c, http.StatusConflict, "request exists", nil)
        return
    }
    respond(c, http.StatusOK, "success", nil)
}

// RespondFriendRequest 同意或拒绝好友申请。
func RespondFriendRequest(c *gin.Context) {
    userId := c.GetString("userId")
    var body struct {
        RequestId string `json:"request_id"`
        Action    string `json:"action"` // accept|reject
    }
    if err := c.ShouldBindJSON(&body); err != nil || (body.Action != "accept" && body.Action != "reject") {
        respond(c, http.StatusBadRequest, "invalid request", nil)
        return
    }
    oid, err := primitive.ObjectIDFromHex(body.RequestId)
    if err != nil {
        respond(c, http.StatusBadRequest, "invalid id", nil)
        return
    }
    var fr model.FriendRequest
    if err := repository.DB().Collection("friend_requests").FindOne(c, bson.M{"_id": oid, "recipientId": userId, "status": "pending"}).Decode(&fr); err != nil {
        respond(c, http.StatusNotFound, "request not found", nil)
        return
    }
    newStatus := map[string]string{"accept": "accepted", "reject": "rejected"}[body.Action]
    _, err = repository.DB().Collection("friend_requests").UpdateByID(c, oid, bson.M{"$set": bson.M{"status": newStatus, "updatedAt": time.Now()}})
    if err != nil {
        respond(c, http.StatusInternalServerError, "server error", nil)
        return
    }
    if newStatus == "accepted" {
        a, b := orderPair(fr.RequesterId, fr.RecipientId)
        _, _ = repository.DB().Collection("friends").InsertOne(c, model.FriendEdge{UserA: a, UserB: b, CreatedAt: time.Now()})
    }
    respond(c, http.StatusOK, "success", nil)
}

func orderPair(a, b string) (string, string) {
    arr := []string{a, b}
    sort.Strings(arr)
    return arr[0], arr[1]
}

// ListFriendRequests 列出与我相关的好友申请（我收到/我发起）。
func ListFriendRequests(c *gin.Context) {
    userId := c.GetString("userId")
    cur, err := repository.DB().Collection("friend_requests").Find(c, bson.M{"$or": []bson.M{{"recipientId": userId}, {"requesterId": userId}}})
    if err != nil {
        respond(c, http.StatusInternalServerError, "server error", nil)
        return
    }
    var list []model.FriendRequest
    if err := cur.All(c, &list); err != nil {
        respond(c, http.StatusInternalServerError, "server error", nil)
        return
    }
    respond(c, http.StatusOK, "success", gin.H{"list": list})
}

// ListFriends 返回我的好友列表。
func ListFriends(c *gin.Context) {
    userId := c.GetString("userId")
    aCur, err := repository.DB().Collection("friends").Find(c, bson.M{"userA": userId})
    if err != nil { respond(c, http.StatusInternalServerError, "server error", nil); return }
    var edgesA []model.FriendEdge
    _ = aCur.All(c, &edgesA)
    bCur, err := repository.DB().Collection("friends").Find(c, bson.M{"userB": userId})
    if err != nil { respond(c, http.StatusInternalServerError, "server error", nil); return }
    var edgesB []model.FriendEdge
    _ = bCur.All(c, &edgesB)
    ids := make([]string, 0, len(edgesA)+len(edgesB))
    for _, e := range edgesA { ids = append(ids, e.UserB) }
    for _, e := range edgesB { ids = append(ids, e.UserA) }
    respond(c, http.StatusOK, "success", gin.H{"friends": ids})
}

// DeleteFriend 解除好友关系。
func DeleteFriend(c *gin.Context) {
    userId := c.GetString("userId")
    other := c.Param("user_id")
    a, b := orderPair(userId, other)
    _, err := repository.DB().Collection("friends").DeleteOne(c, bson.M{"userA": a, "userB": b})
    if err != nil { respond(c, http.StatusInternalServerError, "server error", nil); return }
    respond(c, http.StatusOK, "success", nil)
}

// BlockUser 拉黑某用户，禁止与其交互。
func BlockUser(c *gin.Context) {
    userId := c.GetString("userId")
    other := c.Param("user_id")
    be := model.BlockEdge{UserId: userId, BlockedUserId: other, CreatedAt: time.Now()}
    _, err := repository.DB().Collection("blocks").InsertOne(c, be)
    if err != nil { respond(c, http.StatusConflict, "already blocked", nil); return }
    respond(c, http.StatusOK, "success", nil)
}

// UnblockUser 取消拉黑。
func UnblockUser(c *gin.Context) {
    userId := c.GetString("userId")
    other := c.Param("user_id")
    _, err := repository.DB().Collection("blocks").DeleteOne(c, bson.M{"userId": userId, "blockedUserId": other})
    if err != nil { respond(c, http.StatusInternalServerError, "server error", nil); return }
    respond(c, http.StatusOK, "success", nil)
}

// ListBlocks 列出我的黑名单。
func ListBlocks(c *gin.Context) {
    userId := c.GetString("userId")
    cur, err := repository.DB().Collection("blocks").Find(c, bson.M{"userId": userId})
    if err != nil { respond(c, http.StatusInternalServerError, "server error", nil); return }
    var list []model.BlockEdge
    _ = cur.All(c, &list)
    respond(c, http.StatusOK, "success", gin.H{"list": list})
}

