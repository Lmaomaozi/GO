package controller

import (
    "net/http"
    "time"
    "fmt"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"

    "roleplay/internal/model"
    "roleplay/internal/repository"
)

type sendMsgReq struct {
    ConversationType string                 `json:"conversation_type"` // dm|group|room
    ConversationId   string                 `json:"conversation_id"`
    MessageType      string                 `json:"message_type"`
    Element          map[string]interface{} `json:"element"`
    CharacterId      string                 `json:"character_id"`
}

// SendMessage 发送消息（统一接口，支持私聊/群聊/房间）。
func SendMessage(c *gin.Context) {
    userId := c.GetString("userId")
    var req sendMsgReq
    if err := c.ShouldBindJSON(&req); err != nil || req.ConversationId == "" {
        respond(c, http.StatusBadRequest, "invalid request", nil)
        return
    }
    seq, err := nextSeq(c, req.ConversationId)
    if err != nil { respond(c, http.StatusInternalServerError, "server error", nil); return }
    now := time.Now()
    elemType, _ := req.Element["type"].(string)
    msg := model.Message{
        ConversationId:   req.ConversationId,
        ConversationType: req.ConversationType,
        Seq:              seq,
        SenderUserId:     userId,
        MessageType:      req.MessageType,
        Element:          model.MessageElement{Type: elemType, Data: req.Element},
        CreatedAt:        now,
        UpdatedAt:        now,
    }
    if req.MessageType == "character" {
        msg.CharacterInfo = &model.CharacterInfo{CharacterId: req.CharacterId}
    }
    _, err = repository.DB().Collection("messages").InsertOne(c, msg)
    if err != nil { respond(c, http.StatusInternalServerError, "server error", nil); return }
    upsertConversation(c, req.ConversationId, req.ConversationType, []string{userId}, seq, summarize(msg))
    respond(c, http.StatusOK, "success", gin.H{"seq": seq})
}

// GetMessageHistory 按 seq 进行分页查询历史消息。
func GetMessageHistory(c *gin.Context) {
    convType := c.Query("conversation_type")
    convId := c.Query("conversation_id")
    var lastSeq int64
    fmt.Sscan(c.DefaultQuery("lastSeq", "0"), &lastSeq)
    var limit int64 = 50
    fmt.Sscan(c.DefaultQuery("limit", "50"), &limit)
    if convId == "" { respond(c, http.StatusBadRequest, "missing conversation_id", nil); return }
    filter := bson.M{"conversationId": convId}
    if lastSeq > 0 {
        filter["seq"] = bson.M{"$gt": lastSeq}
    }
    opts := options.Find().SetSort(bson.M{"seq": 1}).SetLimit(limit)
    cur, err := repository.DB().Collection("messages").Find(c, filter, opts)
    if err != nil { respond(c, http.StatusInternalServerError, "server error", nil); return }
    var list []model.Message
    _ = cur.All(c, &list)
    respond(c, http.StatusOK, "success", gin.H{"conversation_type": convType, "conversation_id": convId, "messages": list})
}

func nextSeq(c *gin.Context, conversationId string) (int64, error) {
    var res struct{ Seq int64 `bson:"seq"` }
    upsert := true
    after := options.After
    err := repository.DB().Collection("counters").FindOneAndUpdate(c,
        bson.M{"_id": conversationId},
        bson.M{"$inc": bson.M{"seq": 1}},
        &options.FindOneAndUpdateOptions{Upsert: &upsert, ReturnDocument: &after},
    ).Decode(&res)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return 1, nil
        }
        return 0, err
    }
    return res.Seq, nil
}

func upsertConversation(c *gin.Context, conversationId, conversationType string, participants []string, lastSeq int64, lastMsg string) {
    now := time.Now()
    _, _ = repository.DB().Collection("conversations").UpdateOne(c, bson.M{"conversationId": conversationId}, bson.M{
        "$setOnInsert": bson.M{"participants": participants, "conversationType": conversationType},
        "$set":        bson.M{"lastSeq": lastSeq, "lastMessage": lastMsg, "updatedAt": now},
    }, options.Update().SetUpsert(true))
}

func summarize(m model.Message) string {
    if t, ok := m.Element.Data["text"].(string); ok { return t }
    return m.Element.Type
}

