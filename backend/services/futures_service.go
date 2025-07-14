package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/ccj241/cctrade/config"
	"github.com/ccj241/cctrade/models"
	"github.com/ccj241/cctrade/utils"
	"gorm.io/gorm"
)

type FuturesService struct {
	db          *gorm.DB
	userService *UserService
}

func NewFuturesService() *FuturesService {
	return &FuturesService{
		db:          config.DB,
		userService: NewUserService(),
	}
}

func (fs *FuturesService) CreateFuturesStrategy(userID uint, strategyData map[string]interface{}) (*models.FuturesStrategy, error) {
	strategy := &models.FuturesStrategy{
		UserID: userID,
	}

	if name, ok := strategyData["name"].(string); ok {
		strategy.Name = name
	} else {
		return nil, errors.New("策略名称不能为空")
	}

	if symbol, ok := strategyData["symbol"].(string); ok {
		if err := utils.ValidateSymbol(symbol); err != nil {
			return nil, err
		}
		strategy.Symbol = utils.ToUpper(symbol)
	} else {
		return nil, errors.New("交易对不能为空")
	}

	if strategyType, ok := strategyData["type"].(string); ok {
		strategy.Type = models.StrategyType(strategyType)
	} else {
		return nil, errors.New("策略类型不能为空")
	}

	if side, ok := strategyData["side"].(string); ok {
		strategy.Side = models.OrderSide(side)
	} else {
		return nil, errors.New("交易方向不能为空")
	}

	// 保证金本值
	if marginAmount, ok := strategyData["margin_amount"].(float64); ok {
		if marginAmount <= 0 {
			return nil, errors.New("保证金金额必须大于0")
		}
		strategy.MarginAmount = marginAmount
	} else {
		return nil, errors.New("保证金金额不能为空")
	}

	// 触发价格
	if price, ok := strategyData["price"].(float64); ok && price > 0 {
		strategy.Price = price
	} else {
		return nil, errors.New("触发价格不能为空")
	}

	// 万分比浮动
	if floatBP, ok := strategyData["float_basis_points"].(float64); ok {
		if floatBP < 0 {
			return nil, errors.New("万分比浮动不能为负数")
		}
		strategy.FloatBasisPoints = floatBP
	}

	// 止盈万分比
	if takeProfitBP, ok := strategyData["take_profit_bp"].(float64); ok && takeProfitBP >= 0 {
		strategy.TakeProfitBP = int(takeProfitBP)
	}

	// 止损万分比
	if stopLossBP, ok := strategyData["stop_loss_bp"].(float64); ok && stopLossBP >= 0 {
		strategy.StopLossBP = int(stopLossBP)
	}

	// 杠杆倍数
	if leverage, ok := strategyData["leverage"].(float64); ok {
		if leverage < 1 || leverage > 20 {
			return nil, errors.New("杠杆倍数必须在1-20之间")
		}
		strategy.Leverage = int(leverage)
	} else {
		strategy.Leverage = 8
	}

	if marginType, ok := strategyData["margin_type"].(string); ok {
		strategy.MarginType = models.MarginType(marginType)
	} else {
		strategy.MarginType = models.MarginTypeCross
	}

	if autoRestart, ok := strategyData["auto_restart"].(bool); ok {
		strategy.AutoRestart = autoRestart
	}

	if config, ok := strategyData["config"].(map[string]interface{}); ok {
		strategy.Config = models.StrategyConfig(config)
	}

	if err := fs.db.Create(strategy).Error; err != nil {
		return nil, err
	}

	return strategy, nil
}

func (fs *FuturesService) GetUserFuturesStrategies(userID uint, page, limit int) ([]models.FuturesStrategy, int64, error) {
	var strategies []models.FuturesStrategy
	var total int64

	query := fs.db.Model(&models.FuturesStrategy{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Find(&strategies).Error; err != nil {
		return nil, 0, err
	}

	return strategies, total, nil
}

func (fs *FuturesService) GetFuturesStrategyByID(userID, strategyID uint) (*models.FuturesStrategy, error) {
	var strategy models.FuturesStrategy
	if err := fs.db.Where("id = ? AND user_id = ?", strategyID, userID).First(&strategy).Error; err != nil {
		return nil, err
	}
	return &strategy, nil
}

func (fs *FuturesService) UpdateFuturesStrategy(userID, strategyID uint, updates map[string]interface{}) error {
	allowedFields := []string{"name", "is_active", "auto_restart", "take_profit", "stop_loss", "leverage", "margin_type", "config"}
	filteredUpdates := make(map[string]interface{})

	for field, value := range updates {
		if utils.Contains(allowedFields, field) {
			if field == "leverage" {
				leverage, ok := value.(float64)
				if !ok {
					return errors.New("杠杆倍数必须是数字")
				}
				if leverage < 1 || leverage > 125 {
					return errors.New("杠杆倍数必须在1-125之间")
				}
			}
			filteredUpdates[field] = value
		}
	}

	if len(filteredUpdates) == 0 {
		return errors.New("没有有效的更新字段")
	}

	return fs.db.Model(&models.FuturesStrategy{}).Where("id = ? AND user_id = ?", strategyID, userID).Updates(filteredUpdates).Error
}

func (fs *FuturesService) ToggleFuturesStrategy(userID, strategyID uint) error {
	var strategy models.FuturesStrategy
	if err := fs.db.Where("id = ? AND user_id = ?", strategyID, userID).First(&strategy).Error; err != nil {
		return err
	}

	return fs.db.Model(&strategy).Update("is_active", !strategy.IsActive).Error
}

func (fs *FuturesService) DeleteFuturesStrategy(userID, strategyID uint) error {
	return fs.db.Where("id = ? AND user_id = ?", strategyID, userID).Delete(&models.FuturesStrategy{}).Error
}

func (fs *FuturesService) ExecuteFuturesStrategy(strategy *models.FuturesStrategy) error {
	apiKey, secretKey, err := fs.userService.GetUserAPIKeys(strategy.UserID)
	if err != nil {
		return err
	}

	binanceService, err := NewBinanceService(apiKey, secretKey)
	if err != nil {
		return err
	}

	if err := binanceService.SetFuturesLeverage(context.Background(), strategy.Symbol, strategy.Leverage); err != nil {
		log.Printf("设置杠杆失败: %v", err)
	}

	if err := binanceService.SetFuturesMarginType(context.Background(), strategy.Symbol, strategy.MarginType); err != nil {
		log.Printf("设置保证金模式失败: %v", err)
	}

	switch strategy.Type {
	case models.StrategySimple:
		return fs.executeSimpleFuturesStrategy(strategy, binanceService)
	case models.StrategyIceberg:
		return fs.executeFuturesIcebergStrategy(strategy, binanceService)
	case models.StrategySlowIceberg:
		return fs.executeSlowFuturesIcebergStrategy(strategy, binanceService)
	default:
		return fmt.Errorf("不支持的期货策略类型: %s", strategy.Type)
	}
}

func (fs *FuturesService) executeSimpleFuturesStrategy(strategy *models.FuturesStrategy, binanceService *BinanceService) error {
	currentPrice, err := binanceService.GetFuturesPrice(context.Background(), strategy.Symbol)
	if err != nil {
		return err
	}

	// 检查是否触发
	shouldExecute := false
	if strategy.Price > 0 {
		if strategy.Side == models.OrderSideBuy && currentPrice <= strategy.Price {
			// 做多：当前价格 <= 触发价格
			shouldExecute = true
		} else if strategy.Side == models.OrderSideSell && currentPrice >= strategy.Price {
			// 做空：当前价格 >= 触发价格
			shouldExecute = true
		}
	} else {
		shouldExecute = true
	}

	if !shouldExecute {
		return nil
	}

	// 计算实际下单价格（应用万分比浮动）
	actualOrderPrice := currentPrice
	if strategy.FloatBasisPoints > 0 {
		floatRate := float64(strategy.FloatBasisPoints) / 10000.0
		if strategy.Side == models.OrderSideBuy {
			// 做多时，万分比向下浮动
			actualOrderPrice = currentPrice * (1 - floatRate)
		} else {
			// 做空时，万分比向上浮动
			actualOrderPrice = currentPrice * (1 + floatRate)
		}
	}

	// 计算实际下单数量
	orderValue := strategy.MarginAmount * float64(strategy.Leverage)
	orderQuantity := orderValue / actualOrderPrice

	// 根据做多/做空设置持仓方向
	positionSide := models.PositionSideBoth
	if strategy.Side == models.OrderSideBuy {
		positionSide = models.PositionSideLong
	} else {
		positionSide = models.PositionSideShort
	}

	order := &models.FuturesOrder{
		UserID:        strategy.UserID,
		StrategyID:    &strategy.ID,
		Symbol:        strategy.Symbol,
		Side:          strategy.Side,
		PositionSide:  positionSide,
		Type:          models.OrderTypeLimit,
		Quantity:      orderQuantity,
		Price:         actualOrderPrice,
		ClientOrderID: utils.GenerateUUID(),
	}

	if strategy.Price > 0 {
		order.Type = models.OrderTypeLimit
	}

	resp, err := binanceService.CreateFuturesOrder(context.Background(), order)
	if err != nil {
		return err
	}

	order.OrderID = strconv.FormatInt(resp.OrderID, 10)
	order.Status = models.OrderStatus(resp.Status)

	if err := fs.db.Create(order).Error; err != nil {
		log.Printf("保存期货订单失败: %v", err)
	}

	if strategy.TakeProfitBP > 0 || strategy.StopLossBP > 0 {
		go fs.createStopLossAndTakeProfit(strategy, binanceService, actualOrderPrice, orderQuantity)
	}

	fs.db.Model(strategy).Update("is_completed", true)

	return nil
}

func (fs *FuturesService) executeFuturesIcebergStrategy(strategy *models.FuturesStrategy, binanceService *BinanceService) error {
	config := strategy.Config
	if config == nil {
		config = make(map[string]interface{})
	}
	
	// 获取层数配置，默认10层
	layers := 10
	if configValue, exists := config["layers"]; exists && configValue != nil {
		if layersConfig, ok := configValue.(float64); ok {
			layers = int(layersConfig)
			if layers < 5 || layers > 10 {
				layers = 10
			}
		}
	}
	
	// 获取超时时间（分钟），默认5分钟
	timeoutMinutes := 5
	if configValue, exists := config["timeout_minutes"]; exists && configValue != nil {
		if timeoutConfig, ok := configValue.(float64); ok {
			timeoutMinutes = int(timeoutConfig)
		}
	}
	
	// 获取自定义的层级浮动比例
	var customPriceFloats []float64
	if configValue, exists := config["layer_price_floats"]; exists && configValue != nil {
		if floatsInterface, ok := configValue.([]interface{}); ok {
			customPriceFloats = make([]float64, 0, len(floatsInterface))
			for _, v := range floatsInterface {
				if f, ok := v.(float64); ok {
					customPriceFloats = append(customPriceFloats, f)
				}
			}
		}
	}
	
	// 获取首单浮动比例
	firstLayerFloat := strategy.FloatBasisPoints
	if configValue, exists := config["first_layer_float"]; exists && configValue != nil {
		if firstFloat, ok := configValue.(float64); ok {
			firstLayerFloat = firstFloat
		}
	}
	
	// 获取当前价格
	currentPrice, err := binanceService.GetFuturesPrice(context.Background(), strategy.Symbol)
	if err != nil {
		return err
	}
	
	// 检查是否触发
	shouldExecute := false
	if strategy.Price > 0 {
		if strategy.Side == models.OrderSideBuy && currentPrice <= strategy.Price {
			shouldExecute = true
		} else if strategy.Side == models.OrderSideSell && currentPrice >= strategy.Price {
			shouldExecute = true
		}
	} else {
		shouldExecute = true
	}
	
	if !shouldExecute {
		return nil
	}
	
	// 获取订单簿深度，获取买1价或卖1价
	basePrice := currentPrice
	if depth, err := binanceService.GetFuturesDepth(context.Background(), strategy.Symbol, 5); err == nil {
		if strategy.Side == models.OrderSideBuy && len(depth.Bids) > 0 {
			// 做多时使用买1价
			bidPrice, _ := strconv.ParseFloat(depth.Bids[0].Price, 64)
			if bidPrice > 0 {
				basePrice = bidPrice
			}
		} else if strategy.Side == models.OrderSideSell && len(depth.Asks) > 0 {
			// 做空时使用卖1价
			askPrice, _ := strconv.ParseFloat(depth.Asks[0].Price, 64)
			if askPrice > 0 {
				basePrice = askPrice
			}
		}
	}
	
	// 计算总数量
	totalOrderValue := strategy.MarginAmount * float64(strategy.Leverage)
	totalQuantity := totalOrderValue / basePrice
	
	// 获取层级配置
	quantityRatios := fs.getQuantityRatios(layers)
	priceFloats := fs.getPriceFloats(layers, firstLayerFloat, customPriceFloats)
	
	// 设置持仓方向
	positionSide := models.PositionSideLong
	if strategy.Side == models.OrderSideSell {
		positionSide = models.PositionSideShort
	}
	
	// 创建所有层的订单
	var orders []*models.FuturesOrder
	for i := 0; i < layers; i++ {
		// 计算每层的数量
		layerQuantity := totalQuantity * quantityRatios[i]
		
		// 计算每层的价格
		layerPrice := basePrice
		if priceFloats[i] > 0 {
			floatRate := priceFloats[i] / 10000.0
			if strategy.Side == models.OrderSideBuy {
				// 做多时，价格向下浮动
				layerPrice = basePrice * (1 - floatRate)
			} else {
				// 做空时，价格向上浮动
				layerPrice = basePrice * (1 + floatRate)
			}
		}
		
		order := &models.FuturesOrder{
			UserID:        strategy.UserID,
			StrategyID:    &strategy.ID,
			Symbol:        strategy.Symbol,
			Side:          strategy.Side,
			PositionSide:  positionSide,
			Type:          models.OrderTypeLimit,
			Quantity:      layerQuantity,
			Price:         layerPrice,
			ClientOrderID: fmt.Sprintf("%s_L%d", utils.GenerateUUID(), i+1),
			TimeInForce:   "GTC",
		}
		
		resp, err := binanceService.CreateFuturesOrder(context.Background(), order)
		if err != nil {
			log.Printf("创建第%d层订单失败: %v", i+1, err)
			continue
		}
		
		order.OrderID = strconv.FormatInt(resp.OrderID, 10)
		order.Status = models.OrderStatus(resp.Status)
		orders = append(orders, order)
		
		if err := fs.db.Create(order).Error; err != nil {
			log.Printf("保存第%d层期货订单失败: %v", i+1, err)
		}
	}
	
	// 设置止盈止损（仅第一次）
	if strategy.TakeProfitBP > 0 || strategy.StopLossBP > 0 {
		go fs.createStopLossAndTakeProfit(strategy, binanceService, basePrice, totalQuantity)
	}
	
	// 启动超时监控
	if timeoutMinutes > 0 && len(orders) > 0 {
		go fs.monitorIcebergTimeout(strategy, binanceService, orders, timeoutMinutes)
	}
	
	return nil
}

func (fs *FuturesService) executeSlowFuturesIcebergStrategy(strategy *models.FuturesStrategy, binanceService *BinanceService) error {
	config := strategy.Config
	if config == nil {
		config = make(map[string]interface{})
	}
	
	// 获取层数配置，默认10层
	layers := 10
	if configValue, exists := config["layers"]; exists && configValue != nil {
		if layersConfig, ok := configValue.(float64); ok {
			layers = int(layersConfig)
			if layers < 5 || layers > 10 {
				layers = 10
			}
		}
	}
	
	// 获取超时时间（分钟），默认5分钟
	timeoutMinutes := 5
	if configValue, exists := config["timeout_minutes"]; exists && configValue != nil {
		if timeoutConfig, ok := configValue.(float64); ok {
			timeoutMinutes = int(timeoutConfig)
		}
	}
	
	// 获取自定义的层级浮动比例
	var customPriceFloats []float64
	if configValue, exists := config["layer_price_floats"]; exists && configValue != nil {
		if floatsInterface, ok := configValue.([]interface{}); ok {
			customPriceFloats = make([]float64, 0, len(floatsInterface))
			for _, v := range floatsInterface {
				if f, ok := v.(float64); ok {
					customPriceFloats = append(customPriceFloats, f)
				}
			}
		}
	}
	
	// 获取首单浮动比例
	firstLayerFloat := strategy.FloatBasisPoints
	if configValue, exists := config["first_layer_float"]; exists && configValue != nil {
		if firstFloat, ok := configValue.(float64); ok {
			firstLayerFloat = firstFloat
		}
	}
	
	// 获取当前层级（从state中读取）
	var currentLayer int
	if strategy.State != nil {
		if layerValue, exists := strategy.State["current_layer"]; exists && layerValue != nil {
			if layer, ok := layerValue.(float64); ok {
				currentLayer = int(layer)
			}
		}
	}
	
	// 如果已完成所有层级
	if currentLayer >= layers {
		fs.db.Model(strategy).Update("is_completed", true)
		return nil
	}
	
	// 获取当前价格
	currentPrice, err := binanceService.GetFuturesPrice(context.Background(), strategy.Symbol)
	if err != nil {
		return err
	}
	
	// 检查是否触发
	shouldExecute := false
	if strategy.Price > 0 {
		if strategy.Side == models.OrderSideBuy && currentPrice <= strategy.Price {
			shouldExecute = true
		} else if strategy.Side == models.OrderSideSell && currentPrice >= strategy.Price {
			shouldExecute = true
		}
	} else {
		shouldExecute = true
	}
	
	if !shouldExecute {
		return nil
	}
	
	// 检查上一层订单是否已完成
	if currentLayer > 0 {
		var lastOrder models.FuturesOrder
		if err := fs.db.Where("strategy_id = ? AND client_order_id LIKE ?", strategy.ID, fmt.Sprintf("%%_L%d", currentLayer)).
			Order("created_at desc").First(&lastOrder).Error; err == nil {
			// 检查订单状态
			orderStatus, err := binanceService.GetFuturesOrderStatus(context.Background(), strategy.Symbol, lastOrder.OrderID)
			if err == nil {
				if orderStatus.Status == "NEW" || orderStatus.Status == "PARTIALLY_FILLED" {
					// 检查是否超时
					if time.Since(lastOrder.CreatedAt).Minutes() > float64(timeoutMinutes) {
						// 撤销订单
						_, err := binanceService.CancelFuturesOrder(context.Background(), strategy.Symbol, lastOrder.OrderID)
						if err != nil {
							log.Printf("撤销第%d层订单失败: %v", currentLayer, err)
						}
						fs.db.Model(&lastOrder).Update("status", models.OrderStatusCanceled)
					} else {
						// 订单还在执行中，等待
						return nil
					}
				}
			}
		}
	}
	
	// 获取订单簿深度，获取买1价或卖1价
	basePrice := currentPrice
	if depth, err := binanceService.GetFuturesDepth(context.Background(), strategy.Symbol, 5); err == nil {
		if strategy.Side == models.OrderSideBuy && len(depth.Bids) > 0 {
			bidPrice, _ := strconv.ParseFloat(depth.Bids[0].Price, 64)
			if bidPrice > 0 {
				basePrice = bidPrice
			}
		} else if strategy.Side == models.OrderSideSell && len(depth.Asks) > 0 {
			askPrice, _ := strconv.ParseFloat(depth.Asks[0].Price, 64)
			if askPrice > 0 {
				basePrice = askPrice
			}
		}
	}
	
	// 计算总数量
	totalOrderValue := strategy.MarginAmount * float64(strategy.Leverage)
	totalQuantity := totalOrderValue / basePrice
	
	// 获取层级配置
	quantityRatios := fs.getQuantityRatios(layers)
	priceFloats := fs.getPriceFloats(layers, firstLayerFloat, customPriceFloats)
	
	// 计算当前层的数量
	layerQuantity := totalQuantity * quantityRatios[currentLayer]
	
	// 计算当前层的价格
	layerPrice := basePrice
	if priceFloats[currentLayer] > 0 {
		floatRate := priceFloats[currentLayer] / 10000.0
		if strategy.Side == models.OrderSideBuy {
			layerPrice = basePrice * (1 - floatRate)
		} else {
			layerPrice = basePrice * (1 + floatRate)
		}
	}
	
	// 设置持仓方向
	positionSide := models.PositionSideLong
	if strategy.Side == models.OrderSideSell {
		positionSide = models.PositionSideShort
	}
	
	order := &models.FuturesOrder{
		UserID:        strategy.UserID,
		StrategyID:    &strategy.ID,
		Symbol:        strategy.Symbol,
		Side:          strategy.Side,
		PositionSide:  positionSide,
		Type:          models.OrderTypeLimit,
		Quantity:      layerQuantity,
		Price:         layerPrice,
		ClientOrderID: fmt.Sprintf("%s_L%d", utils.GenerateUUID(), currentLayer+1),
		TimeInForce:   "GTC",
	}
	
	resp, err := binanceService.CreateFuturesOrder(context.Background(), order)
	if err != nil {
		return fmt.Errorf("创建第%d层订单失败: %v", currentLayer+1, err)
	}
	
	order.OrderID = strconv.FormatInt(resp.OrderID, 10)
	order.Status = models.OrderStatus(resp.Status)
	
	if err := fs.db.Create(order).Error; err != nil {
		log.Printf("保存第%d层期货订单失败: %v", currentLayer+1, err)
	}
	
	// 更新当前层级
	if strategy.State == nil {
		strategy.State = make(models.StrategyState)
	}
	strategy.State["current_layer"] = currentLayer + 1
	fs.db.Model(strategy).Update("state", strategy.State)
	
	// 设置止盈止损（仅第一层）
	if currentLayer == 0 && (strategy.TakeProfitBP > 0 || strategy.StopLossBP > 0) {
		go fs.createStopLossAndTakeProfit(strategy, binanceService, basePrice, totalQuantity)
	}
	
	return nil
}

func (fs *FuturesService) createStopLossAndTakeProfit(strategy *models.FuturesStrategy, binanceService *BinanceService, entryPrice float64, quantity float64) {
	time.Sleep(5 * time.Second)

	// 根据做多/做空决定平仓方向
	closeSide := models.OrderSideSell
	positionSide := models.PositionSideLong
	if strategy.Side == models.OrderSideSell {
		closeSide = models.OrderSideBuy
		positionSide = models.PositionSideShort
	}

	// 止损订单
	if strategy.StopLossBP > 0 {
		stopLossRate := float64(strategy.StopLossBP) / 10000.0
		var stopLossPrice float64

		if strategy.Side == models.OrderSideBuy {
			// 做多止损：入场价格 * (1 - 止损万分比)
			stopLossPrice = entryPrice * (1 - stopLossRate)
		} else {
			// 做空止损：入场价格 * (1 + 止损万分比)
			stopLossPrice = entryPrice * (1 + stopLossRate)
		}

		stopLossOrder := &models.FuturesOrder{
			UserID:        strategy.UserID,
			StrategyID:    &strategy.ID,
			Symbol:        strategy.Symbol,
			Side:          closeSide,
			PositionSide:  positionSide,
			Type:          models.OrderTypeStopLoss,
			Quantity:      quantity,
			StopPrice:     stopLossPrice,
			ReduceOnly:    true,
			ClientOrderID: utils.GenerateUUID(),
		}

		resp, err := binanceService.CreateFuturesOrder(context.Background(), stopLossOrder)
		if err != nil {
			log.Printf("创建止损订单失败: %v", err)
		} else {
			stopLossOrder.OrderID = strconv.FormatInt(resp.OrderID, 10)
			stopLossOrder.Status = models.OrderStatus(resp.Status)
			fs.db.Create(stopLossOrder)
		}
	}

	// 止盈订单
	if strategy.TakeProfitBP > 0 {
		takeProfitRate := float64(strategy.TakeProfitBP) / 10000.0
		var takeProfitPrice float64

		if strategy.Side == models.OrderSideBuy {
			// 做多止盈：入场价格 * (1 + 止盈万分比)
			takeProfitPrice = entryPrice * (1 + takeProfitRate)
		} else {
			// 做空止盈：入场价格 * (1 - 止盈万分比)
			takeProfitPrice = entryPrice * (1 - takeProfitRate)
		}

		takeProfitOrder := &models.FuturesOrder{
			UserID:        strategy.UserID,
			StrategyID:    &strategy.ID,
			Symbol:        strategy.Symbol,
			Side:          closeSide,
			PositionSide:  positionSide,
			Type:          models.OrderTypeTakeProfit,
			Quantity:      quantity,
			StopPrice:     takeProfitPrice,
			ReduceOnly:    true,
			ClientOrderID: utils.GenerateUUID(),
		}

		resp, err := binanceService.CreateFuturesOrder(context.Background(), takeProfitOrder)
		if err != nil {
			log.Printf("创建止盈订单失败: %v", err)
		} else {
			takeProfitOrder.OrderID = strconv.FormatInt(resp.OrderID, 10)
			takeProfitOrder.Status = models.OrderStatus(resp.Status)
			fs.db.Create(takeProfitOrder)
		}
	}
}

func oppositeOrder(side models.OrderSide) models.OrderSide {
	if side == models.OrderSideBuy {
		return models.OrderSideSell
	}
	return models.OrderSideBuy
}

func (fs *FuturesService) UpdatePositions(userID uint) error {
	apiKey, secretKey, err := fs.userService.GetUserAPIKeys(userID)
	if err != nil {
		return err
	}

	binanceService, err := NewBinanceService(apiKey, secretKey)
	if err != nil {
		return err
	}
	positions, err := binanceService.GetFuturesPositions(context.Background())
	if err != nil {
		return err
	}

	for _, pos := range positions {
		positionAmt, _ := strconv.ParseFloat(pos.PositionAmt, 64)
		if positionAmt == 0 {
			continue
		}

		entryPrice, _ := strconv.ParseFloat(pos.EntryPrice, 64)
		markPrice, _ := strconv.ParseFloat(pos.MarkPrice, 64)
		unRealizedProfit, _ := strconv.ParseFloat(pos.UnRealizedProfit, 64)
		liquidationPrice, _ := strconv.ParseFloat(pos.LiquidationPrice, 64)
		leverage, _ := strconv.Atoi(pos.Leverage)
		maxNotionalValue, _ := strconv.ParseFloat(pos.MaxNotionalValue, 64)
		isolatedMargin, _ := strconv.ParseFloat(pos.IsolatedMargin, 64)

		position := &models.FuturesPosition{
			UserID:           userID,
			Symbol:           pos.Symbol,
			PositionSide:     models.PositionSide(pos.PositionSide),
			PositionAmt:      positionAmt,
			EntryPrice:       entryPrice,
			MarkPrice:        markPrice,
			UnRealizedProfit: unRealizedProfit,
			LiquidationPrice: liquidationPrice,
			Leverage:         leverage,
			MaxNotionalValue: maxNotionalValue,
			MarginType:       models.MarginType(pos.MarginType),
			IsolatedMargin:   isolatedMargin,
			IsAutoAddMargin:  pos.IsAutoAddMargin == "true",
		}

		var existingPosition models.FuturesPosition
		if err := fs.db.Where("user_id = ? AND symbol = ? AND position_side = ?",
			userID, pos.Symbol, pos.PositionSide).First(&existingPosition).Error; err != nil {
			fs.db.Create(position)
		} else {
			fs.db.Model(&existingPosition).Updates(position)
		}
	}

	return nil
}

func (fs *FuturesService) GetUserPositions(userID uint) ([]models.FuturesPosition, error) {
	var positions []models.FuturesPosition
	if err := fs.db.Where("user_id = ? AND position_amt != 0", userID).Find(&positions).Error; err != nil {
		return nil, err
	}
	return positions, nil
}

func (fs *FuturesService) GetFuturesStats(userID uint) (map[string]interface{}, error) {
	var totalStrategies int64
	var activeStrategies int64
	var totalOrders int64
	var totalProfit float64

	fs.db.Model(&models.FuturesStrategy{}).Where("user_id = ?", userID).Count(&totalStrategies)
	fs.db.Model(&models.FuturesStrategy{}).Where("user_id = ? AND is_active = ?", userID, true).Count(&activeStrategies)
	fs.db.Model(&models.FuturesOrder{}).Where("user_id = ?", userID).Count(&totalOrders)

	var positions []models.FuturesPosition
	fs.db.Where("user_id = ?", userID).Find(&positions)
	for _, position := range positions {
		totalProfit += position.UnRealizedProfit
	}

	stats := map[string]interface{}{
		"total_strategies":  totalStrategies,
		"active_strategies": activeStrategies,
		"total_orders":      totalOrders,
		"total_profit":      totalProfit,
	}

	return stats, nil
}

// getQuantityRatios 获取每层的数量占比
func (fs *FuturesService) getQuantityRatios(layers int) []float64 {
	switch layers {
	case 10:
		return []float64{0.19, 0.17, 0.15, 0.13, 0.11, 0.09, 0.07, 0.05, 0.03, 0.01}
	case 9:
		// 等差数列，第一层最多，第九层最少
		diff := 2.0 / float64(layers*(layers-1)/2)
		ratios := make([]float64, layers)
		for i := 0; i < layers; i++ {
			ratios[i] = (float64(layers-i) * diff)
		}
		return ratios
	case 8:
		diff := 2.0 / float64(layers*(layers-1)/2)
		ratios := make([]float64, layers)
		for i := 0; i < layers; i++ {
			ratios[i] = (float64(layers-i) * diff)
		}
		return ratios
	case 7:
		diff := 2.0 / float64(layers*(layers-1)/2)
		ratios := make([]float64, layers)
		for i := 0; i < layers; i++ {
			ratios[i] = (float64(layers-i) * diff)
		}
		return ratios
	case 6:
		diff := 2.0 / float64(layers*(layers-1)/2)
		ratios := make([]float64, layers)
		for i := 0; i < layers; i++ {
			ratios[i] = (float64(layers-i) * diff)
		}
		return ratios
	case 5:
		diff := 2.0 / float64(layers*(layers-1)/2)
		ratios := make([]float64, layers)
		for i := 0; i < layers; i++ {
			ratios[i] = (float64(layers-i) * diff)
		}
		return ratios
	default:
		return []float64{0.19, 0.17, 0.15, 0.13, 0.11, 0.09, 0.07, 0.05, 0.03, 0.01}
	}
}

// getPriceFloats 获取每层的价格浮动（万分比）
func (fs *FuturesService) getPriceFloats(layers int, firstLayerFloat float64, customFloats []float64) []float64 {
	floats := make([]float64, layers)
	
	// 如果有自定义浮动比例，使用自定义值
	if len(customFloats) > 0 {
		for i := 0; i < layers && i < len(customFloats); i++ {
			floats[i] = customFloats[i]
		}
		// 如果自定义数组长度不够，剩余的使用默认规则
		for i := len(customFloats); i < layers; i++ {
			floats[i] = float64(i * 8)
		}
	} else {
		// 使用默认规则：首层使用firstLayerFloat，后续每层增加万分之八
		floats[0] = firstLayerFloat
		for i := 1; i < layers; i++ {
			floats[i] = float64(i * 8)
		}
	}
	
	return floats
}

// monitorIcebergTimeout 监控冰山策略订单超时
func (fs *FuturesService) monitorIcebergTimeout(strategy *models.FuturesStrategy, binanceService *BinanceService, orders []*models.FuturesOrder, timeoutMinutes int) {
	// 等待超时时间
	time.Sleep(time.Duration(timeoutMinutes) * time.Minute)
	
	ctx := context.Background()
	allCompleted := true
	
	for _, order := range orders {
		// 检查订单状态
		orderStatus, err := binanceService.GetFuturesOrderStatus(ctx, strategy.Symbol, order.OrderID)
		if err != nil {
			log.Printf("获取订单%s状态失败: %v", order.OrderID, err)
			continue
		}
		
		if orderStatus.Status == "NEW" || orderStatus.Status == "PARTIALLY_FILLED" {
			allCompleted = false
			// 撤销未完成的订单
			_, err := binanceService.CancelFuturesOrder(ctx, strategy.Symbol, order.OrderID)
			if err != nil {
				log.Printf("撤销订单%s失败: %v", order.OrderID, err)
			} else {
				fs.db.Model(order).Update("status", models.OrderStatusCanceled)
			}
		}
	}
	
	// 如果所有订单都完成了，标记策略为完成
	if allCompleted {
		fs.db.Model(strategy).Update("is_completed", true)
	}
}
