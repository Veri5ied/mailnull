package main

import (
	"mailnull/api/handlers"
	"mailnull/api/internal/config"
	"mailnull/api/internal/logger"
	"mailnull/api/internal/verifier"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {

	cfg := config.Load()
	log := logger.Init()

	log.Info("Starting MailNull Verification Engine",
		"port", cfg.Port,
		"mode", cfg.Mode,
	)

	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "engine online",
			"mode":   cfg.Mode,
		})
	})

	v := verifier.New(cfg, log)
	wp := verifier.NewWorkerPool(v, 10, 100)
	wp.Start()
	defer wp.Wait()

	verifyHandler := handlers.NewVerifyHandler(wp)

	r.Use(corsMiddleware())

	v1 := r.Group("/v1/mailnull")
	{
		v1.POST("/verify", verifyHandler.Verify)
	}

	r.Run(":" + cfg.Port)
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
