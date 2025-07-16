package routes

import (
	"github.com/ccj241/cctrade/controllers"
	"github.com/ccj241/cctrade/middleware"
	"github.com/ccj241/cctrade/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"time"
)

func SetupRoutes(r *gin.Engine, quantService *services.QuantitativeStrategy, executor *services.StrategyExecutor, logger *zap.Logger) {
	authController := controllers.NewAuthController()
	strategyController := controllers.NewStrategyController()
	futuresController := controllers.NewFuturesController()
	dualInvestmentController := controllers.NewDualInvestmentController()
	withdrawalController := controllers.NewWithdrawalController()
	generalController := controllers.NewGeneralController()
	quantitativeController := controllers.NewQuantitativeController(executor)

	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.ErrorLoggerMiddleware())
	r.Use(middleware.HealthCheckMiddleware())
	r.Use(middleware.RateLimitMiddleware())

	api := r.Group("/api")
	{
		api.POST("/register", authController.Register)
		api.POST("/login", authController.Login)

		authenticated := api.Group("")
		authenticated.Use(middleware.AuthMiddleware())
		{
			authenticated.GET("/profile", authController.GetProfile)
			authenticated.PUT("/profile", authController.UpdateProfile)
			authenticated.POST("/change-password", authController.ChangePassword)
			authenticated.POST("/api-keys", authController.UpdateAPIKeys)
			authenticated.POST("/api-keys/validate", authController.ValidateAPIKeys)

			authenticated.GET("/balance", generalController.GetBalance)
			authenticated.GET("/orders", generalController.GetOrders)
			authenticated.POST("/order", generalController.CreateOrder)
			authenticated.DELETE("/order/:order_id", generalController.CancelOrder)
			authenticated.POST("/batch-cancel-orders", generalController.BatchCancelOrders)
			authenticated.GET("/cancelled-orders", generalController.GetCancelledOrders)
			authenticated.GET("/trading-symbols", generalController.GetTradingSymbols)
			authenticated.GET("/futures-trading-symbols", generalController.GetFuturesTradingSymbols)
			authenticated.GET("/price", generalController.GetPrice)
			authenticated.GET("/diagnose", controllers.DiagnoseBinanceAPI)

			strategies := authenticated.Group("/strategies")
			strategies.Use(middleware.UserRateLimitMiddleware(100, time.Minute))
			{
				strategies.GET("", strategyController.GetUserStrategies)
				strategies.POST("", strategyController.CreateStrategy)
				strategies.GET("/:strategy_id", strategyController.GetStrategyByID)
				strategies.PUT("/:strategy_id", strategyController.UpdateStrategy)
				strategies.POST("/:strategy_id/toggle", strategyController.ToggleStrategy)
				strategies.DELETE("/:strategy_id", strategyController.DeleteStrategy)
				strategies.GET("/:strategy_id/stats", strategyController.GetStrategyStats)
			}

			futures := authenticated.Group("/futures")
			futures.Use(middleware.UserRateLimitMiddleware(100, time.Minute))
			{
				futures.GET("/strategies", futuresController.GetUserFuturesStrategies)
				futures.POST("/strategy", futuresController.CreateFuturesStrategy)
				futures.GET("/strategy/:strategy_id", futuresController.GetFuturesStrategyByID)
				futures.PUT("/strategy/:strategy_id", futuresController.UpdateFuturesStrategy)
				futures.POST("/strategy/:strategy_id/toggle", futuresController.ToggleFuturesStrategy)
				futures.DELETE("/strategy/:strategy_id", futuresController.DeleteFuturesStrategy)
				futures.GET("/positions", futuresController.GetUserPositions)
				futures.POST("/positions/sync", futuresController.UpdatePositions)
				futures.GET("/stats", futuresController.GetFuturesStats)
			}

			dual := authenticated.Group("/dual")
			dual.Use(middleware.UserRateLimitMiddleware(50, time.Minute))
			{
				dual.GET("/products", dualInvestmentController.GetDualInvestmentProducts)
				dual.GET("/strategies", dualInvestmentController.GetUserDualInvestmentStrategies)
				dual.POST("/strategy", dualInvestmentController.CreateDualInvestmentStrategy)
				dual.GET("/strategy/:strategy_id", dualInvestmentController.GetDualInvestmentStrategyByID)
				dual.PUT("/strategy/:strategy_id", dualInvestmentController.UpdateDualInvestmentStrategy)
				dual.POST("/strategy/:strategy_id/toggle", dualInvestmentController.ToggleDualInvestmentStrategy)
				dual.DELETE("/strategy/:strategy_id", dualInvestmentController.DeleteDualInvestmentStrategy)
				dual.GET("/orders", dualInvestmentController.GetUserDualInvestmentOrders)
				dual.GET("/stats", dualInvestmentController.GetDualInvestmentStats)
			}

			withdrawals := authenticated.Group("/withdrawals")
			withdrawals.Use(middleware.UserRateLimitMiddleware(30, time.Minute))
			{
				withdrawals.GET("", withdrawalController.GetUserWithdrawals)
				withdrawals.POST("", withdrawalController.CreateWithdrawalRule)
				withdrawals.GET("/:withdrawal_id", withdrawalController.GetWithdrawalByID)
				withdrawals.PUT("/:withdrawal_id", withdrawalController.UpdateWithdrawal)
				withdrawals.POST("/:withdrawal_id/toggle", withdrawalController.ToggleWithdrawal)
				withdrawals.DELETE("/:withdrawal_id", withdrawalController.DeleteWithdrawal)
				withdrawals.GET("/history", withdrawalController.GetUserWithdrawalHistory)
				withdrawals.POST("/history/sync", withdrawalController.SyncWithdrawalHistory)
				withdrawals.GET("/stats", withdrawalController.GetWithdrawalStats)
			}

			// 量化策略路由
			quantitative := authenticated.Group("/quantitative")
			quantitative.Use(middleware.UserRateLimitMiddleware(100, time.Minute))
			{
				// 评分系统
				quantitative.GET("/scores", quantitativeController.GetScores)
				quantitative.GET("/scores/:symbol", quantitativeController.GetScoreDetails)
				quantitative.GET("/top-pairs", quantitativeController.GetTopPairs)
				quantitative.GET("/historical-scores", quantitativeController.GetHistoricalScores)

				// 仓位管理
				quantitative.GET("/positions", quantitativeController.GetPositions)
				quantitative.GET("/risk-metrics", quantitativeController.GetRiskMetrics)

				// 策略配置
				quantitative.GET("/config", quantitativeController.GetStrategyConfig)
				quantitative.PUT("/config", quantitativeController.UpdateStrategyConfig)

				// 资金管理
				quantitative.PUT("/capital", quantitativeController.UpdateCapital)
				quantitative.POST("/reset-daily-stats", quantitativeController.ResetDailyStats)

				// 性能统计
				quantitative.GET("/performance", quantitativeController.GetPerformanceStats)
			}
		}

		admin := api.Group("/admin")
		admin.Use(middleware.AuthMiddleware())
		admin.Use(middleware.AdminMiddleware())
		{
			admin.GET("/users", authController.GetAllUsers)
			admin.POST("/users/:user_id/approve", authController.ApproveUser)
			admin.PUT("/users/:user_id/status", authController.UpdateUserStatus)
			admin.PUT("/users/:user_id/role", authController.UpdateUserRole)
			admin.POST("/dual/products/sync", dualInvestmentController.SyncDualInvestmentProducts)
		}
	}
}
