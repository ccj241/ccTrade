package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/ccj241/cctrade/utils"
	"github.com/gin-gonic/gin"
)

func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logData := map[string]interface{}{
			"timestamp":  time.Now().Format(time.RFC3339),
			"panic":      recovered,
			"stack":      string(debug.Stack()),
			"client_ip":  c.ClientIP(),
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"user_agent": c.Request.UserAgent(),
		}

		if userID, exists := c.Get("user_id"); exists {
			logData["user_id"] = userID
		}

		logJSON, _ := json.Marshal(logData)
		log.Println("PANIC:", string(logJSON))

		c.JSON(http.StatusInternalServerError, utils.Response{
			Code:    500,
			Message: "服务器内部错误",
		})
	})
}

func HealthCheckMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/ping" {
			c.JSON(http.StatusOK, gin.H{
				"status":    "ok",
				"timestamp": time.Now().Unix(),
				"version":   "2.0.0",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
