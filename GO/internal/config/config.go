package config

import (
    "fmt"
    "time"

    "github.com/spf13/viper"
)

// C holds the loaded configuration
var C Config

// Config represents application configuration.
type Config struct {
    Server struct {
        Port int `mapstructure:"port"`
    } `mapstructure:"server"`
    JWT struct {
        Secret          string `mapstructure:"secret"`
        AccessTTLMin    int    `mapstructure:"access_ttl_minutes"`
        RefreshTTLDays  int    `mapstructure:"refresh_ttl_days"`
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

// Load reads configuration from configs/config.yaml and env overrides.
func Load() error {
    v := viper.New()
    v.SetConfigType("yaml")
    v.SetConfigName("config")
    v.AddConfigPath("./configs")
    v.AddConfigPath(".")
    v.AutomaticEnv()

    // Sensible defaults
    v.SetDefault("server.port", 8080)
    v.SetDefault("jwt.access_ttl_minutes", 30)
    v.SetDefault("jwt.refresh_ttl_days", 14)
    v.SetDefault("sms.enabled", true)

    if err := v.ReadInConfig(); err != nil {
        // allow missing config if env fully provided
        fmt.Printf("warning: using defaults/env, failed to read config: %v\n", err)
    }
    if err := v.Unmarshal(&C); err != nil {
        return err
    }
    return nil
}

// AccessTTL returns the access token TTL as duration.
func AccessTTL() time.Duration { return time.Duration(C.JWT.AccessTTLMin) * time.Minute }

// RefreshTTL returns the refresh token TTL as duration.
func RefreshTTL() time.Duration { return time.Duration(C.JWT.RefreshTTLDays) * 24 * time.Hour }

