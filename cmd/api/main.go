package main

import (
	"log"

	"rkmin/internal/config"
	"rkmin/internal/database"
	delivery "rkmin/internal/delivery/http"
	"rkmin/internal/provcity"
	"rkmin/internal/repository"
	"rkmin/internal/usecase"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	db, err := database.Open(cfg.MySQLDSN)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}

	repo := repository.New(db)
	uploader, err := delivery.NewUploader(cfg.UploadDir)
	if err != nil {
		log.Fatalf("create upload dir: %v", err)
	}
	uc := usecase.New(repo, uploader.Save)
	jwtSvc := delivery.NewJWTService(cfg.JWTSecret, repo)
	handler := delivery.NewHandler(uc, jwtSvc, provcity.New(cfg.ProvAPIURL))

	r := gin.Default()
	r.Static("/uploads", cfg.UploadDir)
	handler.RegisterRoutes(r, cfg.BasePath)

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
