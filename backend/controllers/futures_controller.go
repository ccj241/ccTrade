package controllers

import (
	"github.com/ccj241/cctrade/services"
	"github.com/ccj241/cctrade/utils"
	"github.com/gin-gonic/gin"
	"strconv"
)

type FuturesController struct {
	futuresService *services.FuturesService
}

func NewFuturesController() *FuturesController {
	return &FuturesController{
		futuresService: services.NewFuturesService(),
	}
}

func (fc *FuturesController) CreateFuturesStrategy(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "请求参数无效: "+err.Error())
		return
	}

	strategy, err := fc.futuresService.CreateFuturesStrategy(userID, req)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "期货策略创建成功", strategy)
}

func (fc *FuturesController) GetUserFuturesStrategies(c *gin.Context) {
	userID := c.GetUint("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	strategies, total, err := fc.futuresService.GetUserFuturesStrategies(userID, page, limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "获取期货策略列表失败")
		return
	}

	utils.PaginatedSuccessResponse(c, strategies, total, page, limit)
}

func (fc *FuturesController) GetFuturesStrategyByID(c *gin.Context) {
	userID := c.GetUint("user_id")
	strategyIDStr := c.Param("strategy_id")
	strategyID, err := strconv.ParseUint(strategyIDStr, 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "策略ID无效")
		return
	}

	strategy, err := fc.futuresService.GetFuturesStrategyByID(userID, uint(strategyID))
	if err != nil {
		utils.NotFoundResponse(c, "期货策略不存在")
		return
	}

	utils.SuccessResponse(c, strategy)
}

func (fc *FuturesController) UpdateFuturesStrategy(c *gin.Context) {
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

	if err := fc.futuresService.UpdateFuturesStrategy(userID, uint(strategyID), req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "期货策略更新成功", nil)
}

func (fc *FuturesController) ToggleFuturesStrategy(c *gin.Context) {
	userID := c.GetUint("user_id")
	strategyIDStr := c.Param("strategy_id")
	strategyID, err := strconv.ParseUint(strategyIDStr, 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "策略ID无效")
		return
	}

	if err := fc.futuresService.ToggleFuturesStrategy(userID, uint(strategyID)); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "期货策略状态切换成功", nil)
}

func (fc *FuturesController) DeleteFuturesStrategy(c *gin.Context) {
	userID := c.GetUint("user_id")
	strategyIDStr := c.Param("strategy_id")
	strategyID, err := strconv.ParseUint(strategyIDStr, 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "策略ID无效")
		return
	}

	if err := fc.futuresService.DeleteFuturesStrategy(userID, uint(strategyID)); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "期货策略删除成功", nil)
}

func (fc *FuturesController) GetUserPositions(c *gin.Context) {
	userID := c.GetUint("user_id")

	positions, err := fc.futuresService.GetUserPositions(userID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "获取持仓信息失败")
		return
	}

	utils.SuccessResponse(c, positions)
}

func (fc *FuturesController) UpdatePositions(c *gin.Context) {
	userID := c.GetUint("user_id")

	if err := fc.futuresService.UpdatePositions(userID); err != nil {
		utils.InternalServerErrorResponse(c, "更新持仓信息失败")
		return
	}

	utils.SuccessWithMessage(c, "持仓信息更新成功", nil)
}

func (fc *FuturesController) GetFuturesStats(c *gin.Context) {
	userID := c.GetUint("user_id")

	stats, err := fc.futuresService.GetFuturesStats(userID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "获取期货统计失败")
		return
	}

	utils.SuccessResponse(c, stats)
}
