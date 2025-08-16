package main

import (
	"database/sql"
	"net/http"
	"os"

	"github.com/Corogura/quizmaker/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type apiConfig struct {
	db *database.Queries
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		panic("Error loading .env file")
	}

	port := os.Getenv("PORT")
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
	dbQueries := database.New(db)
	cfg := apiConfig{
		db: dbQueries,
	}
	r := gin.Default()
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"test": "test",
		})
	})
	r.POST("/users", cfg.handlerUsersCreate)
	r.Run()
}
