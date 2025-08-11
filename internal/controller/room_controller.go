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

// JoinRoom 根据 recruit_id 创建/加入演绎房间，并记录选择的角色。
func JoinRoom(c *gin.Context) {
    userId := c.GetString("userId")
    var body struct {
        RecruitId   string `json:"recruit_id"`
        CharacterId string `json:"character_id"`
    }
    if err := c.ShouldBindJSON(&body); err != nil || body.RecruitId == "" {
        respond(c, http.StatusBadRequest, "invalid request", nil)
        return
    }
    rid, err := primitive.ObjectIDFromHex(body.RecruitId)
    if err != nil { respond(c, http.StatusBadRequest, "invalid id", nil); return }
    var th model.Theater
    err = repository.DB().Collection("theaters").FindOne(c, bson.M{"recruitId": rid}).Decode(&th)
    if err != nil {
        th = model.Theater{RecruitId: rid, Title: "演绎房间", Mode: "couple", Status: "active", CreatedAt: time.Now(), UpdatedAt: time.Now()}
        res, _ := repository.DB().Collection("theaters").InsertOne(c, th)
        th.ID = res.InsertedID.(primitive.ObjectID)
    }
    exists := false
    for _, p := range th.Participants { if p.UserId == userId { exists = true; break } }
    if !exists {
        th.Participants = append(th.Participants, model.TheaterParticipant{UserId: userId, CostumeId: body.CharacterId, JoinTime: time.Now()})
        _, _ = repository.DB().Collection("theaters").UpdateByID(c, th.ID, bson.M{"$set": bson.M{"participants": th.Participants, "updatedAt": time.Now()}})
    }
    respond(c, http.StatusOK, "success", gin.H{"room_id": th.ID.Hex()})
}

// GetRoomMessages 复用统一消息历史接口，conversation_id 使用 room_id。
func GetRoomMessages(c *gin.Context) {
    rid := c.Param("id")
    c.Request.URL.RawQuery = "conversation_type=room&conversation_id=" + rid + "&" + c.Request.URL.RawQuery
    GetMessageHistory(c)
}

// SendRoomMessage 复用统一发送接口，conversation_id 使用 room_id。
func SendRoomMessage(c *gin.Context) {
    rid := c.Param("id")
    var body map[string]interface{}
    if err := c.ShouldBindJSON(&body); err != nil { respond(c, http.StatusBadRequest, "invalid request", nil); return }
    body["conversation_type"] = "room"
    body["conversation_id"] = rid
    // Reuse SendMessage by binding again
    c.Set("_override_body", body)
    // Quick rebuild
    type req struct {
        ConversationType string                 `json:"conversation_type"`
        ConversationId   string                 `json:"conversation_id"`
        MessageType      string                 `json:"message_type"`
        Element          map[string]interface{} `json:"element"`
        CharacterId      string                 `json:"character_id"`
    }
    var r req
    r.ConversationType = "room"
    r.ConversationId = rid
    if v, ok := body["message_type"].(string); ok { r.MessageType = v }
    if v, ok := body["element"].(map[string]interface{}); ok { r.Element = v }
    if v, ok := body["character_id"].(string); ok { r.CharacterId = v }
    // Delegate
    SendMessage(c)
}

