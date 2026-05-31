package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	MySQL  MySQLConfig  `yaml:"mysql"`
	Redis  RedisConfig  `yaml:"redis"`
	Auth   AuthConfig   `yaml:"auth"`
}

type AuthConfig struct {
	JWTSecret string `yaml:"jwt_secret"`
	TokenTTL  string `yaml:"token_ttl"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

type MySQLConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	Database        string `yaml:"database"`
	Username        string `yaml:"username"`
	Password        string `yaml:"password"`
	Charset         string `yaml:"charset"`
	ParseTime       bool   `yaml:"parse_time"`
	Loc             string `yaml:"loc"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	ConnMaxLifetime string `yaml:"conn_max_lifetime"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	PoolSize int    `yaml:"pool_size"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	cfg.applyEnvOverrides()
	cfg.setDefaults()

	return &cfg, nil
}

func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

func (c *MySQLConfig) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
		c.Charset,
		c.ParseTime,
		c.Loc,
	)
}

func (c *MySQLConfig) ConnMaxLifetimeDuration() time.Duration {
	d, err := time.ParseDuration(c.ConnMaxLifetime)
	if err != nil {
		return time.Hour
	}
	return d
}

func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func (c *Config) applyEnvOverrides() {
	if v := os.Getenv("CARD_SERVER_HOST"); v != "" {
		c.Server.Host = v
	}
	if v := os.Getenv("CARD_SERVER_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			c.Server.Port = port
		}
	}
	if v := os.Getenv("CARD_SERVER_MODE"); v != "" {
		c.Server.Mode = v
	}

	if v := os.Getenv("CARD_MYSQL_HOST"); v != "" {
		c.MySQL.Host = v
	}
	if v := os.Getenv("CARD_MYSQL_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			c.MySQL.Port = port
		}
	}
	if v := os.Getenv("CARD_MYSQL_DATABASE"); v != "" {
		c.MySQL.Database = v
	}
	if v := os.Getenv("CARD_MYSQL_USERNAME"); v != "" {
		c.MySQL.Username = v
	}
	if v := os.Getenv("CARD_MYSQL_PASSWORD"); v != "" {
		c.MySQL.Password = v
	}

	if v := os.Getenv("CARD_REDIS_HOST"); v != "" {
		c.Redis.Host = v
	}
	if v := os.Getenv("CARD_REDIS_PORT"); v != "" {
		if port, err := strconv.Atoi(v); v != "" && err == nil {
			c.Redis.Port = port
		}
	}
	if v := os.Getenv("CARD_REDIS_PASSWORD"); v != "" {
		c.Redis.Password = v
	}
	if v := os.Getenv("CARD_REDIS_DB"); v != "" {
		if db, err := strconv.Atoi(v); err == nil {
			c.Redis.DB = db
		}
	}

	if v := os.Getenv("CARD_JWT_SECRET"); v != "" {
		c.Auth.JWTSecret = v
	}
	if v := os.Getenv("CARD_TOKEN_TTL"); v != "" {
		c.Auth.TokenTTL = v
	}
}

func (c *Config) setDefaults() {
	if c.Server.Host == "" {
		c.Server.Host = "0.0.0.0"
	}
	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}
	if c.Server.Mode == "" {
		c.Server.Mode = "debug"
	}
	if c.MySQL.Charset == "" {
		c.MySQL.Charset = "utf8mb4"
	}
	if c.MySQL.Loc == "" {
		c.MySQL.Loc = "Local"
	}
	if c.MySQL.MaxOpenConns == 0 {
		c.MySQL.MaxOpenConns = 25
	}
	if c.MySQL.MaxIdleConns == 0 {
		c.MySQL.MaxIdleConns = 10
	}
	if c.MySQL.ConnMaxLifetime == "" {
		c.MySQL.ConnMaxLifetime = "1h"
	}
	if c.Redis.PoolSize == 0 {
		c.Redis.PoolSize = 10
	}
	if c.Auth.JWTSecret == "" {
		c.Auth.JWTSecret = "change-me-in-production"
	}
	if c.Auth.TokenTTL == "" {
		c.Auth.TokenTTL = "720h"
	}
}

func (c *AuthConfig) TokenDuration() time.Duration {
	d, err := time.ParseDuration(c.TokenTTL)
	if err != nil {
		return 720 * time.Hour
	}
	return d
}
