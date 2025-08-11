package indexer

import (
    "context"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.uber.org/zap"

    "roleplay/internal/repository"
)

// EnsureAllIndexes 创建各集合所需索引（幂等）。
func EnsureAllIndexes(ctx context.Context) error {
    db := repository.DB()
    // users 用户集合
    if err := createIndexes(ctx, db.Collection("users"), []mongo.IndexModel{
        {Keys: bson.D{{Key: "phone", Value: 1}}, Options: options.Index().SetUnique(true)},
        {Keys: bson.D{{Key: "userId", Value: 1}}, Options: options.Index().SetUnique(true)},
        {Keys: bson.D{{Key: "userOpenId", Value: 1}}, Options: options.Index().SetUnique(true)},
        {Keys: bson.D{{Key: "createdAt", Value: -1}}},
    }); err != nil { return err }

    // auth_codes 短信验证码集合（TTL）
    if err := createIndexes(ctx, db.Collection("auth_codes"), []mongo.IndexModel{
        {Keys: bson.D{{Key: "phone", Value: 1}, {Key: "scene", Value: 1}, {Key: "createdAt", Value: -1}}},
        {Keys: bson.D{{Key: "expireAt", Value: 1}}, Options: options.Index().SetExpireAfterSeconds(0)},
    }); err != nil { return err }

    // friend_requests 好友申请集合
    if err := createIndexes(ctx, db.Collection("friend_requests"), []mongo.IndexModel{
        {Keys: bson.D{{Key: "recipientId", Value: 1}, {Key: "status", Value: 1}, {Key: "createdAt", Value: -1}}},
        {Keys: bson.D{{Key: "requesterId", Value: 1}, {Key: "status", Value: 1}, {Key: "createdAt", Value: -1}}},
        {Keys: bson.D{{Key: "requesterId", Value: 1}, {Key: "recipientId", Value: 1}, {Key: "status", Value: 1}}},
    }); err != nil { return err }

    // friends 好友关系集合
    if err := createIndexes(ctx, db.Collection("friends"), []mongo.IndexModel{
        {Keys: bson.D{{Key: "userA", Value: 1}, {Key: "userB", Value: 1}}, Options: options.Index().SetUnique(true)},
        {Keys: bson.D{{Key: "userA", Value: 1}}},
        {Keys: bson.D{{Key: "userB", Value: 1}}},
    }); err != nil { return err }

    // blocks 黑名单集合
    if err := createIndexes(ctx, db.Collection("blocks"), []mongo.IndexModel{
        {Keys: bson.D{{Key: "userId", Value: 1}, {Key: "blockedUserId", Value: 1}}, Options: options.Index().SetUnique(true)},
        {Keys: bson.D{{Key: "userId", Value: 1}}},
    }); err != nil { return err }

    // groups 与 group_members 群与成员集合
    if err := createIndexes(ctx, db.Collection("groups"), []mongo.IndexModel{
        {Keys: bson.D{{Key: "ownerId", Value: 1}, {Key: "createdAt", Value: -1}}},
    }); err != nil { return err }
    if err := createIndexes(ctx, db.Collection("group_members"), []mongo.IndexModel{
        {Keys: bson.D{{Key: "groupId", Value: 1}, {Key: "userId", Value: 1}}, Options: options.Index().SetUnique(true)},
        {Keys: bson.D{{Key: "userId", Value: 1}}},
        {Keys: bson.D{{Key: "groupId", Value: 1}}},
    }); err != nil { return err }

    // conversations & messages & counters 会话、消息与序号计数器集合
    if err := createIndexes(ctx, db.Collection("conversations"), []mongo.IndexModel{
        {Keys: bson.D{{Key: "conversationId", Value: 1}}, Options: options.Index().SetUnique(true)},
        {Keys: bson.D{{Key: "participants", Value: 1}}},
        {Keys: bson.D{{Key: "updatedAt", Value: -1}}},
    }); err != nil { return err }
    if err := createIndexes(ctx, db.Collection("messages"), []mongo.IndexModel{
        {Keys: bson.D{{Key: "conversationId", Value: 1}, {Key: "seq", Value: 1}}, Options: options.Index().SetUnique(true)},
        {Keys: bson.D{{Key: "conversationId", Value: 1}, {Key: "createdAt", Value: -1}}},
    }); err != nil { return err }
    // if err := createIndexes(ctx, db.Collection("counters"), []mongo.IndexModel{
    //     {Keys: bson.D{{Key: "_id", Value: 1}}, Options: options.Index().SetUnique(true)},
    // }); err != nil { return err }
    //完全移除对_id的索引定义（因为MongoDB会自动创建）
    if err := createIndexes(ctx, db.Collection("counters"), []mongo.IndexModel{}); err != nil { return err }

    // theaters 演绎房间集合
    if err := createIndexes(ctx, db.Collection("theaters"), []mongo.IndexModel{
        {Keys: bson.D{{Key: "recruitId", Value: 1}}},
        {Keys: bson.D{{Key: "status", Value: 1}}},
    }); err != nil { return err }

    // follow_edges 关注关系集合（防重复关注）
    if err := createIndexes(ctx, db.Collection("follow_edges"), []mongo.IndexModel{
        {Keys: bson.D{{Key: "followerId", Value: 1}, {Key: "followingId", Value: 1}}, Options: options.Index().SetUnique(true)}},
    ); err != nil { return err }

    // user_stats 用户统计（userId 唯一）
    if err := createIndexes(ctx, db.Collection("user_stats"), []mongo.IndexModel{
        {Keys: bson.D{{Key: "userId", Value: 1}}, Options: options.Index().SetUnique(true)}},
    ); err != nil { return err }

    // user_activities 活动流
    if err := createIndexes(ctx, db.Collection("user_activities"), []mongo.IndexModel{
        {Keys: bson.D{{Key: "userId", Value: 1}, {Key: "createdAt", Value: -1}}},
    }); err != nil { return err }

    zap.L().Info("indexes ensured")
    return nil
}

func createIndexes(ctx context.Context, col *mongo.Collection, models []mongo.IndexModel) error {
    if len(models) == 0 {
        return nil
    }
    _, err := col.Indexes().CreateMany(ctx, models)
    return err
}

