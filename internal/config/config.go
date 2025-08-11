package config

import (
    "fmt"
    "time"

    "github.com/spf13/viper"
)

var C Config

type Config struct {
    Server struct {
        Port int `mapstructure:"port"`
    } `mapstructure:"server"`
    JWT struct {
        Secret         string `mapstructure:"secret"`
        AccessTTLMin   int    `mapstructure:"access_ttl_minutes"`
        RefreshTTLDays int    `mapstructure:"refresh_ttl_days"`
    } `mapstructure:"jwt"`
    Mongo struct {
        URI      string `mapstructure:"uri"`
        Database string `mapstructure:"database"`
    } `mapstructure:"mongo"`
    SMS struct {
        Enabled  bool   `mapstructure:"enabled"`
        MockCode string `mapstructure:"mock_code"`
    } `mapstructure:"sms"`
}

func Load() error {
    v := viper.New()
    v.SetConfigType("yaml")
    v.SetConfigName("config")
    v.AddConfigPath("./configs")
    v.AddConfigPath(".")
    v.AutomaticEnv()

    v.SetDefault("server.port", 8080)
    v.SetDefault("jwt.access_ttl_minutes", 30)
    v.SetDefault("jwt.refresh_ttl_days", 14)

    if err := v.ReadInConfig(); err != nil {
        fmt.Printf("warning: using defaults/env, failed to read config: %v\n", err)
    }
    if err := v.Unmarshal(&C); err != nil {
        return err
    }
    return nil
}

func AccessTTL() time.Duration { return time.Duration(C.JWT.AccessTTLMin) * time.Minute }
func RefreshTTL() time.Duration { return time.Duration(C.JWT.RefreshTTLDays) * 24 * time.Hour }

