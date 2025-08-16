package main

import (
	"net/http"
	"time"

	"github.com/Corogura/quizmaker/internal/auth"
	"github.com/Corogura/quizmaker/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUsersCreate(c *gin.Context) {
	type parameters struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var params parameters
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't decode parameters"})
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't hash password"})
		return
	}
	err = cfg.db.CreateUser(c.Request.Context(), database.CreateUserParams{
		ID:        uuid.New().String(),
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		Email:     params.Email,
		HashedPw:  hashedPassword,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't create user"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{})
}

func (cfg *apiConfig) handlerUsersGet(c *gin.Context) {
	// come back to this later
}
