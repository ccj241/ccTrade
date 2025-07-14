package migrations

import (
	"github.com/ccj241/cctrade/config"
	"github.com/ccj241/cctrade/models"
	"log"
)

func RunMigrations() error {
	log.Println("开始数据库迁移...")

	if err := config.AutoMigrate(); err != nil {
		return err
	}

	if err := createIndexes(); err != nil {
		return err
	}

	if err := createDefaultAdmin(); err != nil {
		return err
	}

	log.Println("数据库迁移完成")
	return nil
}

func createIndexes() error {
	log.Println("创建数据库索引...")

	db := config.DB

	queries := []string{
		"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)",
		"CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)",
		"CREATE INDEX IF NOT EXISTS idx_users_status ON users(status)",
		"CREATE INDEX IF NOT EXISTS idx_strategies_user_id ON strategies(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_strategies_symbol ON strategies(symbol)",
		"CREATE INDEX IF NOT EXISTS idx_strategies_is_active ON strategies(is_active)",
		"CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_orders_strategy_id ON orders(strategy_id)",
		"CREATE INDEX IF NOT EXISTS idx_orders_symbol ON orders(symbol)",
		"CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status)",
		"CREATE INDEX IF NOT EXISTS idx_orders_order_id ON orders(order_id)",
		"CREATE INDEX IF NOT EXISTS idx_trades_order_id ON trades(order_id)",
		"CREATE INDEX IF NOT EXISTS idx_trades_symbol ON trades(symbol)",
		"CREATE INDEX IF NOT EXISTS idx_prices_symbol ON prices(symbol)",
		"CREATE INDEX IF NOT EXISTS idx_futures_strategies_user_id ON futures_strategies(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_futures_strategies_symbol ON futures_strategies(symbol)",
		"CREATE INDEX IF NOT EXISTS idx_futures_orders_user_id ON futures_orders(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_futures_orders_strategy_id ON futures_orders(strategy_id)",
		"CREATE INDEX IF NOT EXISTS idx_futures_orders_symbol ON futures_orders(symbol)",
		"CREATE INDEX IF NOT EXISTS idx_futures_positions_user_id ON futures_positions(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_futures_positions_symbol ON futures_positions(symbol)",
		"CREATE INDEX IF NOT EXISTS idx_dual_investment_strategies_user_id ON dual_investment_strategies(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_dual_investment_orders_user_id ON dual_investment_orders(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_dual_investment_orders_strategy_id ON dual_investment_orders(strategy_id)",
		"CREATE INDEX IF NOT EXISTS idx_withdrawals_user_id ON withdrawals(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_withdrawals_asset ON withdrawals(asset)",
		"CREATE INDEX IF NOT EXISTS idx_withdrawal_histories_user_id ON withdrawal_histories(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_withdrawal_histories_withdrawal_id ON withdrawal_histories(withdrawal_id)",
	}

	for _, query := range queries {
		if err := db.Exec(query).Error; err != nil {
			log.Printf("创建索引失败: %s, 错误: %v", query, err)
		}
	}

	log.Println("数据库索引创建完成")
	return nil
}

func createDefaultAdmin() error {
	log.Println("创建默认管理员账户...")

	var count int64
	config.DB.Model(&models.User{}).Where("role = ?", models.RoleAdmin).Count(&count)

	if count > 0 {
		log.Println("管理员账户已存在，跳过创建")
		return nil
	}

	admin := &models.User{
		Username: "admin",
		Email:    "admin@binance-trading.com",
		Role:     models.RoleAdmin,
		Status:   models.StatusActive,
	}

	if err := admin.HashPassword("admin123456"); err != nil {
		return err
	}

	if err := config.DB.Create(admin).Error; err != nil {
		return err
	}

	log.Println("默认管理员账户创建成功 - 用户名: admin, 密码: admin123456")
	return nil
}

func CreateTestData() error {
	log.Println("创建测试数据...")

	testUser := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
		Status:   models.StatusActive,
	}

	if err := testUser.HashPassword("test123456"); err != nil {
		return err
	}

	var existingUser models.User
	if err := config.DB.Where("username = ?", testUser.Username).First(&existingUser).Error; err != nil {
		if err := config.DB.Create(testUser).Error; err != nil {
			return err
		}
		log.Println("测试用户创建成功 - 用户名: testuser, 密码: test123456")
	}

	testStrategies := []models.Strategy{
		{
			UserID:       testUser.ID,
			Name:         "BTC网格策略",
			Symbol:       "BTCUSDT",
			Type:         models.StrategyGrid,
			Side:         models.OrderSideBuy,
			Quantity:     0.001,
			Price:        45000,
			TriggerPrice: 44000,
			Config: models.StrategyConfig{
				"upper_price": 50000.0,
				"lower_price": 40000.0,
				"grid_count":  10.0,
			},
		},
		{
			UserID:   testUser.ID,
			Name:     "ETH定投策略",
			Symbol:   "ETHUSDT",
			Type:     models.StrategyDCA,
			Side:     models.OrderSideBuy,
			Quantity: 0.01,
			Config: models.StrategyConfig{
				"interval":     24.0,
				"total_amount": 1000.0,
			},
		},
	}

	for _, strategy := range testStrategies {
		var existing models.Strategy
		if err := config.DB.Where("user_id = ? AND name = ?", strategy.UserID, strategy.Name).First(&existing).Error; err != nil {
			if err := config.DB.Create(&strategy).Error; err != nil {
				log.Printf("创建测试策略失败: %v", err)
			}
		}
	}

	testPrices := []models.Price{
		{Symbol: "BTCUSDT", Price: 45000.50},
		{Symbol: "ETHUSDT", Price: 3000.25},
		{Symbol: "BNBUSDT", Price: 300.75},
		{Symbol: "ADAUSDT", Price: 0.45},
		{Symbol: "DOTUSDT", Price: 6.20},
	}

	for _, price := range testPrices {
		var existing models.Price
		if err := config.DB.Where("symbol = ?", price.Symbol).First(&existing).Error; err != nil {
			if err := config.DB.Create(&price).Error; err != nil {
				log.Printf("创建测试价格数据失败: %v", err)
			}
		}
	}

	log.Println("测试数据创建完成")
	return nil
}
