package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Corogura/quizmaker/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type apiConfig struct {
	db        *database.Queries
	jwtSecret string
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		panic("Error loading .env file")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	port := fmt.Sprintf(":%s", os.Getenv("PORT"))
	if port == "" {
		port = ":8080" // Default port if not set
	}
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		panic("DATABASE_URL environment variable is not set")
	}
	db, err := sql.Open("libsql", dbURL)
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}
	defer db.Close()
	dbQueries := database.New(db)
	cfg := apiConfig{
		db:        dbQueries,
		jwtSecret: jwtSecret,
	}
	r := gin.Default()
	r.LoadHTMLFiles("static/quiz.html")
	// ---------- Register routes ----------
	r.POST("/users/create", cfg.handlerUsersCreate)
	r.POST("/users/login", cfg.handlerUsersLogin)
	r.PUT("/users/password", cfg.handlerUpdatePassword)
	r.GET("/users/refresh", cfg.handlerRefreshJWT)
	r.PUT("/users/revoke", cfg.handlerRevokeRefreshToken)
	r.GET("/users/validate", cfg.handlerValidateJWT)
	r.POST("/quizzes", cfg.handlerQuizzesCreate)
	r.POST("/quizzes/:path", cfg.handlerQuestionsCreate)
	r.DELETE("/quizzes/:path", cfg.handlerQuizzesDelete)
	r.DELETE("/quizzes/:path/questions/:question_number", cfg.handlerQuestionsDelete)
	r.PUT("/quizzes/:path", cfg.handlerUpdateQuizTitle)
	r.StaticFile("/", "./static/index.html")
	r.GET("/quizzes", cfg.handlerGetAllQuizzesForUser)
	r.GET("/quizzes/:path", cfg.handlerServeQuizPage)
	r.GET("/quizzes/:path/questions", cfg.handlerGetAllQuestionsInQuiz)
	r.GET("/quizzes/:path/owner", cfg.handlerChechOwnerOfQuiz)
	r.Static("/static", "./static")
	// ---------- End of routes ----------

	srv := &http.Server{
		Addr:    port,
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %s", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Server exiting gracefully")
}
