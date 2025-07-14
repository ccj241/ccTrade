package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ccj241/cctrade/models"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB    *gorm.DB
	Redis *redis.Client
)

func InitDatabase(config *Config) error {
	var err error
	
	// 使用环境变量判断是否使用SQLite进行开发
	if os.Getenv("USE_SQLITE") == "true" {
		// 使用SQLite作为开发数据库
		DB, err = gorm.Open(sqlite.Open("binance_trading.db"), &gorm.Config{
			Logger:                                   logger.Default.LogMode(logger.Info),
			DisableForeignKeyConstraintWhenMigrating: true,
		})
		if err != nil {
			return fmt.Errorf("SQLite数据库连接失败: %v", err)
		}
		log.Println("使用SQLite数据库进行开发")
	} else {
		// 使用MySQL作为生产数据库
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
			config.Database.Username,
			config.Database.Password,
			config.Database.Host,
			config.Database.Port,
			config.Database.Database,
			config.Database.Charset,
		)
		
		DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger:                                   logger.Default.LogMode(logger.Info),
			DisableForeignKeyConstraintWhenMigrating: true,
		})
		if err != nil {
			return fmt.Errorf("MySQL数据库连接失败: %v", err)
		}
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("获取数据库实例失败: %v", err)
	}

	sqlDB.SetMaxIdleConns(config.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(config.Database.ConnMaxLifetime) * time.Second)

	log.Println("数据库连接成功")
	return nil
}

func InitRedis(config *Config) error {
	Redis = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.Redis.Host, config.Redis.Port),
		Password: config.Redis.Password,
		DB:       config.Redis.DB,
		PoolSize: config.Redis.PoolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := Redis.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("redis连接失败: %v", err)
	}

	log.Println("Redis连接成功")
	return nil
}

func AutoMigrate() error {
	err := DB.AutoMigrate(
		&models.User{},
		&models.Strategy{},
		&models.Order{},
		&models.Trade{},
		&models.Price{},
		&models.FuturesStrategy{},
		&models.FuturesOrder{},
		&models.FuturesPosition{},
		&models.DualInvestmentProduct{},
		&models.DualInvestmentStrategy{},
		&models.DualInvestmentOrder{},
		&models.Withdrawal{},
		&models.WithdrawalHistory{},
	)
	if err != nil {
		return fmt.Errorf("数据库迁移失败: %v", err)
	}

	log.Println("数据库迁移完成")
	return nil
}

func CloseDatabase() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

func CloseRedis() error {
	if Redis != nil {
		return Redis.Close()
	}
	return nil
}
