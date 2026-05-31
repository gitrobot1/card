package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/time/card/backend/internal/middleware"
	"github.com/time/card/backend/internal/service"
)

type AuthHandler struct {
	Auth *service.AuthService
}

type loginRequest struct {
	Username string `json:"username"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	result, err := h.Auth.LoginByUsername(req.Username)
	if err != nil {
		if errors.Is(err, service.ErrInvalidUsername) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "username must be 2-16 chars and contain only letters, numbers, underscore or Chinese",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := c.Get(middleware.ContextUserIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := h.Auth.GetUserByID(userID.(uint64))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}
