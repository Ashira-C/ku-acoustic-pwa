package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Default to localhost for native dev (no Windows Firewall prompt).
	// In Docker, HTTP_ADDR=:8080 listens on all interfaces inside the container.
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = "localhost:8080"
	}

	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}