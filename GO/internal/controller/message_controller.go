package controller

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.mongodb.org/mongo-driver/bson/primitive"

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

// SendMessage sends a message in dm/group/room unified API.
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
    msg := model.Message{
        ConversationId:  req.ConversationId,
        ConversationType: req.ConversationType,
        Seq:             seq,
        SenderUserId:    userId,
        MessageType:     req.MessageType,
        Element:         model.MessageElement{Type: req.Element["type"].(string), Data: req.Element},
        CreatedAt:       now,
        UpdatedAt:       now,
    }
    if req.MessageType == "character" {
        msg.CharacterInfo = &model.CharacterInfo{CharacterId: req.CharacterId}
    }
    _, err = repository.DB().Collection("messages").InsertOne(c, msg)
    if err != nil { respond(c, http.StatusInternalServerError, "server error", nil); return }
    // update conversation summary
    upsertConversation(c, req.ConversationId, req.ConversationType, []string{userId}, seq, summarize(msg))
    respond(c, http.StatusOK, "success", gin.H{"seq": seq})
}

// GetMessageHistory paginates by seq.
func GetMessageHistory(c *gin.Context) {
    convType := c.Query("conversation_type")
    convId := c.Query("conversation_id")
    lastSeq := parseInt64(c.DefaultQuery("lastSeq", "0"))
    limit := int(parseInt64(c.DefaultQuery("limit", "50")))
    if convId == "" { respond(c, http.StatusBadRequest, "missing conversation_id", nil); return }
    filter := bson.M{"conversationId": convId}
    if lastSeq > 0 {
        filter["seq"] = bson.M{"$gt": lastSeq}
    }
    opts := options.Find().SetSort(bson.M{"seq": 1}).SetLimit(int64(limit))
    cur, err := repository.DB().Collection("messages").Find(c, filter, opts)
    if err != nil { respond(c, http.StatusInternalServerError, "server error", nil); return }
    var list []model.Message
    _ = cur.All(c, &list)
    respond(c, http.StatusOK, "success", gin.H{"conversation_type": convType, "conversation_id": convId, "messages": list})
}

func parseInt64(s string) int64 {
    var x int64
    _, _ = fmt.Sscan(s, &x)
    return x
}

func nextSeq(c *gin.Context, conversationId string) (int64, error) {
    // counters: _id = conversationId, seq: int64
    var res struct{ Seq int64 `bson:"seq"` }
    err := repository.DB().Collection("counters").FindOneAndUpdate(c,
        bson.M{"_id": conversationId},
        bson.M{"$inc": bson.M{"seq": 1}},
        &mongo.FindOneAndUpdateOptions{Upsert: &[]bool{true}[0], ReturnDocument: mongo.After},
    ).Decode(&res)
    if err != nil {
        // Initialize if missing
        _, _ = repository.DB().Collection("counters").InsertOne(c, bson.M{"_id": conversationId, "seq": 1})
        res.Seq = 1
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

