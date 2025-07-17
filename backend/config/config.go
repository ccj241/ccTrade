package config

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Redis    RedisConfig    `json:"redis"`
	JWT      JWTConfig      `json:"jwt"`
	Binance  BinanceConfig  `json:"binance"`
	Security SecurityConfig `json:"security"`
}

type ServerConfig struct {
	Host         string `json:"host"`
	Port         string `json:"port"`
	Mode         string `json:"mode"`
	ReadTimeout  int    `json:"read_timeout"`
	WriteTimeout int    `json:"write_timeout"`
	CORSOrigins  string `json:"cors_origins"`
}

type DatabaseConfig struct {
	Host            string `json:"host"`
	Port            string `json:"port"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	Database        string `json:"database"`
	Charset         string `json:"charset"`
	MaxIdleConns    int    `json:"max_idle_conns"`
	MaxOpenConns    int    `json:"max_open_conns"`
	ConnMaxLifetime int    `json:"conn_max_lifetime"`
}

type RedisConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
	PoolSize int    `json:"pool_size"`
}

type JWTConfig struct {
	Secret     string `json:"secret"`
	ExpiresIn  int    `json:"expires_in"`
	Issuer     string `json:"issuer"`
	RefreshTTL int    `json:"refresh_ttl"`
}

type BinanceConfig struct {
	BaseURL          string `json:"base_url"`
	FuturesBaseURL   string `json:"futures_base_url"`
	TestNet          bool   `json:"testnet"`
	Timeout          int    `json:"timeout"`
	SpotAgentCode    string `json:"spot_agent_code"`
	FuturesAgentCode string `json:"futures_agent_code"`
	RecvWindow       int64  `json:"recv_window"`
	// 全局API密钥配置（用于测试或默认配置）
	GlobalAPIKey    string `json:"global_api_key"`
	GlobalSecretKey string `json:"global_secret_key"`
}

type SecurityConfig struct {
	EncryptionKey    string `json:"encryption_key"`
	PasswordMinLen   int    `json:"password_min_len"`
	MaxLoginAttempts int    `json:"max_login_attempts"`
	RateLimitRPM     int    `json:"rate_limit_rpm"`
}

var AppConfig *Config

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("未找到.env文件，使用环境变量")
	}

	config := &Config{
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Port:         getEnv("SERVER_PORT", "8080"),
			Mode:         getEnv("GIN_MODE", "debug"),
			ReadTimeout:  getEnvAsInt("READ_TIMEOUT", 30),
			WriteTimeout: getEnvAsInt("WRITE_TIMEOUT", 30),
			CORSOrigins:  getEnv("CORS_ORIGINS", "*"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "3306"),
			Username:        getEnv("DB_USERNAME", "root"),
			Password:        getEnv("DB_PASSWORD", ""),
			Database:        getEnv("DB_DATABASE", "binance_trading"),
			Charset:         getEnv("DB_CHARSET", "utf8mb4"),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 25),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 100),
			ConnMaxLifetime: getEnvAsInt("DB_CONN_MAX_LIFETIME", 300),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
			PoolSize: getEnvAsInt("REDIS_POOL_SIZE", 10),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "your-secret-key"),
			ExpiresIn:  getEnvAsInt("JWT_EXPIRES_IN", 3600),
			Issuer:     getEnv("JWT_ISSUER", "binance-trading-system"),
			RefreshTTL: getEnvAsInt("JWT_REFRESH_TTL", 86400),
		},
		Binance: BinanceConfig{
			BaseURL:          getEnv("BINANCE_BASE_URL", "https://api.binance.com"),
			FuturesBaseURL:   getEnv("BINANCE_FUTURES_BASE_URL", "https://fapi.binance.com"),
			TestNet:          getEnvAsBool("BINANCE_TESTNET", false),
			Timeout:          getEnvAsInt("BINANCE_TIMEOUT", 30),
			SpotAgentCode:    getEnv("BINANCE_SPOT_AGENT_CODE", "JW9QZKMK"),
			FuturesAgentCode: getEnv("BINANCE_FUTURES_AGENT_CODE", "mNY8WNSQ"),
			RecvWindow:       getEnvAsInt64("BINANCE_RECV_WINDOW", 60000),
			GlobalAPIKey:     getEnv("BINANCE_API_KEY", ""),
			GlobalSecretKey:  getEnv("BINANCE_SECRET_KEY", ""),
		},
		Security: SecurityConfig{
			EncryptionKey:    "", // Will be set below
			PasswordMinLen:   getEnvAsInt("PASSWORD_MIN_LEN", 8),
			MaxLoginAttempts: getEnvAsInt("MAX_LOGIN_ATTEMPTS", 5),
			RateLimitRPM:     getEnvAsInt("RATE_LIMIT_RPM", 60),
		},
	}

	// 处理加密密钥
	encryptionKey := os.Getenv("ENCRYPTION_KEY")
	if encryptionKey == "" {
		log.Println("警告: ENCRYPTION_KEY 环境变量未设置，生成临时密钥")
		// 生成默认的32字符密钥
		generatedKey, err := generateRandomKey(32)
		if err != nil {
			log.Printf("生成临时密钥失败: %v，使用预设密钥", err)
			encryptionKey = "binance-trading-system-default-key"
		} else {
			encryptionKey = generatedKey
		}
		log.Printf("已生成临时加密密钥，建议设置 ENCRYPTION_KEY 环境变量")
	}

	// 验证密钥长度（AES-256需要32字节）
	if len(encryptionKey) != 32 {
		log.Printf("警告: ENCRYPTION_KEY 长度不是32字符（当前长度: %d），自动调整", len(encryptionKey))
		if len(encryptionKey) > 32 {
			encryptionKey = encryptionKey[:32]
		} else {
			// 补齐到32字符
			encryptionKey = encryptionKey + "0123456789abcdefghijklmnopqrstuvwxyz"[:32-len(encryptionKey)]
		}
		log.Printf("已调整加密密钥长度到32字符")
	}

	config.Security.EncryptionKey = encryptionKey

	// 验证JWT密钥
	if config.JWT.Secret == "your-secret-key" {
		log.Println("警告: JWT_SECRET 使用默认值，请设置安全的密钥")
		// 生成随机JWT密钥
		if randomKey, err := generateRandomKey(32); err == nil {
			config.JWT.Secret = randomKey
			log.Println("已生成随机JWT密钥，请将其保存到环境变量中")
		}
	}

	AppConfig = config
	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// generateRandomKey 生成指定长度的随机密钥
func generateRandomKey(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}
