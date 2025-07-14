package middleware

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"time"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func LoggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logData := map[string]interface{}{
			"timestamp":  param.TimeStamp.Format(time.RFC3339),
			"status":     param.StatusCode,
			"latency":    param.Latency.String(),
			"client_ip":  param.ClientIP,
			"method":     param.Method,
			"path":       param.Path,
			"user_agent": param.Request.UserAgent(),
			"error":      param.ErrorMessage,
		}

		logJSON, _ := json.Marshal(logData)
		return string(logJSON) + "\n"
	})
}

func RequestResponseLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		latency := time.Since(start)

		logData := map[string]interface{}{
			"timestamp":     time.Now().Format(time.RFC3339),
			"status":        c.Writer.Status(),
			"latency_ms":    latency.Milliseconds(),
			"client_ip":     c.ClientIP(),
			"method":        c.Request.Method,
			"path":          c.Request.URL.Path,
			"query":         c.Request.URL.RawQuery,
			"user_agent":    c.Request.UserAgent(),
			"request_body":  string(requestBody),
			"response_body": blw.body.String(),
		}

		if userID, exists := c.Get("user_id"); exists {
			logData["user_id"] = userID
		}

		if username, exists := c.Get("username"); exists {
			logData["username"] = username
		}

		logJSON, _ := json.Marshal(logData)
		log.Println(string(logJSON))
	}
}

func ErrorLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				logData := map[string]interface{}{
					"timestamp":  time.Now().Format(time.RFC3339),
					"error":      err.Error(),
					"type":       err.Type,
					"client_ip":  c.ClientIP(),
					"method":     c.Request.Method,
					"path":       c.Request.URL.Path,
					"user_agent": c.Request.UserAgent(),
				}

				if userID, exists := c.Get("user_id"); exists {
					logData["user_id"] = userID
				}

				logJSON, _ := json.Marshal(logData)
				log.Println(string(logJSON))
			}
		}
	}
}
