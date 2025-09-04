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
	if quiz.DeletedAt.Valid {
		c.JSON(http.StatusGone, gin.H{"error": "Quiz has been deleted"})
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
	questionCount, err := cfg.db.GetQuestionCountInQuiz(c.Request.Context(), quiz.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't retrieve question count"})
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
		ID:             uuid.New().String(),
		QuizID:         quiz.ID,
		QuestionNumber: questionCount + 1, // Increment the question number
		QuestionText:   params.Question,
		Choice1:        params.Choice1,
		Choice2:        params.Choice2,
		Choice3:        params.Choice3,
		Choice4:        params.Choice4,
		Answer:         params.Answer,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't create question"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Question created successfully"})
}

func (cfg *apiConfig) handlerQuizzesDelete(c *gin.Context) {
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
	if quiz.DeletedAt.Valid {
		c.JSON(http.StatusGone, gin.H{"error": "Quiz has already been deleted"})
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
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to delete this quiz"})
		return
	}
	err = cfg.db.DeleteQuiz(c.Request.Context(), database.DeleteQuizParams{
		ID: quiz.ID,
		DeletedAt: sql.NullString{
			String: time.Now().UTC().Format(time.RFC3339),
			Valid:  true,
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't delete quiz"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Quiz deleted successfully"})
}

func (cfg *apiConfig) handlerQuestionsDelete(c *gin.Context) {
	if c.Param("question_number") == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Question ID is required"})
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
	quiz, err := cfg.db.GetQuizIDFromPath(c.Request.Context(), c.Param("path"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Quiz not found"})
		return
	}
	if quiz.UserID != userID.String() {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to delete this question"})
		return
	}
	question, err := cfg.db.GetQuestionFromQuestionNumber(c.Request.Context(), database.GetQuestionFromQuestionNumberParams{
		ID:     c.Param("question_number"),
		QuizID: quiz.ID,
	})
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't retrieve question"})
		return
	}
	if question.DeletedAt.Valid {
		c.JSON(http.StatusGone, gin.H{"error": "Question has already been deleted"})
		return
	}
	err = cfg.db.DeleteQuizQuestion(c.Request.Context(), database.DeleteQuizQuestionParams{
		ID: question.ID,
		DeletedAt: sql.NullString{
			String: time.Now().UTC().Format(time.RFC3339),
			Valid:  true,
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't delete question"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Question deleted successfully"})
}

func (cfg *apiConfig) handlerUpdateQuizTitle(c *gin.Context) {
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
	if quiz.DeletedAt.Valid {
		c.JSON(http.StatusGone, gin.H{"error": "Quiz has been deleted"})
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
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to update this quiz"})
		return
	}
	type parameters struct {
		NewTitle string `json:"new_title"`
	}
	var params parameters
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid parameters"})
		return
	}
	err = cfg.db.UpdateQuizTitle(c.Request.Context(), database.UpdateQuizTitleParams{
		ID:        quiz.ID,
		Title:     params.NewTitle,
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't update quiz title"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Quiz title updated successfully"})
}

func (cfg *apiConfig) handlerGetAllQuizzesForUser(c *gin.Context) {
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
	quizzes, err := cfg.db.GetAllQuizzesByUserID(c.Request.Context(), userID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't retrieve quizzes"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"quizzes": quizzes})
}

func (cfg *apiConfig) handlerGetAllQuestionsInQuiz(c *gin.Context) {
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
	if quiz.DeletedAt.Valid {
		c.JSON(http.StatusGone, gin.H{"error": "Quiz has been deleted"})
		return
	}
	questions, err := cfg.db.GetAllQuestionsInQuiz(c.Request.Context(), quiz.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't retrieve questions"})
		return
	}
	type QuestionWithChoices struct {
		ID           string `json:"id"`
		QuestionText string `json:"question_text"`
		Choices      []struct {
			ChoiceText string `json:"choice_text"`
			IsCorrect  bool   `json:"is_correct"`
		} `json:"choices"`
	}
	var formattedQuestions []QuestionWithChoices
	for _, q := range questions {
		formattedQuestion := QuestionWithChoices{
			ID:           q.ID,
			QuestionText: q.QuestionText,
			Choices: []struct {
				ChoiceText string `json:"choice_text"`
				IsCorrect  bool   `json:"is_correct"`
			}{},
		}
		choices := []struct {
			ChoiceText string `json:"choice_text"`
			IsCorrect  bool   `json:"is_correct"`
		}{
			{ChoiceText: q.Choice1, IsCorrect: q.Answer == 1},
			{ChoiceText: q.Choice2, IsCorrect: q.Answer == 2},
			{ChoiceText: q.Choice3, IsCorrect: q.Answer == 3},
			{ChoiceText: q.Choice4, IsCorrect: q.Answer == 4},
		}
		formattedQuestion.Choices = choices
		formattedQuestions = append(formattedQuestions, formattedQuestion)
	}
	c.JSON(http.StatusOK, gin.H{"questions": formattedQuestions})
}

func (cfg *apiConfig) handlerServeQuizPage(c *gin.Context) {
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
	if quiz.DeletedAt.Valid {
		c.JSON(http.StatusGone, gin.H{"error": "Quiz has been deleted"})
		return
	}
	c.HTML(http.StatusOK, "quiz.html", gin.H{
		"title": quiz.Title,
	})
}

func (cfg *apiConfig) handlerChechOwnerOfQuiz(c *gin.Context) {
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
	if quiz.DeletedAt.Valid {
		c.JSON(http.StatusGone, gin.H{"error": "Quiz has been deleted"})
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
		c.JSON(http.StatusForbidden, gin.H{"is_owner": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"is_owner": true})
}
