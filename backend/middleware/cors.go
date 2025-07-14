package middleware

import (
	"github.com/ccj241/cctrade/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"strings"
)

func CORSMiddleware() gin.HandlerFunc {
	origins := strings.Split(config.AppConfig.Server.CORSOrigins, ",")

	corsConfig := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "X-Total-Count"},
		AllowCredentials: true,
		MaxAge:           12 * 3600,
	}

	if len(origins) == 1 && origins[0] == "*" {
		corsConfig.AllowAllOrigins = true
		corsConfig.AllowCredentials = false
	} else {
		corsConfig.AllowOrigins = origins
	}

	return cors.New(corsConfig)
}
