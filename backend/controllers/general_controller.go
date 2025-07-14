package controllers

import (
	"github.com/ccj241/cctrade/config"
	"github.com/ccj241/cctrade/models"
	"github.com/ccj241/cctrade/services"
	"github.com/ccj241/cctrade/utils"
	"github.com/gin-gonic/gin"
	"strconv"
)

type GeneralController struct {
	userService *services.UserService
}

func NewGeneralController() *GeneralController {
	return &GeneralController{
		userService: services.NewUserService(),
	}
}

func (gc *GeneralController) GetBalance(c *gin.Context) {
	userID := c.GetUint("user_id")

	apiKey, secretKey, err := gc.userService.GetUserAPIKeys(userID)
	if err != nil {
		utils.BadRequestResponse(c, "请先设置API密钥")
		return
	}

	binanceService, err := services.NewBinanceService(apiKey, secretKey)
	if err != nil {
		utils.InternalServerErrorResponse(c, "创建Binance服务失败: "+err.Error())
		return
	}

	account, err := binanceService.GetAccountInfo(c.Request.Context())
	if err != nil {
		utils.InternalServerErrorResponse(c, "获取账户信息失败: "+err.Error())
		return
	}

	// 尝试获取期货账户信息，但在测试网环境下容忍失败
	var futuresAccount interface{}
	futuresAccount, err = binanceService.GetFuturesAccountInfo(c.Request.Context())
	if err != nil {
		// 检查是否是测试网环境
		if config.AppConfig.Binance.TestNet {
			// 在测试网环境下，期货API可能不可用
			futuresAccount = map[string]interface{}{
				"error": "Futures API not available on testnet",
				"message": err.Error(),
			}
		} else {
			utils.InternalServerErrorResponse(c, "获取期货账户信息失败: "+err.Error())
			return
		}
	}

	response := map[string]interface{}{
		"spot":    account,
		"futures": futuresAccount,
	}

	utils.SuccessResponse(c, response)
}

func (gc *GeneralController) GetOrders(c *gin.Context) {
	userID := c.GetUint("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")
	symbol := c.Query("symbol")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	var orders []models.Order
	var total int64

	query := config.DB.Model(&models.Order{}).Where("user_id = ?", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if symbol != "" {
		query = query.Where("symbol = ?", utils.ToUpper(symbol))
	}

	if err := query.Count(&total).Error; err != nil {
		utils.InternalServerErrorResponse(c, "获取订单数量失败")
		return
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at desc").Find(&orders).Error; err != nil {
		utils.InternalServerErrorResponse(c, "获取订单列表失败")
		return
	}

	utils.PaginatedSuccessResponse(c, orders, total, page, limit)
}

func (gc *GeneralController) CreateOrder(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req struct {
		Symbol    string  `json:"symbol" binding:"required"`
		Side      string  `json:"side" binding:"required"`
		Type      string  `json:"type" binding:"required"`
		Quantity  float64 `json:"quantity" binding:"required"`
		Price     float64 `json:"price"`
		StopPrice float64 `json:"stop_price"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "请求参数无效: "+err.Error())
		return
	}

	if err := utils.ValidateSymbol(req.Symbol); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	if err := utils.ValidateQuantity(req.Quantity); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	if req.Price > 0 {
		if err := utils.ValidatePrice(req.Price); err != nil {
			utils.BadRequestResponse(c, err.Error())
			return
		}
	}

	apiKey, secretKey, err := gc.userService.GetUserAPIKeys(userID)
	if err != nil {
		utils.BadRequestResponse(c, "请先设置API密钥")
		return
	}

	binanceService, err := services.NewBinanceService(apiKey, secretKey)
	if err != nil {
		utils.InternalServerErrorResponse(c, "创建Binance服务失败: "+err.Error())
		return
	}

	order := &models.Order{
		UserID:        userID,
		Symbol:        utils.ToUpper(req.Symbol),
		Side:          models.OrderSide(req.Side),
		Type:          models.OrderType(req.Type),
		Quantity:      req.Quantity,
		Price:         req.Price,
		StopPrice:     req.StopPrice,
		ClientOrderID: utils.GenerateUUID(),
	}

	resp, err := binanceService.CreateSpotOrder(c.Request.Context(), order)
	if err != nil {
		utils.BadRequestResponse(c, "创建订单失败: "+err.Error())
		return
	}

	order.OrderID = strconv.FormatInt(resp.OrderID, 10)
	order.Status = models.OrderStatus(resp.Status)

	if err := config.DB.Create(order).Error; err != nil {
		utils.InternalServerErrorResponse(c, "保存订单失败")
		return
	}

	utils.SuccessWithMessage(c, "订单创建成功", order)
}

func (gc *GeneralController) CancelOrder(c *gin.Context) {
	userID := c.GetUint("user_id")
	orderIDStr := c.Param("order_id")

	var order models.Order
	if err := config.DB.Where("order_id = ? AND user_id = ?", orderIDStr, userID).First(&order).Error; err != nil {
		utils.NotFoundResponse(c, "订单不存在")
		return
	}

	if order.Status != models.OrderStatusNew && order.Status != models.OrderStatusPending {
		utils.BadRequestResponse(c, "订单状态不允许取消")
		return
	}

	apiKey, secretKey, err := gc.userService.GetUserAPIKeys(userID)
	if err != nil {
		utils.BadRequestResponse(c, "请先设置API密钥")
		return
	}

	binanceService, err := services.NewBinanceService(apiKey, secretKey)
	if err != nil {
		utils.InternalServerErrorResponse(c, "创建Binance服务失败: "+err.Error())
		return
	}

	_, err = binanceService.CancelSpotOrder(c.Request.Context(), order.Symbol, order.OrderID)
	if err != nil {
		utils.BadRequestResponse(c, "取消订单失败: "+err.Error())
		return
	}

	config.DB.Model(&order).Update("status", models.OrderStatusCanceled)

	utils.SuccessWithMessage(c, "订单取消成功", nil)
}

func (gc *GeneralController) BatchCancelOrders(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req struct {
		Symbol   string   `json:"symbol"`
		OrderIDs []string `json:"order_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "请求参数无效: "+err.Error())
		return
	}

	apiKey, secretKey, err := gc.userService.GetUserAPIKeys(userID)
	if err != nil {
		utils.BadRequestResponse(c, "请先设置API密钥")
		return
	}

	binanceService, err := services.NewBinanceService(apiKey, secretKey)
	if err != nil {
		utils.InternalServerErrorResponse(c, "创建Binance服务失败: "+err.Error())
		return
	}

	var successCount int
	var failedOrders []string

	for _, orderID := range req.OrderIDs {
		var order models.Order
		if err := config.DB.Where("order_id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
			failedOrders = append(failedOrders, orderID)
			continue
		}

		if order.Status != models.OrderStatusNew && order.Status != models.OrderStatusPending {
			failedOrders = append(failedOrders, orderID)
			continue
		}

		_, err := binanceService.CancelSpotOrder(c.Request.Context(), order.Symbol, order.OrderID)
		if err != nil {
			failedOrders = append(failedOrders, orderID)
			continue
		}

		config.DB.Model(&order).Update("status", models.OrderStatusCanceled)
		successCount++
	}

	response := map[string]interface{}{
		"success_count": successCount,
		"failed_orders": failedOrders,
		"total_orders":  len(req.OrderIDs),
	}

	utils.SuccessWithMessage(c, "批量取消订单完成", response)
}

func (gc *GeneralController) GetCancelledOrders(c *gin.Context) {
	userID := c.GetUint("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	var orders []models.Order
	var total int64

	query := config.DB.Model(&models.Order{}).Where("user_id = ? AND status = ?", userID, models.OrderStatusCanceled)

	if err := query.Count(&total).Error; err != nil {
		utils.InternalServerErrorResponse(c, "获取已取消订单数量失败")
		return
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("updated_at desc").Find(&orders).Error; err != nil {
		utils.InternalServerErrorResponse(c, "获取已取消订单列表失败")
		return
	}

	utils.PaginatedSuccessResponse(c, orders, total, page, limit)
}

func (gc *GeneralController) GetTradingSymbols(c *gin.Context) {
	userID := c.GetUint("user_id")

	apiKey, secretKey, err := gc.userService.GetUserAPIKeys(userID)
	if err != nil {
		utils.BadRequestResponse(c, "请先设置API密钥")
		return
	}

	binanceService, err := services.NewBinanceService(apiKey, secretKey)
	if err != nil {
		utils.InternalServerErrorResponse(c, "创建Binance服务失败: "+err.Error())
		return
	}

	symbols, err := binanceService.GetTradingSymbols(c.Request.Context())
	if err != nil {
		utils.InternalServerErrorResponse(c, "获取交易对信息失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, symbols)
}

func (gc *GeneralController) GetFuturesTradingSymbols(c *gin.Context) {
	userID := c.GetUint("user_id")

	apiKey, secretKey, err := gc.userService.GetUserAPIKeys(userID)
	if err != nil {
		utils.BadRequestResponse(c, "请先设置API密钥")
		return
	}

	binanceService, err := services.NewBinanceService(apiKey, secretKey)
	if err != nil {
		utils.InternalServerErrorResponse(c, "创建Binance服务失败: "+err.Error())
		return
	}

	symbols, err := binanceService.GetFuturesTradingSymbols(c.Request.Context())
	if err != nil {
		utils.InternalServerErrorResponse(c, "获取期货交易对信息失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, symbols)
}

func (gc *GeneralController) GetPrice(c *gin.Context) {
	symbol := c.Query("symbol")
	if symbol == "" {
		utils.BadRequestResponse(c, "交易对不能为空")
		return
	}

	if err := utils.ValidateSymbol(symbol); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	userID := c.GetUint("user_id")
	apiKey, secretKey, err := gc.userService.GetUserAPIKeys(userID)
	if err != nil {
		utils.BadRequestResponse(c, "请先设置API密钥")
		return
	}

	binanceService, err := services.NewBinanceService(apiKey, secretKey)
	if err != nil {
		utils.InternalServerErrorResponse(c, "创建Binance服务失败: "+err.Error())
		return
	}

	price, err := binanceService.GetPrice(c.Request.Context(), utils.ToUpper(symbol))
	if err != nil {
		utils.InternalServerErrorResponse(c, "获取价格失败: "+err.Error())
		return
	}

	response := map[string]interface{}{
		"symbol": utils.ToUpper(symbol),
		"price":  price,
	}

	utils.SuccessResponse(c, response)
}
