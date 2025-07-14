package controllers

import (
	"github.com/ccj241/cctrade/services"
	"github.com/ccj241/cctrade/utils"
	"github.com/gin-gonic/gin"
	"strconv"
)

type WithdrawalController struct {
	withdrawalService *services.WithdrawalService
}

func NewWithdrawalController() *WithdrawalController {
	return &WithdrawalController{
		withdrawalService: services.NewWithdrawalService(),
	}
}

func (wc *WithdrawalController) CreateWithdrawalRule(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "请求参数无效: "+err.Error())
		return
	}

	withdrawal, err := wc.withdrawalService.CreateWithdrawalRule(userID, req)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "提币规则创建成功", withdrawal)
}

func (wc *WithdrawalController) GetUserWithdrawals(c *gin.Context) {
	userID := c.GetUint("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	withdrawals, total, err := wc.withdrawalService.GetUserWithdrawals(userID, page, limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "获取提币规则列表失败")
		return
	}

	utils.PaginatedSuccessResponse(c, withdrawals, total, page, limit)
}

func (wc *WithdrawalController) GetWithdrawalByID(c *gin.Context) {
	userID := c.GetUint("user_id")
	withdrawalIDStr := c.Param("withdrawal_id")
	withdrawalID, err := strconv.ParseUint(withdrawalIDStr, 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "提币规则ID无效")
		return
	}

	withdrawal, err := wc.withdrawalService.GetWithdrawalByID(userID, uint(withdrawalID))
	if err != nil {
		utils.NotFoundResponse(c, "提币规则不存在")
		return
	}

	utils.SuccessResponse(c, withdrawal)
}

func (wc *WithdrawalController) UpdateWithdrawal(c *gin.Context) {
	userID := c.GetUint("user_id")
	withdrawalIDStr := c.Param("withdrawal_id")
	withdrawalID, err := strconv.ParseUint(withdrawalIDStr, 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "提币规则ID无效")
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "请求参数无效: "+err.Error())
		return
	}

	if err := wc.withdrawalService.UpdateWithdrawal(userID, uint(withdrawalID), req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "提币规则更新成功", nil)
}

func (wc *WithdrawalController) ToggleWithdrawal(c *gin.Context) {
	userID := c.GetUint("user_id")
	withdrawalIDStr := c.Param("withdrawal_id")
	withdrawalID, err := strconv.ParseUint(withdrawalIDStr, 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "提币规则ID无效")
		return
	}

	if err := wc.withdrawalService.ToggleWithdrawal(userID, uint(withdrawalID)); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "提币规则状态切换成功", nil)
}

func (wc *WithdrawalController) DeleteWithdrawal(c *gin.Context) {
	userID := c.GetUint("user_id")
	withdrawalIDStr := c.Param("withdrawal_id")
	withdrawalID, err := strconv.ParseUint(withdrawalIDStr, 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "提币规则ID无效")
		return
	}

	if err := wc.withdrawalService.DeleteWithdrawal(userID, uint(withdrawalID)); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "提币规则删除成功", nil)
}

func (wc *WithdrawalController) GetUserWithdrawalHistory(c *gin.Context) {
	userID := c.GetUint("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	histories, total, err := wc.withdrawalService.GetUserWithdrawalHistory(userID, page, limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "获取提币历史失败")
		return
	}

	utils.PaginatedSuccessResponse(c, histories, total, page, limit)
}

func (wc *WithdrawalController) SyncWithdrawalHistory(c *gin.Context) {
	userID := c.GetUint("user_id")

	if err := wc.withdrawalService.SyncWithdrawalHistory(userID); err != nil {
		utils.InternalServerErrorResponse(c, "同步提币历史失败")
		return
	}

	utils.SuccessWithMessage(c, "提币历史同步成功", nil)
}

func (wc *WithdrawalController) GetWithdrawalStats(c *gin.Context) {
	userID := c.GetUint("user_id")

	stats, err := wc.withdrawalService.GetWithdrawalStats(userID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "获取提币统计失败")
		return
	}

	utils.SuccessResponse(c, stats)
}
