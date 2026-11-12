package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// localhost only for now: avoids the Windows Firewall prompt.
	// Address will move to config (env) in the scaffold step, since Docker needs 0.0.0.0.
	if err := r.Run("localhost:8080"); err != nil {
		log.Fatal(err)
	}
}