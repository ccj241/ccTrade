package middleware

import (
	"github.com/ccj241/cctrade/config"
	"github.com/ccj241/cctrade/models"
	"github.com/ccj241/cctrade/utils"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.UnauthorizedResponse(c, "缺少认证头")
			c.Abort()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			utils.UnauthorizedResponse(c, "认证头格式错误")
			c.Abort()
			return
		}

		token := tokenParts[1]
		claims, err := utils.ValidateJWT(token)
		if err != nil {
			utils.UnauthorizedResponse(c, "无效的token")
			c.Abort()
			return
		}

		var user models.User
		if err := config.DB.First(&user, claims.UserID).Error; err != nil {
			utils.UnauthorizedResponse(c, "用户不存在")
			c.Abort()
			return
		}

		if user.Status != models.StatusActive {
			utils.UnauthorizedResponse(c, "用户账户已被禁用")
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("user", user)

		c.Next()
	}
}

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			utils.ForbiddenResponse(c, "权限不足")
			c.Abort()
			return
		}

		if role.(string) != string(models.RoleAdmin) {
			utils.ForbiddenResponse(c, "需要管理员权限")
			c.Abort()
			return
		}

		c.Next()
	}
}

func UserOwnershipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			utils.UnauthorizedResponse(c, "未认证")
			c.Abort()
			return
		}

		role, _ := c.Get("role")
		if role.(string) == string(models.RoleAdmin) {
			c.Next()
			return
		}

		requestedUserID := c.Query("user_id")
		if requestedUserID == "" {
			requestedUserID = c.Param("user_id")
		}

		if requestedUserID != "" {
			currentUserID := userID.(uint)
			if requestedUserID != strconv.FormatUint(uint64(currentUserID), 10) {
				utils.ForbiddenResponse(c, "只能访问自己的资源")
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
