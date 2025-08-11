package indexer

import (
    "context"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.uber.org/zap"

    "roleplay/internal/repository"
)

// EnsureAllIndexes creates required indexes for collections. It is idempotent.
func EnsureAllIndexes(ctx context.Context) error {
    db := repository.DB()
    // users
    if err := createIndexes(ctx, db.Collection("users"), []mongo.IndexModel{
        {Keys: bson.D{{Key: "phone", Value: 1}}, Options: options.Index().SetUnique(true)},
        {Keys: bson.D{{Key: "userId", Value: 1}}, Options: options.Index().SetUnique(true)},
        {Keys: bson.D{{Key: "userOpenId", Value: 1}}, Options: options.Index().SetUnique(true)},
        {Keys: bson.D{{Key: "createdAt", Value: -1}}},
    }); err != nil {
        return err
    }
    // auth_codes with TTL
    if err := createIndexes(ctx, db.Collection("auth_codes"), []mongo.IndexModel{
        {Keys: bson.D{{Key: "phone", Value: 1}, {Key: "scene", Value: 1}, {Key: "createdAt", Value: -1}}},
        {Keys: bson.D{{Key: "expireAt", Value: 1}}, Options: options.Index().SetExpireAfterSeconds(0)},
    }); err != nil {
        return err
    }
    // friend_requests
    if err := createIndexes(ctx, db.Collection("friend_requests"), []mongo.IndexModel{
        {Keys: bson.D{{Key: "recipientId", Value: 1}, {Key: "status", Value: 1}, {Key: "createdAt", Value: -1}}},
        {Keys: bson.D{{Key: "requesterId", Value: 1}, {Key: "status", Value: 1}, {Key: "createdAt", Value: -1}}},
        {Keys: bson.D{{Key: "requesterId", Value: 1}, {Key: "recipientId", Value: 1}, {Key: "status", Value: 1}}},
    }); err != nil {
        return err
    }
    // friends
    if err := createIndexes(ctx, db.Collection("friends"), []mongo.IndexModel{
        {Keys: bson.D{{Key: "userA", Value: 1}, {Key: "userB", Value: 1}}, Options: options.Index().SetUnique(true)},
        {Keys: bson.D{{Key: "userA", Value: 1}}},
        {Keys: bson.D{{Key: "userB", Value: 1}}},
    }); err != nil {
        return err
    }
    // blocks
    if err := createIndexes(ctx, db.Collection("blocks"), []mongo.IndexModel{
        {Keys: bson.D{{Key: "userId", Value: 1}, {Key: "blockedUserId", Value: 1}}, Options: options.Index().SetUnique(true)},
        {Keys: bson.D{{Key: "userId", Value: 1}}},
    }); err != nil {
        return err
    }
    // groups & members
    if err := createIndexes(ctx, db.Collection("groups"), []mongo.IndexModel{
        {Keys: bson.D{{Key: "ownerId", Value: 1}, {Key: "createdAt", Value: -1}}},
    }); err != nil {
        return err
    }
    if err := createIndexes(ctx, db.Collection("group_members"), []mongo.IndexModel{
        {Keys: bson.D{{Key: "groupId", Value: 1}, {Key: "userId", Value: 1}}, Options: options.Index().SetUnique(true)},
        {Keys: bson.D{{Key: "userId", Value: 1}}},
        {Keys: bson.D{{Key: "groupId", Value: 1}}},
    }); err != nil {
        return err
    }
    // conversations & messages & counters
    if err := createIndexes(ctx, db.Collection("conversations"), []mongo.IndexModel{
        {Keys: bson.D{{Key: "conversationId", Value: 1}}, Options: options.Index().SetUnique(true)},
        {Keys: bson.D{{Key: "participants", Value: 1}}},
        {Keys: bson.D{{Key: "updatedAt", Value: -1}}},
    }); err != nil {
        return err
    }
    if err := createIndexes(ctx, db.Collection("messages"), []mongo.IndexModel{
        {Keys: bson.D{{Key: "conversationId", Value: 1}, {Key: "seq", Value: 1}}, Options: options.Index().SetUnique(true)},
        {Keys: bson.D{{Key: "conversationId", Value: 1}, {Key: "createdAt", Value: -1}}},
    }); err != nil {
        return err
    }
    if err := createIndexes(ctx, db.Collection("counters"), []mongo.IndexModel{
        {Keys: bson.D{{Key: "_id", Value: 1}}, Options: options.Index().SetUnique(true)},
    }); err != nil {
        return err
    }
    // theaters
    if err := createIndexes(ctx, db.Collection("theaters"), []mongo.IndexModel{
        {Keys: bson.D{{Key: "recruitId", Value: 1}}},
        {Keys: bson.D{{Key: "status", Value: 1}}},
    }); err != nil {
        return err
    }
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

