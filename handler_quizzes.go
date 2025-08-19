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
		Title        string `json:"title"`
		QuestionText string `json:"question_text"`
		Choice1      string `json:"choice1"`
		Choice2      string `json:"choice2"`
		Choice3      string `json:"choice3"`
		Choice4      string `json:"choice4"`
		Answer       int64  `json:"answer"`
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
	err = cfg.db.CreateQuizQuestions(c.Request.Context(), database.CreateQuizQuestionsParams{
		ID:           uuid.New().String(),
		QuizID:       quizID,
		QuestionText: params.QuestionText,
		Choice1:      params.Choice1,
		Choice2:      params.Choice2,
		Choice3:      params.Choice3,
		Choice4:      params.Choice4,
		Answer:       params.Answer,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't create quiz questions"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"quiz_id": quizID, "path": path})
}
