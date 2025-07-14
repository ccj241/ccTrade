package controllers

import (
	"github.com/ccj241/cctrade/models"
	"github.com/ccj241/cctrade/services"
	"github.com/ccj241/cctrade/utils"
	"github.com/gin-gonic/gin"
	"strconv"
)

type AuthController struct {
	userService *services.UserService
}

func NewAuthController() *AuthController {
	return &AuthController{
		userService: services.NewUserService(),
	}
}

func (ac *AuthController) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "请求参数无效: "+err.Error())
		return
	}

	user, err := ac.userService.CreateUser(req.Username, req.Email, req.Password)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "注册成功，等待管理员审核", user)
}

func (ac *AuthController) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "请求参数无效: "+err.Error())
		return
	}

	user, token, err := ac.userService.LoginUser(req.Username, req.Password)
	if err != nil {
		utils.UnauthorizedResponse(c, err.Error())
		return
	}

	response := map[string]interface{}{
		"user":  user,
		"token": token,
	}

	utils.SuccessWithMessage(c, "登录成功", response)
}

func (ac *AuthController) GetProfile(c *gin.Context) {
	userID := c.GetUint("user_id")

	user, err := ac.userService.GetUserByID(userID)
	if err != nil {
		utils.NotFoundResponse(c, "用户不存在")
		return
	}

	stats, err := ac.userService.GetUserStats(userID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "获取用户统计失败")
		return
	}

	response := map[string]interface{}{
		"user":  user,
		"stats": stats,
	}

	utils.SuccessResponse(c, response)
}

func (ac *AuthController) UpdateProfile(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "请求参数无效: "+err.Error())
		return
	}

	if err := ac.userService.UpdateUserProfile(userID, req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "更新成功", nil)
}

func (ac *AuthController) ChangePassword(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "请求参数无效: "+err.Error())
		return
	}

	if err := ac.userService.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "密码修改成功", nil)
}

func (ac *AuthController) UpdateAPIKeys(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req struct {
		APIKey    string `json:"api_key" binding:"required"`
		SecretKey string `json:"secret_key" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "请求参数无效: "+err.Error())
		return
	}

	if err := ac.userService.UpdateUserAPIKeys(userID, req.APIKey, req.SecretKey); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "API密钥更新成功", nil)
}

// ValidateAPIKeys 验证API密钥是否有效
func (ac *AuthController) ValidateAPIKeys(c *gin.Context) {
	userID := c.GetUint("user_id")

	if err := ac.userService.ValidateUserAPIKeys(userID); err != nil {
		utils.BadRequestResponse(c, "API密钥验证失败: "+err.Error())
		return
	}

	utils.SuccessWithMessage(c, "API密钥验证成功", nil)
}

func (ac *AuthController) GetAllUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	users, total, err := ac.userService.GetAllUsers(page, limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "获取用户列表失败")
		return
	}

	utils.PaginatedSuccessResponse(c, users, total, page, limit)
}

func (ac *AuthController) ApproveUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "用户ID无效")
		return
	}

	if err := ac.userService.UpdateUserStatus(uint(userID), models.StatusActive); err != nil {
		utils.InternalServerErrorResponse(c, "审核用户失败")
		return
	}

	utils.SuccessWithMessage(c, "用户审核成功", nil)
}

func (ac *AuthController) UpdateUserStatus(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "用户ID无效")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "请求参数无效: "+err.Error())
		return
	}

	status := models.UserStatus(req.Status)
	if status != models.StatusActive && status != models.StatusDisabled && status != models.StatusPending {
		utils.BadRequestResponse(c, "无效的用户状态")
		return
	}

	if err := ac.userService.UpdateUserStatus(uint(userID), status); err != nil {
		utils.InternalServerErrorResponse(c, "更新用户状态失败")
		return
	}

	utils.SuccessWithMessage(c, "用户状态更新成功", nil)
}

func (ac *AuthController) UpdateUserRole(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "用户ID无效")
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "请求参数无效: "+err.Error())
		return
	}

	role := models.UserRole(req.Role)
	if role != models.RoleAdmin && role != models.RoleUser {
		utils.BadRequestResponse(c, "无效的用户角色")
		return
	}

	if err := ac.userService.UpdateUserRole(uint(userID), role); err != nil {
		utils.InternalServerErrorResponse(c, "更新用户角色失败")
		return
	}

	utils.SuccessWithMessage(c, "用户角色更新成功", nil)
}
