package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type HealthHandler struct {
	DB    *gorm.DB
	Redis *redis.Client
}

func (h *HealthHandler) Check(c *gin.Context) {
	mysqlOK := "ok"
	if sqlDB, err := h.DB.DB(); err != nil {
		mysqlOK = err.Error()
	} else if err := sqlDB.Ping(); err != nil {
		mysqlOK = err.Error()
	}

	redisOK := "ok"
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()
	if err := h.Redis.Ping(ctx).Err(); err != nil {
		redisOK = err.Error()
	}

	status := http.StatusOK
	if mysqlOK != "ok" || redisOK != "ok" {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, gin.H{
		"status": "card-backend",
		"mysql":  mysqlOK,
		"redis":  redisOK,
	})
}
