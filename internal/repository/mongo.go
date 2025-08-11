package repository

import (
    "context"
    "time"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.uber.org/zap"

    "roleplay/internal/config"
)

var (
    mongoClient *mongo.Client
    mongoDB     *mongo.Database
)

// InitMongo 初始化全局 MongoDB 客户端与数据库句柄。
func InitMongo(ctx context.Context) error {
    client, err := mongo.NewClient(options.Client().ApplyURI(config.C.Mongo.URI))
    if err != nil {
        return err
    }
    ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
    defer cancel()
    if err := client.Connect(ctx); err != nil {
        return err
    }
    if err := client.Ping(ctx, nil); err != nil {
        return err
    }
    mongoClient = client
    mongoDB = client.Database(config.C.Mongo.Database)
    zap.L().Info("mongo connected", zap.String("db", config.C.Mongo.Database))
    return nil
}

// CloseMongo 优雅关闭 Mongo 客户端。
func CloseMongo(ctx context.Context) error {
    if mongoClient == nil {
        return nil
    }
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()
    return mongoClient.Disconnect(ctx)
}

// DB 返回默认数据库句柄。
func DB() *mongo.Database { return mongoDB }

