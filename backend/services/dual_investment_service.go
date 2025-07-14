package services

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/ccj241/cctrade/config"
	"github.com/ccj241/cctrade/models"
	"github.com/ccj241/cctrade/utils"
	"gorm.io/gorm"
)

type DualInvestmentService struct {
	db          *gorm.DB
	userService *UserService
}

func NewDualInvestmentService() *DualInvestmentService {
	return &DualInvestmentService{
		db:          config.DB,
		userService: NewUserService(),
	}
}

func (dis *DualInvestmentService) GetDualInvestmentProducts() ([]models.DualInvestmentProduct, error) {
	var products []models.DualInvestmentProduct
	if err := dis.db.Where("is_active = ?", true).Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (dis *DualInvestmentService) CreateDualInvestmentStrategy(userID uint, strategyData map[string]interface{}) (*models.DualInvestmentStrategy, error) {
	strategy := &models.DualInvestmentStrategy{
		UserID: userID,
	}

	if name, ok := strategyData["name"].(string); ok {
		strategy.Name = name
	} else {
		return nil, errors.New("策略名称不能为空")
	}

	if productID, ok := strategyData["product_id"].(string); ok {
		var product models.DualInvestmentProduct
		if err := dis.db.Where("product_id = ? AND is_active = ?", productID, true).First(&product).Error; err != nil {
			return nil, errors.New("产品不存在或已下架")
		}
		strategy.ProductID = productID
		strategy.BaseAsset = product.BaseAsset
		strategy.QuoteAsset = product.QuoteAsset
	} else {
		return nil, errors.New("产品ID不能为空")
	}

	if investmentType, ok := strategyData["investment_type"].(string); ok {
		strategy.InvestmentType = investmentType
	} else {
		return nil, errors.New("投资类型不能为空")
	}

	if amount, ok := strategyData["amount"].(float64); ok {
		if amount <= 0 {
			return nil, errors.New("投资金额必须大于0")
		}
		strategy.Amount = amount
	} else {
		return nil, errors.New("投资金额不能为空")
	}

	if triggerPrice, ok := strategyData["trigger_price"].(float64); ok && triggerPrice > 0 {
		strategy.TriggerPrice = triggerPrice
	}

	if minYieldRate, ok := strategyData["min_yield_rate"].(float64); ok && minYieldRate >= 0 {
		strategy.MinYieldRate = minYieldRate
	}

	if autoReinvest, ok := strategyData["auto_reinvest"].(bool); ok {
		strategy.AutoReinvest = autoReinvest
	}

	if ladderSteps, ok := strategyData["ladder_steps"].(float64); ok && ladderSteps > 0 {
		strategy.LadderSteps = int(ladderSteps)
		if strategy.InvestmentType == "ladder" {
			strategy.AmountPerStep = strategy.Amount / ladderSteps
		}
	} else if strategy.InvestmentType == "ladder" {
		return nil, errors.New("梯度投资需要设置步数")
	}

	if err := dis.db.Create(strategy).Error; err != nil {
		return nil, err
	}

	return strategy, nil
}

func (dis *DualInvestmentService) GetUserDualInvestmentStrategies(userID uint, page, limit int) ([]models.DualInvestmentStrategy, int64, error) {
	var strategies []models.DualInvestmentStrategy
	var total int64

	query := dis.db.Model(&models.DualInvestmentStrategy{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Find(&strategies).Error; err != nil {
		return nil, 0, err
	}

	return strategies, total, nil
}

func (dis *DualInvestmentService) GetDualInvestmentStrategyByID(userID, strategyID uint) (*models.DualInvestmentStrategy, error) {
	var strategy models.DualInvestmentStrategy
	if err := dis.db.Where("id = ? AND user_id = ?", strategyID, userID).First(&strategy).Error; err != nil {
		return nil, err
	}
	return &strategy, nil
}

func (dis *DualInvestmentService) UpdateDualInvestmentStrategy(userID, strategyID uint, updates map[string]interface{}) error {
	allowedFields := []string{"name", "is_active", "auto_reinvest", "min_yield_rate", "trigger_price"}
	filteredUpdates := make(map[string]interface{})

	for field, value := range updates {
		if utils.Contains(allowedFields, field) {
			filteredUpdates[field] = value
		}
	}

	if len(filteredUpdates) == 0 {
		return errors.New("没有有效的更新字段")
	}

	return dis.db.Model(&models.DualInvestmentStrategy{}).Where("id = ? AND user_id = ?", strategyID, userID).Updates(filteredUpdates).Error
}

func (dis *DualInvestmentService) ToggleDualInvestmentStrategy(userID, strategyID uint) error {
	var strategy models.DualInvestmentStrategy
	if err := dis.db.Where("id = ? AND user_id = ?", strategyID, userID).First(&strategy).Error; err != nil {
		return err
	}

	return dis.db.Model(&strategy).Update("is_active", !strategy.IsActive).Error
}

func (dis *DualInvestmentService) DeleteDualInvestmentStrategy(userID, strategyID uint) error {
	return dis.db.Where("id = ? AND user_id = ?", strategyID, userID).Delete(&models.DualInvestmentStrategy{}).Error
}

func (dis *DualInvestmentService) ExecuteDualInvestmentStrategy(strategy *models.DualInvestmentStrategy) error {
	apiKey, secretKey, err := dis.userService.GetUserAPIKeys(strategy.UserID)
	if err != nil {
		return err
	}

	binanceService, err := NewBinanceService(apiKey, secretKey)
	if err != nil {
		return err
	}

	switch strategy.InvestmentType {
	case "single":
		return dis.executeSingleInvestment(strategy, binanceService)
	case "auto_reinvest":
		return dis.executeAutoReinvestment(strategy, binanceService)
	case "ladder":
		return dis.executeLadderInvestment(strategy, binanceService)
	case "price_trigger":
		return dis.executePriceTriggerInvestment(strategy, binanceService)
	default:
		return errors.New("不支持的投资类型")
	}
}

func (dis *DualInvestmentService) executeSingleInvestment(strategy *models.DualInvestmentStrategy, binanceService *BinanceService) error {
	var existingOrder models.DualInvestmentOrder
	if err := dis.db.Where("strategy_id = ? AND status NOT IN ?", strategy.ID, []string{"SETTLED", "FAILED"}).First(&existingOrder).Error; err == nil {
		return nil
	}

	return dis.createDualInvestmentOrder(strategy, strategy.Amount, binanceService)
}

func (dis *DualInvestmentService) executeAutoReinvestment(strategy *models.DualInvestmentStrategy, binanceService *BinanceService) error {
	var lastOrder models.DualInvestmentOrder
	if err := dis.db.Where("strategy_id = ?", strategy.ID).Order("created_at desc").First(&lastOrder).Error; err != nil {
		return dis.createDualInvestmentOrder(strategy, strategy.Amount, binanceService)
	}

	if lastOrder.Status == "SETTLED" {
		return dis.createDualInvestmentOrder(strategy, strategy.Amount, binanceService)
	}

	return nil
}

func (dis *DualInvestmentService) executeLadderInvestment(strategy *models.DualInvestmentStrategy, binanceService *BinanceService) error {
	var completedSteps int64
	dis.db.Model(&models.DualInvestmentOrder{}).Where("strategy_id = ? AND status = ?", strategy.ID, "PURCHASED").Count(&completedSteps)

	if int(completedSteps) >= strategy.LadderSteps {
		return nil
	}

	return dis.createDualInvestmentOrder(strategy, strategy.AmountPerStep, binanceService)
}

func (dis *DualInvestmentService) executePriceTriggerInvestment(strategy *models.DualInvestmentStrategy, binanceService *BinanceService) error {
	if strategy.TriggerPrice <= 0 {
		return errors.New("价格触发投资需要设置触发价格")
	}

	symbol := strategy.BaseAsset + strategy.QuoteAsset
	currentPrice, err := binanceService.GetPrice(context.Background(), symbol)
	if err != nil {
		return err
	}

	if currentPrice < strategy.TriggerPrice {
		return nil
	}

	var existingOrder models.DualInvestmentOrder
	if err := dis.db.Where("strategy_id = ? AND status NOT IN ?", strategy.ID, []string{"SETTLED", "FAILED"}).First(&existingOrder).Error; err == nil {
		return nil
	}

	return dis.createDualInvestmentOrder(strategy, strategy.Amount, binanceService)
}

func (dis *DualInvestmentService) createDualInvestmentOrder(strategy *models.DualInvestmentStrategy, amount float64, binanceService *BinanceService) error {
	var product models.DualInvestmentProduct
	if err := dis.db.Where("product_id = ?", strategy.ProductID).First(&product).Error; err != nil {
		return err
	}

	if amount < product.MinAmount || amount > product.MaxAmount {
		return errors.New("投资金额超出产品限制")
	}

	if product.YieldRate < strategy.MinYieldRate {
		return nil
	}

	order := &models.DualInvestmentOrder{
		UserID:         strategy.UserID,
		StrategyID:     &strategy.ID,
		ProductID:      strategy.ProductID,
		OrderID:        utils.GenerateUUID(),
		Amount:         amount,
		Currency:       strategy.QuoteAsset,
		YieldRate:      product.YieldRate,
		Duration:       product.Duration,
		SettlementDate: product.SettlementDate,
		Status:         "PURCHASED",
		PurchaseTime:   time.Now().Unix(),
	}

	if err := dis.db.Create(order).Error; err != nil {
		return err
	}

	log.Printf("创建双币投资订单成功: %s, 金额: %.2f, 收益率: %.2f%%",
		order.OrderID, order.Amount, order.YieldRate*100)

	return nil
}

func (dis *DualInvestmentService) GetUserDualInvestmentOrders(userID uint, page, limit int) ([]models.DualInvestmentOrder, int64, error) {
	var orders []models.DualInvestmentOrder
	var total int64

	query := dis.db.Model(&models.DualInvestmentOrder{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at desc").Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (dis *DualInvestmentService) SettleDualInvestmentOrders() error {
	var orders []models.DualInvestmentOrder
	currentTime := time.Now().Unix()

	if err := dis.db.Where("status = ? AND settlement_time <= ?", "PURCHASED", currentTime).Find(&orders).Error; err != nil {
		return err
	}

	for _, order := range orders {
		order.Status = "SETTLED"
		order.SettlementTime = currentTime

		if err := dis.db.Save(&order).Error; err != nil {
			log.Printf("结算双币投资订单失败: %v", err)
			continue
		}

		log.Printf("结算双币投资订单成功: %s", order.OrderID)
	}

	return nil
}

func (dis *DualInvestmentService) GetDualInvestmentStats(userID uint) (map[string]interface{}, error) {
	var totalStrategies int64
	var activeStrategies int64
	var totalOrders int64
	var totalAmount float64
	var totalYield float64

	dis.db.Model(&models.DualInvestmentStrategy{}).Where("user_id = ?", userID).Count(&totalStrategies)
	dis.db.Model(&models.DualInvestmentStrategy{}).Where("user_id = ? AND is_active = ?", userID, true).Count(&activeStrategies)
	dis.db.Model(&models.DualInvestmentOrder{}).Where("user_id = ?", userID).Count(&totalOrders)

	var orders []models.DualInvestmentOrder
	dis.db.Where("user_id = ?", userID).Find(&orders)

	for _, order := range orders {
		totalAmount += order.Amount
		if order.Status == "SETTLED" {
			totalYield += order.Amount * order.YieldRate
		}
	}

	stats := map[string]interface{}{
		"total_strategies":  totalStrategies,
		"active_strategies": activeStrategies,
		"total_orders":      totalOrders,
		"total_amount":      totalAmount,
		"total_yield":       totalYield,
		"yield_rate":        utils.CalculatePercent(totalYield, totalAmount),
	}

	return stats, nil
}

func (dis *DualInvestmentService) SyncDualInvestmentProducts() error {
	mockProducts := []models.DualInvestmentProduct{
		{
			ProductID:      "BTC001",
			ProductName:    "BTC双币投资7天期",
			BaseAsset:      "BTC",
			QuoteAsset:     "USDT",
			MinAmount:      100,
			MaxAmount:      100000,
			Duration:       7,
			SettlementDate: time.Now().AddDate(0, 0, 7).Format("2006-01-02"),
			DeliveryPrice:  50000,
			YieldRate:      0.05,
			IsActive:       true,
		},
		{
			ProductID:      "ETH001",
			ProductName:    "ETH双币投资14天期",
			BaseAsset:      "ETH",
			QuoteAsset:     "USDT",
			MinAmount:      50,
			MaxAmount:      50000,
			Duration:       14,
			SettlementDate: time.Now().AddDate(0, 0, 14).Format("2006-01-02"),
			DeliveryPrice:  3000,
			YieldRate:      0.08,
			IsActive:       true,
		},
		{
			ProductID:      "BNB001",
			ProductName:    "BNB双币投资30天期",
			BaseAsset:      "BNB",
			QuoteAsset:     "USDT",
			MinAmount:      10,
			MaxAmount:      10000,
			Duration:       30,
			SettlementDate: time.Now().AddDate(0, 0, 30).Format("2006-01-02"),
			DeliveryPrice:  300,
			YieldRate:      0.12,
			IsActive:       true,
		},
	}

	for _, product := range mockProducts {
		var existing models.DualInvestmentProduct
		if err := dis.db.Where("product_id = ?", product.ProductID).First(&existing).Error; err != nil {
			if err := dis.db.Create(&product).Error; err != nil {
				log.Printf("创建双币产品失败: %v", err)
			}
		} else {
			if err := dis.db.Model(&existing).Updates(&product).Error; err != nil {
				log.Printf("更新双币产品失败: %v", err)
			}
		}
	}

	return nil
}
