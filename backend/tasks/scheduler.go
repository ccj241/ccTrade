package tasks

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/ccj241/cctrade/config"
	"github.com/ccj241/cctrade/models"
	"github.com/ccj241/cctrade/services"
)

type Scheduler struct {
	ctx                   context.Context
	cancel                context.CancelFunc
	strategyService       *services.StrategyService
	futuresService        *services.FuturesService
	dualInvestmentService *services.DualInvestmentService
	withdrawalService     *services.WithdrawalService
	userService           *services.UserService
}

func NewScheduler() *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		ctx:                   ctx,
		cancel:                cancel,
		strategyService:       services.NewStrategyService(),
		futuresService:        services.NewFuturesService(),
		dualInvestmentService: services.NewDualInvestmentService(),
		withdrawalService:     services.NewWithdrawalService(),
		userService:           services.NewUserService(),
	}
}

func (s *Scheduler) Start() {
	log.Println("启动定时任务调度器")

	go s.priceMonitorTask()
	go s.orderCheckTask()
	go s.withdrawalCheckTask()
	go s.dualInvestmentTask()
	go s.futuresMonitorTask()

	log.Println("所有定时任务已启动")
}

func (s *Scheduler) Stop() {
	log.Println("停止定时任务调度器")
	s.cancel()
}

func (s *Scheduler) priceMonitorTask() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	log.Println("价格监控任务已启动，每30秒检查一次")

	for {
		select {
		case <-s.ctx.Done():
			log.Println("价格监控任务已停止")
			return
		case <-ticker.C:
			if err := s.updatePrices(); err != nil {
				log.Printf("更新价格失败: %v", err)
			}
		}
	}
}

func (s *Scheduler) orderCheckTask() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	log.Println("订单检查任务已启动，每30秒检查一次")

	for {
		select {
		case <-s.ctx.Done():
			log.Println("订单检查任务已停止")
			return
		case <-ticker.C:
			if err := s.checkOrders(); err != nil {
				log.Printf("检查订单失败: %v", err)
			}

			if err := s.executeActiveStrategies(); err != nil {
				log.Printf("执行活跃策略失败: %v", err)
			}
		}
	}
}

func (s *Scheduler) withdrawalCheckTask() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	log.Println("提币检查任务已启动，每5分钟检查一次")

	for {
		select {
		case <-s.ctx.Done():
			log.Println("提币检查任务已停止")
			return
		case <-ticker.C:
			if err := s.withdrawalService.CheckWithdrawalRules(); err != nil {
				log.Printf("检查提币规则失败: %v", err)
			}
		}
	}
}

func (s *Scheduler) dualInvestmentTask() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	log.Println("双币投资任务已启动，每1小时检查一次")

	for {
		select {
		case <-s.ctx.Done():
			log.Println("双币投资任务已停止")
			return
		case <-ticker.C:
			if err := s.executeDualInvestmentStrategies(); err != nil {
				log.Printf("执行双币投资策略失败: %v", err)
			}

			if err := s.dualInvestmentService.SettleDualInvestmentOrders(); err != nil {
				log.Printf("结算双币投资订单失败: %v", err)
			}
		}
	}
}

func (s *Scheduler) futuresMonitorTask() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	log.Println("期货监控任务已启动，每1分钟检查一次")

	for {
		select {
		case <-s.ctx.Done():
			log.Println("期货监控任务已停止")
			return
		case <-ticker.C:
			if err := s.executeActiveFuturesStrategies(); err != nil {
				log.Printf("执行期货策略失败: %v", err)
			}

			if err := s.updateAllPositions(); err != nil {
				log.Printf("更新持仓信息失败: %v", err)
			}
		}
	}
}

func (s *Scheduler) updatePrices() error {
	symbols := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT", "ADAUSDT", "DOTUSDT", "XRPUSDT", "LTCUSDT", "LINKUSDT"}

	for _, symbol := range symbols {
		var user models.User
		if err := config.DB.Where("api_key != '' AND secret_key != ''").First(&user).Error; err != nil {
			continue
		}

		apiKey, secretKey, err := s.userService.GetUserAPIKeys(user.ID)
		if err != nil {
			continue
		}

		binanceService, err := services.NewBinanceService(apiKey, secretKey)
		if err != nil {
			log.Printf("创建Binance服务失败: %v", err)
			continue
		}

		price, err := binanceService.GetPrice(context.Background(), symbol)
		if err != nil {
			log.Printf("获取%s价格失败: %v", symbol, err)
			continue
		}

		var priceModel models.Price
		if err := config.DB.Where("symbol = ?", symbol).First(&priceModel).Error; err != nil {
			priceModel = models.Price{
				Symbol: symbol,
				Price:  price,
			}
			config.DB.Create(&priceModel)
		} else {
			config.DB.Model(&priceModel).Update("price", price)
		}

		ctx := context.Background()
		config.Redis.Set(ctx, "price:"+symbol, price, 5*time.Minute)
	}

	return nil
}

func (s *Scheduler) checkOrders() error {
	var orders []models.Order
	if err := config.DB.Where("status IN ?", []string{"NEW", "PARTIALLY_FILLED"}).Find(&orders).Error; err != nil {
		return err
	}

	for _, order := range orders {
		apiKey, secretKey, err := s.userService.GetUserAPIKeys(order.UserID)
		if err != nil {
			continue
		}

		binanceService, err := services.NewBinanceService(apiKey, secretKey)
		if err != nil {
			log.Printf("创建Binance服务失败: %v", err)
			continue
		}

		orderStatus, err := binanceService.GetSpotOrderStatus(context.Background(), order.Symbol, order.OrderID)
		if err != nil {
			log.Printf("检查订单状态失败: %v", err)
			continue
		}

		executedQty, _ := strconv.ParseFloat(orderStatus.ExecutedQuantity, 64)
		cumulativeQuoteQty, _ := strconv.ParseFloat(orderStatus.CummulativeQuoteQuantity, 64)

		config.DB.Model(&order).Updates(map[string]interface{}{
			"status":                orderStatus.Status,
			"executed_qty":          executedQty,
			"cumulative_quote_qty": cumulativeQuoteQty,
		})

		// 如果订单关联了策略，并且是慢冰山策略，更新策略状态
		if order.StrategyID != nil && (orderStatus.Status == "FILLED" || orderStatus.Status == "PARTIALLY_FILLED") {
			var strategy models.Strategy
			if err := config.DB.First(&strategy, *order.StrategyID).Error; err == nil {
				if strategy.Type == models.StrategySlowIceberg && strategy.State != nil {
					strategyState := strategy.State
					
					// 更新已成交数量
					if orderStatus.Status == "FILLED" {
						// 订单完全成交，当前层已成交数量清零
						strategyState["layer_filled_quantity"] = 0.0
					} else if orderStatus.Status == "PARTIALLY_FILLED" {
						// 部分成交，更新当前层已成交数量
						currentLayerFilled, _ := strategyState["layer_filled_quantity"].(float64)
						strategyState["layer_filled_quantity"] = currentLayerFilled + executedQty - order.ExecutedQty
					}
					
					// 更新总成交数量
					totalFilled, _ := strategyState["total_filled_quantity"].(float64)
					strategyState["total_filled_quantity"] = totalFilled + executedQty - order.ExecutedQty
					
					// 保存策略状态
					config.DB.Model(&strategy).Update("state", strategyState)
				}
			}
		}
	}

	return nil
}

func (s *Scheduler) executeActiveStrategies() error {
	var strategies []models.Strategy
	if err := config.DB.Where("is_active = ? AND is_completed = ?", true, false).Find(&strategies).Error; err != nil {
		return err
	}

	for _, strategy := range strategies {
		if err := s.strategyService.ExecuteStrategy(&strategy); err != nil {
			log.Printf("执行策略%d失败: %v", strategy.ID, err)
		}
	}

	return nil
}

func (s *Scheduler) executeActiveFuturesStrategies() error {
	var strategies []models.FuturesStrategy
	if err := config.DB.Where("is_active = ? AND is_completed = ?", true, false).Find(&strategies).Error; err != nil {
		return err
	}

	for _, strategy := range strategies {
		if err := s.futuresService.ExecuteFuturesStrategy(&strategy); err != nil {
			log.Printf("执行期货策略%d失败: %v", strategy.ID, err)
		}
	}

	return nil
}

func (s *Scheduler) executeDualInvestmentStrategies() error {
	var strategies []models.DualInvestmentStrategy
	if err := config.DB.Where("is_active = ?", true).Find(&strategies).Error; err != nil {
		return err
	}

	for _, strategy := range strategies {
		if err := s.dualInvestmentService.ExecuteDualInvestmentStrategy(&strategy); err != nil {
			log.Printf("执行双币投资策略%d失败: %v", strategy.ID, err)
		}
	}

	return nil
}

func (s *Scheduler) updateAllPositions() error {
	var users []models.User
	if err := config.DB.Where("api_key != '' AND secret_key != ''").Find(&users).Error; err != nil {
		return err
	}

	for _, user := range users {
		if err := s.futuresService.UpdatePositions(user.ID); err != nil {
			log.Printf("更新用户%d持仓失败: %v", user.ID, err)
		}
	}

	return nil
}
