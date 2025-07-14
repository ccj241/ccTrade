package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"strconv"
	"time"

	"github.com/ccj241/cctrade/config"
	"github.com/ccj241/cctrade/models"
	"github.com/ccj241/cctrade/utils"
	"gorm.io/gorm"
)

type StrategyService struct {
	db          *gorm.DB
	userService *UserService
}

func NewStrategyService() *StrategyService {
	return &StrategyService{
		db:          config.DB,
		userService: NewUserService(),
	}
}

func (ss *StrategyService) CreateStrategy(userID uint, strategyData map[string]interface{}) (*models.Strategy, error) {
	strategy := &models.Strategy{
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

	if quantity, ok := strategyData["quantity"].(float64); ok {
		if err := utils.ValidateQuantity(quantity); err != nil {
			return nil, err
		}
		strategy.Quantity = quantity
	}

	if price, ok := strategyData["price"].(float64); ok && price > 0 {
		strategy.Price = price
	}

	if triggerPrice, ok := strategyData["trigger_price"].(float64); ok && triggerPrice > 0 {
		strategy.TriggerPrice = triggerPrice
	}

	if stopPrice, ok := strategyData["stop_price"].(float64); ok && stopPrice > 0 {
		strategy.StopPrice = stopPrice
	}

	if takeProfit, ok := strategyData["take_profit"].(float64); ok && takeProfit > 0 {
		strategy.TakeProfit = takeProfit
	}

	if stopLoss, ok := strategyData["stop_loss"].(float64); ok && stopLoss > 0 {
		strategy.StopLoss = stopLoss
	}

	if autoRestart, ok := strategyData["auto_restart"].(bool); ok {
		strategy.AutoRestart = autoRestart
	}

	if config, ok := strategyData["config"].(map[string]interface{}); ok {
		strategy.Config = models.StrategyConfig(config)
	}

	if err := ss.validateStrategyConfig(strategy); err != nil {
		return nil, err
	}

	if err := ss.db.Create(strategy).Error; err != nil {
		return nil, err
	}

	return strategy, nil
}

func (ss *StrategyService) validateStrategyConfig(strategy *models.Strategy) error {
	switch strategy.Type {
	case models.StrategySimple:
		return ss.validateSimpleStrategy(strategy)
	case models.StrategyGrid:
		return ss.validateGridStrategy(strategy)
	case models.StrategyDCA:
		return ss.validateDCAStrategy(strategy)
	case models.StrategyIceberg, models.StrategySlowIceberg:
		return ss.validateIcebergStrategy(strategy)
	default:
		return nil
	}
}

func (ss *StrategyService) validateSimpleStrategy(strategy *models.Strategy) error {
	// 简单策略必须有触发价格
	if strategy.TriggerPrice <= 0 {
		return errors.New("简单策略需要设置触发价格")
	}

	config := strategy.Config
	if config == nil {
		config = make(models.StrategyConfig)
		strategy.Config = config
	}

	// 验证价格浮动
	priceFloat, ok := config["price_float"].(float64)
	if !ok {
		config["price_float"] = 0.0 // 默认不浮动
	} else if priceFloat < 0 || priceFloat > 10000 {
		return errors.New("价格浮动必须在0-10000万分比之间")
	}

	// 验证超时时间
	timeout, ok := config["timeout"].(float64)
	if !ok || timeout <= 0 {
		config["timeout"] = 5.0 // 默认5分钟
	}

	return nil
}

func (ss *StrategyService) validateGridStrategy(strategy *models.Strategy) error {
	config := strategy.Config

	upperPrice, ok := config["upper_price"].(float64)
	if !ok || upperPrice <= 0 {
		return errors.New("网格策略需要设置上限价格")
	}

	lowerPrice, ok := config["lower_price"].(float64)
	if !ok || lowerPrice <= 0 {
		return errors.New("网格策略需要设置下限价格")
	}

	if upperPrice <= lowerPrice {
		return errors.New("上限价格必须大于下限价格")
	}

	gridCount, ok := config["grid_count"].(float64)
	if !ok || gridCount < 2 {
		return errors.New("网格数量至少为2")
	}

	return nil
}

func (ss *StrategyService) validateDCAStrategy(strategy *models.Strategy) error {
	config := strategy.Config

	interval, ok := config["interval"].(float64)
	if !ok || interval <= 0 {
		return errors.New("DCA策略需要设置投资间隔")
	}

	totalAmount, ok := config["total_amount"].(float64)
	if !ok || totalAmount <= 0 {
		return errors.New("DCA策略需要设置总投资金额")
	}

	return nil
}

func (ss *StrategyService) validateIcebergStrategy(strategy *models.Strategy) error {
	// 冰山策略必须有触发价格
	if strategy.TriggerPrice <= 0 {
		return errors.New("冰山策略需要设置触发价格")
	}

	config := strategy.Config
	if config == nil {
		config = make(models.StrategyConfig)
		strategy.Config = config
	}

	// 验证层数
	layers, ok := config["layers"].(float64)
	if !ok || layers < 5 || layers > 10 {
		return errors.New("冰山策略的层数必须在5-10之间")
	}

	// 验证超时时间
	timeout, ok := config["timeout"].(float64)
	if !ok || timeout <= 0 {
		config["timeout"] = 5.0 // 默认5分钟
	}

	// 验证价格浮动步长
	_, ok = config["price_float_step"].(float64)
	if !ok {
		config["price_float_step"] = 8.0 // 默认万分之8
	}

	// 验证层级数量分布
	layerQuantities, ok := config["layer_quantities"].([]interface{})
	if !ok || len(layerQuantities) != int(layers) {
		// 自动生成默认分布
		quantities := ss.calculateDefaultLayerQuantities(int(layers))
		config["layer_quantities"] = quantities
	}

	// 验证层级价格浮动
	layerPriceFloats, ok := config["layer_price_floats"].([]interface{})
	if !ok || len(layerPriceFloats) != int(layers) {
		// 自动生成默认价格浮动
		priceFloats := ss.calculateDefaultLayerPriceFloats(int(layers))
		config["layer_price_floats"] = priceFloats
	}

	return nil
}

func (ss *StrategyService) GetUserStrategies(userID uint, page, limit int) ([]models.Strategy, int64, error) {
	var strategies []models.Strategy
	var total int64

	query := ss.db.Model(&models.Strategy{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Find(&strategies).Error; err != nil {
		return nil, 0, err
	}

	return strategies, total, nil
}

func (ss *StrategyService) GetStrategyByID(userID, strategyID uint) (*models.Strategy, error) {
	var strategy models.Strategy
	if err := ss.db.Where("id = ? AND user_id = ?", strategyID, userID).First(&strategy).Error; err != nil {
		return nil, err
	}
	return &strategy, nil
}

func (ss *StrategyService) UpdateStrategy(userID, strategyID uint, updates map[string]interface{}) error {
	allowedFields := []string{"name", "is_active", "auto_restart", "take_profit", "stop_loss", "config"}
	filteredUpdates := make(map[string]interface{})

	for field, value := range updates {
		if utils.Contains(allowedFields, field) {
			filteredUpdates[field] = value
		}
	}

	if len(filteredUpdates) == 0 {
		return errors.New("没有有效的更新字段")
	}

	return ss.db.Model(&models.Strategy{}).Where("id = ? AND user_id = ?", strategyID, userID).Updates(filteredUpdates).Error
}

func (ss *StrategyService) ToggleStrategy(userID, strategyID uint) error {
	var strategy models.Strategy
	if err := ss.db.Where("id = ? AND user_id = ?", strategyID, userID).First(&strategy).Error; err != nil {
		return err
	}

	return ss.db.Model(&strategy).Update("is_active", !strategy.IsActive).Error
}

func (ss *StrategyService) DeleteStrategy(userID, strategyID uint) error {
	return ss.db.Where("id = ? AND user_id = ?", strategyID, userID).Delete(&models.Strategy{}).Error
}

func (ss *StrategyService) ExecuteStrategy(strategy *models.Strategy) error {
	apiKey, secretKey, err := ss.userService.GetUserAPIKeys(strategy.UserID)
	if err != nil {
		return err
	}

	binanceService, err := NewBinanceService(apiKey, secretKey)
	if err != nil {
		return err
	}

	switch strategy.Type {
	case models.StrategySimple:
		return ss.executeSimpleStrategy(strategy, binanceService)
	case models.StrategyIceberg:
		return ss.executeIcebergStrategy(strategy, binanceService)
	case models.StrategySlowIceberg:
		return ss.executeSlowIcebergStrategy(strategy, binanceService)
	case models.StrategyGrid:
		return ss.executeGridStrategy(strategy, binanceService)
	case models.StrategyDCA:
		return ss.executeDCAStrategy(strategy, binanceService)
	default:
		return fmt.Errorf("不支持的策略类型: %s", strategy.Type)
	}
}

func (ss *StrategyService) executeSimpleStrategy(strategy *models.Strategy, binanceService *BinanceService) error {
	currentPrice, err := binanceService.GetPrice(context.Background(), strategy.Symbol)
	if err != nil {
		return err
	}

	shouldExecute := false
	if strategy.TriggerPrice > 0 {
		if strategy.Side == models.OrderSideBuy && currentPrice <= strategy.TriggerPrice {
			shouldExecute = true
		} else if strategy.Side == models.OrderSideSell && currentPrice >= strategy.TriggerPrice {
			shouldExecute = true
		}
	} else {
		shouldExecute = true
	}

	if !shouldExecute {
		return nil
	}

	order := &models.Order{
		UserID:        strategy.UserID,
		StrategyID:    &strategy.ID,
		Symbol:        strategy.Symbol,
		Side:          strategy.Side,
		Type:          models.OrderTypeMarket,
		Quantity:      strategy.Quantity,
		Price:         strategy.Price,
		ClientOrderID: utils.GenerateUUID(),
	}

	if strategy.Price > 0 {
		order.Type = models.OrderTypeLimit
		order.TimeInForce = "GTC"
	}

	resp, err := binanceService.CreateSpotOrder(context.Background(), order)
	if err != nil {
		return err
	}

	order.OrderID = strconv.FormatInt(resp.OrderID, 10)
	order.Status = models.OrderStatus(resp.Status)

	if err := ss.db.Create(order).Error; err != nil {
		log.Printf("保存订单失败: %v", err)
	}

	ss.db.Model(strategy).Update("is_completed", true)

	return nil
}

func (ss *StrategyService) executeIcebergStrategy(strategy *models.Strategy, binanceService *BinanceService) error {
	config := strategy.Config
	layers := int(config["layers"].(float64))
	layerQuantities := config["layer_quantities"].([]interface{})
	layerPriceFloats := config["layer_price_floats"].([]interface{})
	timeout := int(config["timeout"].(float64))

	// 获取当前价格
	currentPrice, err := binanceService.GetPrice(context.Background(), strategy.Symbol)
	if err != nil {
		return err
	}

	// 检查是否触发
	if strategy.Side == models.OrderSideBuy && currentPrice > strategy.TriggerPrice {
		return nil // 买入时当前价格高于触发价，不执行
	}
	if strategy.Side == models.OrderSideSell && currentPrice < strategy.TriggerPrice {
		return nil // 卖出时当前价格低于触发价，不执行
	}

	// 检查现有订单
	var existingOrders []models.Order
	ss.db.Where("strategy_id = ? AND status IN ?", strategy.ID, []string{"NEW", "PARTIALLY_FILLED"}).Find(&existingOrders)

	// 如果有活跃订单，检查是否超时
	if len(existingOrders) > 0 {
		oldestOrder := existingOrders[0]
		if time.Since(oldestOrder.CreatedAt).Minutes() > float64(timeout) {
			// 取消所有未完成订单
			for _, order := range existingOrders {
				if _, err := binanceService.CancelSpotOrder(context.Background(), order.Symbol, order.OrderID); err != nil {
					log.Printf("取消订单失败: %v", err)
				}
				ss.db.Model(&order).Update("status", "CANCELED")
			}
		} else {
			return nil // 还有活跃订单且未超时
		}
	}

	// 获取买一/卖一价
	depth, err := binanceService.GetOrderBook(context.Background(), strategy.Symbol, 5)
	if err != nil {
		return err
	}

	var basePrice float64
	if strategy.Side == models.OrderSideBuy {
		if len(depth.Bids) > 0 {
			basePrice, _ = strconv.ParseFloat(depth.Bids[0].Price, 64)
		} else {
			basePrice = currentPrice * 0.999 // 没有买一价时使用当前价格-0.1%
		}
	} else {
		if len(depth.Asks) > 0 {
			basePrice, _ = strconv.ParseFloat(depth.Asks[0].Price, 64)
		} else {
			basePrice = currentPrice * 1.001 // 没有卖一价时使用当前价格+0.1%
		}
	}

	// 创建所有层的订单
	for i := 0; i < layers; i++ {
		quantityRatio := layerQuantities[i].(float64)
		priceFloat := layerPriceFloats[i].(float64)
		
		// 计算该层数量
		layerQty := strategy.Quantity * quantityRatio
		
		// 计算该层价格
		var layerPrice float64
		if strategy.Side == models.OrderSideBuy {
			// 买入时向下浮动
			layerPrice = basePrice * (1 - priceFloat/10000)
		} else {
			// 卖出时向上浮动
			layerPrice = basePrice * (1 + priceFloat/10000)
		}

		order := &models.Order{
			UserID:        strategy.UserID,
			StrategyID:    &strategy.ID,
			Symbol:        strategy.Symbol,
			Side:          strategy.Side,
			Type:          models.OrderTypeLimit,
			Quantity:      layerQty,
			Price:         layerPrice,
			TimeInForce:   "GTC",
			ClientOrderID: utils.GenerateUUID(),
		}

		resp, err := binanceService.CreateSpotOrder(context.Background(), order)
		if err != nil {
			log.Printf("创建第%d层订单失败: %v", i+1, err)
			continue
		}

		order.OrderID = strconv.FormatInt(resp.OrderID, 10)
		order.Status = models.OrderStatus(resp.Status)

		if err := ss.db.Create(order).Error; err != nil {
			log.Printf("保存订单失败: %v", err)
		}
	}

	return nil
}

func (ss *StrategyService) executeSlowIcebergStrategy(strategy *models.Strategy, binanceService *BinanceService) error {
	config := strategy.Config
	layers := int(config["layers"].(float64))
	layerQuantities := config["layer_quantities"].([]interface{})
	layerPriceFloats := config["layer_price_floats"].([]interface{})
	timeout := int(config["timeout"].(float64))
	
	// 获取当前价格
	currentPrice, err := binanceService.GetPrice(context.Background(), strategy.Symbol)
	if err != nil {
		return err
	}
	
	// 检查是否触发
	if strategy.Side == models.OrderSideBuy && currentPrice > strategy.TriggerPrice {
		return nil
	}
	if strategy.Side == models.OrderSideSell && currentPrice < strategy.TriggerPrice {
		return nil
	}
	
	// 获取策略的执行状态
	var strategyState map[string]interface{}
	if strategy.State != nil {
		strategyState = strategy.State
	} else {
		strategyState = make(map[string]interface{})
	}
	
	// 初始化状态
	currentLayer, _ := strategyState["current_layer"].(float64)
	currentLayerInt := int(currentLayer)
	
	layerFilledQuantity, _ := strategyState["layer_filled_quantity"].(float64)
	totalFilledQuantity, _ := strategyState["total_filled_quantity"].(float64)
	
	// 检查是否所有层都已完成
	if currentLayerInt >= layers {
		ss.db.Model(strategy).Update("is_completed", true)
		return nil
	}
	
	// 检查现有活跃订单
	var activeOrders []models.Order
	ss.db.Where("strategy_id = ? AND status IN ?", strategy.ID, []string{"NEW", "PARTIALLY_FILLED"}).Find(&activeOrders)
	
	if len(activeOrders) > 0 {
		// 检查超时
		for _, order := range activeOrders {
			if time.Since(order.CreatedAt).Minutes() > float64(timeout) {
				// 获取订单最新状态
				orderResp, err := binanceService.GetSpotOrderStatus(context.Background(), order.Symbol, order.OrderID)
				if err == nil {
					// 更新订单状态
					executedQty, _ := strconv.ParseFloat(orderResp.ExecutedQuantity, 64)
					if executedQty > 0 {
						// 部分成交，更新已成交数量
						layerFilledQuantity += executedQty
						totalFilledQuantity += executedQty
						
						// 更新数据库中的订单信息
						ss.db.Model(&order).Updates(map[string]interface{}{
							"executed_qty": executedQty,
							"status": orderResp.Status,
						})
					}
				}
				
				// 取消超时订单
				if _, err := binanceService.CancelSpotOrder(context.Background(), order.Symbol, order.OrderID); err != nil {
					log.Printf("取消订单失败: %v", err)
				} else {
					ss.db.Model(&order).Update("status", "CANCELED")
				}
			}
		}
		
		// 重新检查活跃订单
		ss.db.Where("strategy_id = ? AND status IN ?", strategy.ID, []string{"NEW", "PARTIALLY_FILLED"}).Find(&activeOrders)
		if len(activeOrders) > 0 {
			return nil // 还有活跃订单未超时
		}
	}
	
	// 计算当前层的总数量
	var currentLayerTotalQty float64
	if currentLayerInt < layers {
		quantityRatio := layerQuantities[currentLayerInt].(float64)
		currentLayerTotalQty = strategy.Quantity * quantityRatio
	}
	
	// 检查当前层是否已完成
	if layerFilledQuantity >= currentLayerTotalQty - 0.00000001 { // 使用极小值避免浮点数精度问题
		// 当前层已完成，进入下一层
		currentLayerInt++
		layerFilledQuantity = 0
		
		// 检查是否所有层都已完成
		if currentLayerInt >= layers {
			strategyState["current_layer"] = float64(currentLayerInt)
			strategyState["layer_filled_quantity"] = 0
			strategyState["total_filled_quantity"] = totalFilledQuantity
			ss.db.Model(strategy).Update("state", strategyState)
			ss.db.Model(strategy).Update("is_completed", true)
			return nil
		}
		
		// 更新当前层的总数量
		quantityRatio := layerQuantities[currentLayerInt].(float64)
		currentLayerTotalQty = strategy.Quantity * quantityRatio
	}
	
	// 计算当前层剩余需要挂单的数量
	remainingQty := currentLayerTotalQty - layerFilledQuantity
	if remainingQty <= 0.00000001 { // 避免极小数量
		// 进入下一层
		currentLayerInt++
		layerFilledQuantity = 0
		if currentLayerInt >= layers {
			strategyState["current_layer"] = float64(currentLayerInt)
			strategyState["layer_filled_quantity"] = 0
			strategyState["total_filled_quantity"] = totalFilledQuantity
			ss.db.Model(strategy).Update("state", strategyState)
			ss.db.Model(strategy).Update("is_completed", true)
			return nil
		}
		quantityRatio := layerQuantities[currentLayerInt].(float64)
		remainingQty = strategy.Quantity * quantityRatio
	}
	
	// 获取最新的买一/卖一价（每次挂单都重新获取，确保价格最新）
	depth, err := binanceService.GetOrderBook(context.Background(), strategy.Symbol, 5)
	if err != nil {
		return err
	}
	
	var basePrice float64
	if strategy.Side == models.OrderSideBuy {
		if len(depth.Bids) > 0 {
			basePrice, _ = strconv.ParseFloat(depth.Bids[0].Price, 64)
		} else {
			basePrice = currentPrice * 0.999
		}
	} else {
		if len(depth.Asks) > 0 {
			basePrice, _ = strconv.ParseFloat(depth.Asks[0].Price, 64)
		} else {
			basePrice = currentPrice * 1.001
		}
	}
	
	// 根据最新买卖1价和当前层的价格浮动百分比计算价格
	priceFloat := layerPriceFloats[currentLayerInt].(float64)
	var layerPrice float64
	if strategy.Side == models.OrderSideBuy {
		layerPrice = basePrice * (1 - priceFloat/10000)
	} else {
		layerPrice = basePrice * (1 + priceFloat/10000)
	}
	
	// 创建订单（挂当前层的剩余数量）
	order := &models.Order{
		UserID:        strategy.UserID,
		StrategyID:    &strategy.ID,
		Symbol:        strategy.Symbol,
		Side:          strategy.Side,
		Type:          models.OrderTypeLimit,
		Quantity:      remainingQty,
		Price:         layerPrice,
		TimeInForce:   "GTC",
		ClientOrderID: utils.GenerateUUID(),
	}
	
	log.Printf("慢冰山策略: 第%d层，挂单数量: %.8f，基准价格(买/卖1价): %.8f，浮动万分之%.0f，最终价格: %.8f", 
		currentLayerInt+1, remainingQty, basePrice, priceFloat, layerPrice)
	
	resp, err := binanceService.CreateSpotOrder(context.Background(), order)
	if err != nil {
		log.Printf("创建订单失败: %v", err)
		return err
	}
	
	order.OrderID = strconv.FormatInt(resp.OrderID, 10)
	order.Status = models.OrderStatus(resp.Status)
	
	if err := ss.db.Create(order).Error; err != nil {
		log.Printf("保存订单失败: %v", err)
	}
	
	// 更新策略状态
	strategyState["current_layer"] = float64(currentLayerInt)
	strategyState["layer_filled_quantity"] = layerFilledQuantity
	strategyState["total_filled_quantity"] = totalFilledQuantity
	ss.db.Model(strategy).Update("state", strategyState)
	
	return nil
}

func (ss *StrategyService) executeGridStrategy(strategy *models.Strategy, binanceService *BinanceService) error {
	config := strategy.Config
	upperPrice := config["upper_price"].(float64)
	lowerPrice := config["lower_price"].(float64)
	gridCount := int(config["grid_count"].(float64))

	priceGap := (upperPrice - lowerPrice) / float64(gridCount-1)

	currentPrice, err := binanceService.GetPrice(context.Background(), strategy.Symbol)
	if err != nil {
		return err
	}

	if currentPrice < lowerPrice || currentPrice > upperPrice {
		return nil
	}

	var existingOrders []models.Order
	ss.db.Where("strategy_id = ? AND status = ?", strategy.ID, "NEW").Find(&existingOrders)

	if len(existingOrders) >= gridCount {
		return nil
	}

	for i := 0; i < gridCount; i++ {
		price := lowerPrice + float64(i)*priceGap

		exists := false
		for _, order := range existingOrders {
			if math.Abs(order.Price-price) < 0.001 {
				exists = true
				break
			}
		}

		if exists {
			continue
		}

		side := models.OrderSideBuy
		if price > currentPrice {
			side = models.OrderSideSell
		}

		order := &models.Order{
			UserID:        strategy.UserID,
			StrategyID:    &strategy.ID,
			Symbol:        strategy.Symbol,
			Side:          side,
			Type:          models.OrderTypeLimit,
			Quantity:      strategy.Quantity / float64(gridCount),
			Price:         price,
			TimeInForce:   "GTC",
			ClientOrderID: utils.GenerateUUID(),
		}

		resp, err := binanceService.CreateSpotOrder(context.Background(), order)
		if err != nil {
			log.Printf("创建网格订单失败: %v", err)
			continue
		}

		order.OrderID = strconv.FormatInt(resp.OrderID, 10)
		order.Status = models.OrderStatus(resp.Status)

		if err := ss.db.Create(order).Error; err != nil {
			log.Printf("保存订单失败: %v", err)
		}
	}

	return nil
}

func (ss *StrategyService) executeDCAStrategy(strategy *models.Strategy, binanceService *BinanceService) error {
	config := strategy.Config
	interval := config["interval"].(float64)
	totalAmount := config["total_amount"].(float64)

	var lastOrder models.Order
	if err := ss.db.Where("strategy_id = ?", strategy.ID).Order("created_at desc").First(&lastOrder).Error; err == nil {
		if time.Since(lastOrder.CreatedAt).Hours() < interval {
			return nil
		}
	}

	var totalInvested float64
	var orders []models.Order
	ss.db.Where("strategy_id = ? AND status = ?", strategy.ID, "FILLED").Find(&orders)

	for _, order := range orders {
		totalInvested += order.CumulativeQuoteQty
	}

	if totalInvested >= totalAmount {
		ss.db.Model(strategy).Update("is_completed", true)
		return nil
	}

	order := &models.Order{
		UserID:        strategy.UserID,
		StrategyID:    &strategy.ID,
		Symbol:        strategy.Symbol,
		Side:          models.OrderSideBuy,
		Type:          models.OrderTypeMarket,
		Quantity:      strategy.Quantity,
		ClientOrderID: utils.GenerateUUID(),
	}

	resp, err := binanceService.CreateSpotOrder(context.Background(), order)
	if err != nil {
		return err
	}

	order.OrderID = strconv.FormatInt(resp.OrderID, 10)
	order.Status = models.OrderStatus(resp.Status)

	if err := ss.db.Create(order).Error; err != nil {
		log.Printf("保存订单失败: %v", err)
	}

	return nil
}

func (ss *StrategyService) GetStrategyStats(userID, strategyID uint) (map[string]interface{}, error) {
	var strategy models.Strategy
	if err := ss.db.Where("id = ? AND user_id = ?", strategyID, userID).First(&strategy).Error; err != nil {
		return nil, err
	}

	var totalOrders int64
	var filledOrders int64
	var totalVolume float64
	var totalProfit float64

	ss.db.Model(&models.Order{}).Where("strategy_id = ?", strategyID).Count(&totalOrders)
	ss.db.Model(&models.Order{}).Where("strategy_id = ? AND status = ?", strategyID, "FILLED").Count(&filledOrders)

	var orders []models.Order
	ss.db.Where("strategy_id = ? AND status = ?", strategyID, "FILLED").Find(&orders)

	for _, order := range orders {
		totalVolume += order.CumulativeQuoteQty
	}

	stats := map[string]interface{}{
		"total_orders":  totalOrders,
		"filled_orders": filledOrders,
		"total_volume":  totalVolume,
		"total_profit":  totalProfit,
		"is_active":     strategy.IsActive,
		"is_completed":  strategy.IsCompleted,
	}

	return stats, nil
}

// calculateDefaultLayerQuantities 计算默认的层级数量分布
func (ss *StrategyService) calculateDefaultLayerQuantities(layers int) []float64 {
	quantities := make([]float64, layers)
	
	if layers == 10 {
		// 10层的特殊分布
		distribution := []float64{0.19, 0.17, 0.15, 0.13, 0.11, 0.09, 0.07, 0.05, 0.03, 0.01}
		copy(quantities, distribution)
	} else {
		// 其他层数使用等差数列
		totalSteps := float64(layers * (layers + 1) / 2)
		for i := 0; i < layers; i++ {
			quantities[i] = float64(layers-i) / totalSteps
		}
	}
	
	return quantities
}

// calculateDefaultLayerPriceFloats 计算默认的层级价格浮动
func (ss *StrategyService) calculateDefaultLayerPriceFloats(layers int) []float64 {
	priceFloats := make([]float64, layers)
	
	// 每层递增万分之8
	for i := 0; i < layers; i++ {
		priceFloats[i] = float64(i * 8)
	}
	
	return priceFloats
}

// syncOrderStatus 同步订单状态并更新策略状态
func (ss *StrategyService) syncOrderStatus(strategy *models.Strategy, order *models.Order, binanceService *BinanceService) error {
	// 获取订单最新状态
	orderResp, err := binanceService.GetSpotOrderStatus(context.Background(), order.Symbol, order.OrderID)
	if err != nil {
		return err
	}
	
	// 更新订单状态
	executedQty, _ := strconv.ParseFloat(orderResp.ExecutedQuantity, 64)
	
	updates := map[string]interface{}{
		"status": orderResp.Status,
		"executed_qty": executedQty,
	}
	
	if err := ss.db.Model(order).Updates(updates).Error; err != nil {
		return err
	}
	
	// 如果是慢冰山策略，更新策略状态
	if strategy.Type == models.StrategySlowIceberg && orderResp.Status == "FILLED" {
		var strategyState map[string]interface{}
		if strategy.State != nil {
			strategyState = strategy.State
		} else {
			strategyState = make(map[string]interface{})
		}
		
		// 更新已成交数量
		totalFilledQuantity, _ := strategyState["total_filled_quantity"].(float64)
		totalFilledQuantity += executedQty
		strategyState["total_filled_quantity"] = totalFilledQuantity
		
		// 当前层已成交数量清零（因为该订单已完全成交）
		strategyState["layer_filled_quantity"] = 0.0
		
		// 更新策略状态
		if err := ss.db.Model(strategy).Update("state", strategyState).Error; err != nil {
			return err
		}
	}
	
	return nil
}
