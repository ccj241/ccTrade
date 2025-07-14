package controllers

import (
	"github.com/ccj241/cctrade/services"
	"github.com/ccj241/cctrade/utils"
	"github.com/gin-gonic/gin"
	"strconv"
)

type StrategyController struct {
	strategyService *services.StrategyService
}

func NewStrategyController() *StrategyController {
	return &StrategyController{
		strategyService: services.NewStrategyService(),
	}
}

func (sc *StrategyController) CreateStrategy(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "请求参数无效: "+err.Error())
		return
	}

	strategy, err := sc.strategyService.CreateStrategy(userID, req)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "策略创建成功", strategy)
}

func (sc *StrategyController) GetUserStrategies(c *gin.Context) {
	userID := c.GetUint("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	strategies, total, err := sc.strategyService.GetUserStrategies(userID, page, limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "获取策略列表失败")
		return
	}

	utils.PaginatedSuccessResponse(c, strategies, total, page, limit)
}

func (sc *StrategyController) GetStrategyByID(c *gin.Context) {
	userID := c.GetUint("user_id")
	strategyIDStr := c.Param("strategy_id")
	strategyID, err := strconv.ParseUint(strategyIDStr, 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "策略ID无效")
		return
	}

	strategy, err := sc.strategyService.GetStrategyByID(userID, uint(strategyID))
	if err != nil {
		utils.NotFoundResponse(c, "策略不存在")
		return
	}

	utils.SuccessResponse(c, strategy)
}

func (sc *StrategyController) UpdateStrategy(c *gin.Context) {
	userID := c.GetUint("user_id")
	strategyIDStr := c.Param("strategy_id")
	strategyID, err := strconv.ParseUint(strategyIDStr, 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "策略ID无效")
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "请求参数无效: "+err.Error())
		return
	}

	if err := sc.strategyService.UpdateStrategy(userID, uint(strategyID), req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "策略更新成功", nil)
}

func (sc *StrategyController) ToggleStrategy(c *gin.Context) {
	userID := c.GetUint("user_id")
	strategyIDStr := c.Param("strategy_id")
	strategyID, err := strconv.ParseUint(strategyIDStr, 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "策略ID无效")
		return
	}

	if err := sc.strategyService.ToggleStrategy(userID, uint(strategyID)); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "策略状态切换成功", nil)
}

func (sc *StrategyController) DeleteStrategy(c *gin.Context) {
	userID := c.GetUint("user_id")
	strategyIDStr := c.Param("strategy_id")
	strategyID, err := strconv.ParseUint(strategyIDStr, 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "策略ID无效")
		return
	}

	if err := sc.strategyService.DeleteStrategy(userID, uint(strategyID)); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "策略删除成功", nil)
}

func (sc *StrategyController) GetStrategyStats(c *gin.Context) {
	userID := c.GetUint("user_id")
	strategyIDStr := c.Param("strategy_id")
	strategyID, err := strconv.ParseUint(strategyIDStr, 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "策略ID无效")
		return
	}

	stats, err := sc.strategyService.GetStrategyStats(userID, uint(strategyID))
	if err != nil {
		utils.NotFoundResponse(c, "策略不存在")
		return
	}

	utils.SuccessResponse(c, stats)
}
