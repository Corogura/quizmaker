package main

import (
	"database/sql"
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

func (cfg *apiConfig) handlerUsersLogin(c *gin.Context) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var params parameters
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't decode parameters"})
		return
	}
	user, err := cfg.db.GetUserByEmail(c.Request.Context(), params.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}
	if err := auth.CheckPasswordHash(user.HashedPw, params.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}
	token, err := auth.MakeJWT(uuid.MustParse(user.ID), cfg.jwtSecret, 24*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't create token"})
		return
	}
	rt, _ := auth.MakeRefreshToken()
	rToken, err := cfg.db.CreateRefreshToken(c.Request.Context(), database.CreateRefreshTokenParams{
		Token:     rt,
		UserID:    user.ID,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		ExpiresAt: time.Now().UTC().Add(720 * time.Hour).Format(time.RFC3339),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't create refresh token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":            user.ID,
		"email":         user.Email,
		"created_at":    user.CreatedAt,
		"updated_at":    user.UpdatedAt,
		"token":         token,
		"refresh_token": rToken.Token,
		"expires_in":    720 * 60 * 60, // 720 hours in seconds
	})
}

func (cfg *apiConfig) handlerRefreshJWT(c *gin.Context) {
	tokenString, err := auth.GetBearerToken(c.Request.Header)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}
	dbUser, err := cfg.db.GetUserFromRefreshToken(c.Request.Context(), tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}
	accessToken, err := auth.MakeJWT(uuid.MustParse(dbUser.ID), cfg.jwtSecret, 24*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't create access token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token":      accessToken,
		"expires_in": 24 * 60 * 60, // 24 hours in seconds
	})
}

func (cfg *apiConfig) handlerRevokeRefreshToken(c *gin.Context) {
	tokenString, err := auth.GetBearerToken(c.Request.Header)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}
	err = cfg.db.RevokeRefreshToken(c.Request.Context(), database.RevokeRefreshTokenParams{
		Token:     tokenString,
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		RevokedAt: sql.NullString{
			String: time.Now().UTC().Format(time.RFC3339),
			Valid:  true,
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't revoke refresh token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Refresh token revoked successfully"})
}

func (cfg *apiConfig) handlerUpdatePassword(c *gin.Context) {
	type parameters struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	var params parameters
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't decode parameters"})
		return
	}
	bearerToken, err := auth.GetBearerToken(c.Request.Header)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}
	userID, err := auth.ValidateJWT(bearerToken, cfg.jwtSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user, err := cfg.db.GetUser(c.Request.Context(), userID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't find user"})
		return
	}
	if err := auth.CheckPasswordHash(user.HashedPw, params.CurrentPassword); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
		return
	}

	hashedPassword, err := auth.HashPassword(params.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't hash password"})
		return
	}

	err = cfg.db.UpdatePassword(c.Request.Context(), database.UpdatePasswordParams{
		ID:        userID.String(),
		HashedPw:  hashedPassword,
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't update password"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}
