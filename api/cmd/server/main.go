package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Ashira-C/ku-acoustic-pwa/api/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	log.Printf("starting API (env=%s) on %s", cfg.AppEnv, cfg.HTTPAddr)
	if err := r.Run(cfg.HTTPAddr); err != nil {
		log.Fatal(err)
	}
}