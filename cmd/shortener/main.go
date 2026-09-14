package main

import (
	"log"
	"net/http"
	"os"

	"url-shortener/internal/model"
	"url-shortener/internal/repository"
	"url-shortener/internal/service"
	"url-shortener/internal/handler"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("DB_DSN")

	if dsn == "" {
		log.Fatalf("DB_DSN environment variable is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&model.URL{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	log.Println("Migration completed successfully")

	repo := repository.NewURLRepository(db)
	svc := service.NewShortenerService(repo)
	h := handler.NewHTTPHandler(svc)

	router := h.InitRoutes()
	port := ":8080"

	log.Printf("Starting server on %s", port)

	if err := http.ListenAndServe(port, router); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}