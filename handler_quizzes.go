package main

import (
	"net/http"
	"time"

	"github.com/Corogura/quizmaker/internal/auth"
	"github.com/Corogura/quizmaker/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerQuizzesCreate(c *gin.Context) {
	type parameters struct {
		Title string `json:"title"`
	}
	bearer, err := auth.GetBearerToken(c.Request.Header)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header"})
		return
	}
	userID, err := auth.ValidateJWT(bearer, cfg.jwtSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}
	var params parameters
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't decode parameters"})
		return
	}
	quizID := uuid.New().String()
	path := generatePath()
	err = cfg.db.CreateQuiz(c.Request.Context(), database.CreateQuizParams{
		ID:        quizID,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		Title:     params.Title,
		UserID:    userID.String(),
		Path:      path,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't create quiz"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"quiz_id": quizID, "path": path})
}

func (cfg *apiConfig) handlerQuestionsCreate(c *gin.Context) {
	path := c.Param("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path is required"})
		return
	}
	quiz, err := cfg.db.GetQuizIDFromPath(c.Request.Context(), path)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Quiz not found"})
		return
	}
	bearer, err := auth.GetBearerToken(c.Request.Header)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header"})
		return
	}
	userID, err := auth.ValidateJWT(bearer, cfg.jwtSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}
	if quiz.UserID != userID.String() {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to add questions to this quiz"})
		return
	}
	type parameters struct {
		Question string `json:"question"`
		Choice1  string `json:"choice1"`
		Choice2  string `json:"choice2"`
		Choice3  string `json:"choice3"`
		Choice4  string `json:"choice4"`
		Answer   int64  `json:"answer"`
	}
	var params parameters
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't decode parameters"})
		return
	}
	err = cfg.db.CreateQuizQuestions(c.Request.Context(), database.CreateQuizQuestionsParams{
		ID:           uuid.New().String(),
		QuizID:       quiz.ID,
		QuestionText: params.Question,
		Choice1:      params.Choice1,
		Choice2:      params.Choice2,
		Choice3:      params.Choice3,
		Choice4:      params.Choice4,
		Answer:       params.Answer,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't create question"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Question created successfully"})
}
