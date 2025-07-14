package controllers

import (
	"github.com/ccj241/cctrade/services"
	"github.com/ccj241/cctrade/utils"
	"github.com/gin-gonic/gin"
	"strconv"
)

type DualInvestmentController struct {
	dualInvestmentService *services.DualInvestmentService
}

func NewDualInvestmentController() *DualInvestmentController {
	return &DualInvestmentController{
		dualInvestmentService: services.NewDualInvestmentService(),
	}
}

func (dic *DualInvestmentController) GetDualInvestmentProducts(c *gin.Context) {
	products, err := dic.dualInvestmentService.GetDualInvestmentProducts()
	if err != nil {
		utils.InternalServerErrorResponse(c, "获取双币投资产品失败")
		return
	}

	utils.SuccessResponse(c, products)
}

func (dic *DualInvestmentController) CreateDualInvestmentStrategy(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "请求参数无效: "+err.Error())
		return
	}

	strategy, err := dic.dualInvestmentService.CreateDualInvestmentStrategy(userID, req)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "双币投资策略创建成功", strategy)
}

func (dic *DualInvestmentController) GetUserDualInvestmentStrategies(c *gin.Context) {
	userID := c.GetUint("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	strategies, total, err := dic.dualInvestmentService.GetUserDualInvestmentStrategies(userID, page, limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "获取双币投资策略列表失败")
		return
	}

	utils.PaginatedSuccessResponse(c, strategies, total, page, limit)
}

func (dic *DualInvestmentController) GetDualInvestmentStrategyByID(c *gin.Context) {
	userID := c.GetUint("user_id")
	strategyIDStr := c.Param("strategy_id")
	strategyID, err := strconv.ParseUint(strategyIDStr, 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "策略ID无效")
		return
	}

	strategy, err := dic.dualInvestmentService.GetDualInvestmentStrategyByID(userID, uint(strategyID))
	if err != nil {
		utils.NotFoundResponse(c, "双币投资策略不存在")
		return
	}

	utils.SuccessResponse(c, strategy)
}

func (dic *DualInvestmentController) UpdateDualInvestmentStrategy(c *gin.Context) {
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

	if err := dic.dualInvestmentService.UpdateDualInvestmentStrategy(userID, uint(strategyID), req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "双币投资策略更新成功", nil)
}

func (dic *DualInvestmentController) ToggleDualInvestmentStrategy(c *gin.Context) {
	userID := c.GetUint("user_id")
	strategyIDStr := c.Param("strategy_id")
	strategyID, err := strconv.ParseUint(strategyIDStr, 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "策略ID无效")
		return
	}

	if err := dic.dualInvestmentService.ToggleDualInvestmentStrategy(userID, uint(strategyID)); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "双币投资策略状态切换成功", nil)
}

func (dic *DualInvestmentController) DeleteDualInvestmentStrategy(c *gin.Context) {
	userID := c.GetUint("user_id")
	strategyIDStr := c.Param("strategy_id")
	strategyID, err := strconv.ParseUint(strategyIDStr, 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "策略ID无效")
		return
	}

	if err := dic.dualInvestmentService.DeleteDualInvestmentStrategy(userID, uint(strategyID)); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "双币投资策略删除成功", nil)
}

func (dic *DualInvestmentController) GetUserDualInvestmentOrders(c *gin.Context) {
	userID := c.GetUint("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	orders, total, err := dic.dualInvestmentService.GetUserDualInvestmentOrders(userID, page, limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "获取双币投资订单失败")
		return
	}

	utils.PaginatedSuccessResponse(c, orders, total, page, limit)
}

func (dic *DualInvestmentController) GetDualInvestmentStats(c *gin.Context) {
	userID := c.GetUint("user_id")

	stats, err := dic.dualInvestmentService.GetDualInvestmentStats(userID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "获取双币投资统计失败")
		return
	}

	utils.SuccessResponse(c, stats)
}

func (dic *DualInvestmentController) SyncDualInvestmentProducts(c *gin.Context) {
	if err := dic.dualInvestmentService.SyncDualInvestmentProducts(); err != nil {
		utils.InternalServerErrorResponse(c, "同步双币投资产品失败")
		return
	}

	utils.SuccessWithMessage(c, "双币投资产品同步成功", nil)
}
