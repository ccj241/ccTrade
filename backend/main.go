package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ccj241/cctrade/config"
	"github.com/ccj241/cctrade/migrations"
	"github.com/ccj241/cctrade/routes"
	"github.com/ccj241/cctrade/services"
	"github.com/ccj241/cctrade/tasks"
	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("启动币安交易系统 v2.0.0")

	appConfig := config.LoadConfig()

	if err := config.InitDatabase(appConfig); err != nil {
		log.Printf("数据库初始化失败: %v", err)
		log.Println("警告：数据库未连接，将使用SQLite或只读模式")
		// 尝试使用SQLite作为备用数据库
		os.Setenv("USE_SQLITE", "true")
		if err := config.InitDatabase(appConfig); err != nil {
			log.Printf("备用SQLite数据库初始化失败: %v", err)
			log.Println("警告：无数据库连接，某些功能将受限")
		}
	}
	defer config.CloseDatabase()

	if err := config.InitRedis(appConfig); err != nil {
		log.Println("Redis初始化失败:", err)
		log.Println("警告：Redis未连接，某些功能可能受限")
		// 不退出程序，继续运行
	} else {
		defer config.CloseRedis()
	}

	// 只有在数据库连接成功时才运行迁移
	if config.DB != nil {
		if err := migrations.RunMigrations(); err != nil {
			log.Printf("数据库迁移失败: %v", err)
			// 继续运行，不退出
		}
	} else {
		log.Println("跳过数据库迁移（无数据库连接）")
	}

	if os.Getenv("CREATE_TEST_DATA") == "true" && config.DB != nil {
		if err := migrations.CreateTestData(); err != nil {
			log.Printf("创建测试数据失败: %v", err)
		}
	}

	gin.SetMode(appConfig.Server.Mode)

	r := gin.New()

	// 初始化策略执行器
	strategyExecutor := services.NewStrategyExecutor(config.DB)
	if err := strategyExecutor.Start(); err != nil {
		log.Printf("策略执行器启动失败: %v", err)
	}
	defer strategyExecutor.Stop()

	routes.SetupRoutes(r, strategyExecutor)

	scheduler := tasks.NewScheduler()
	scheduler.Start()
	defer scheduler.Stop()

	dualInvestmentService := services.NewDualInvestmentService()
	if err := dualInvestmentService.SyncDualInvestmentProducts(); err != nil {
		log.Printf("同步双币投资产品失败: %v", err)
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", appConfig.Server.Host, appConfig.Server.Port),
		Handler:      r,
		ReadTimeout:  time.Duration(appConfig.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(appConfig.Server.WriteTimeout) * time.Second,
	}

	go func() {
		log.Printf("服务器启动在 %s:%s", appConfig.Server.Host, appConfig.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("服务器启动失败: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("正在关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("服务器强制关闭失败: %v", err)
	}

	<-ctx.Done()
	log.Println("超时5秒")

	log.Println("服务器已退出")
}
