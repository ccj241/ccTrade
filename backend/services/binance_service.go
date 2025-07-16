package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/common"
	"github.com/adshao/go-binance/v2/futures"
	"github.com/ccj241/cctrade/config"
	"github.com/ccj241/cctrade/models"
	"github.com/ccj241/cctrade/utils"
	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

var (
	// 错误定义
	ErrInvalidAPIKey       = errors.New("invalid API key")
	ErrInvalidSecretKey    = errors.New("invalid secret key")
	ErrInvalidSymbol       = errors.New("invalid trading symbol")
	ErrInvalidOrderType    = errors.New("invalid order type")
	ErrInvalidOrderSide    = errors.New("invalid order side")
	ErrInvalidQuantity     = errors.New("invalid quantity")
	ErrInvalidPrice        = errors.New("invalid price")
	ErrOrderNotFound       = errors.New("order not found")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrRateLimitExceeded   = errors.New("rate limit exceeded")
	ErrAPIPermissionDenied = errors.New("API permission denied")
)

// BinanceError 币安API错误封装
type BinanceError struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
}

// BinanceService 币安服务接口定义
type IBinanceService interface {
	GetSpotClient() (*binance.Client, error)
	GetFuturesClient() (*futures.Client, error)
	GetAccountInfo(ctx context.Context) (*binance.Account, error)
	GetFuturesAccountInfo(ctx context.Context) (*futures.Account, error)
	GetPrice(ctx context.Context, symbol string) (float64, error)
	GetFuturesPrice(ctx context.Context, symbol string) (float64, error)
	CreateSpotOrder(ctx context.Context, order *models.Order) (*binance.CreateOrderResponse, error)
	CreateFuturesOrder(ctx context.Context, order *models.FuturesOrder) (*futures.CreateOrderResponse, error)
	CancelSpotOrder(ctx context.Context, symbol, orderID string) (*binance.CancelOrderResponse, error)
	CancelFuturesOrder(ctx context.Context, symbol, orderID string) (*futures.CancelOrderResponse, error)
	GetSpotOrderStatus(ctx context.Context, symbol, orderID string) (*binance.Order, error)
	GetFuturesOrderStatus(ctx context.Context, symbol, orderID string) (*futures.Order, error)
	GetFuturesPositions(ctx context.Context) ([]*futures.PositionRisk, error)
	SetFuturesLeverage(ctx context.Context, symbol string, leverage int) error
	SetFuturesMarginType(ctx context.Context, symbol string, marginType models.MarginType) error
	GetWithdrawHistory(ctx context.Context, asset string, limit int) ([]*binance.Withdraw, error)
	Withdraw(ctx context.Context, asset, address, network string, amount float64, addressTag string) (*binance.CreateWithdrawResponse, error)
	GetTradingSymbols(ctx context.Context) ([]binance.Symbol, error)
	GetFuturesTradingSymbols(ctx context.Context) ([]futures.Symbol, error)
	ValidateAPICredentials(ctx context.Context) error
	DiagnoseAPIConnection(ctx context.Context) (map[string]interface{}, error)
	GetOrderBook(ctx context.Context, symbol string, limit int) (*binance.DepthResponse, error)
}

// BinanceService 币安服务实现
type BinanceService struct {
	apiKey            string // 存储的是解密后的API密钥
	secretKey         string // 存储的是解密后的Secret密钥
	testNet           bool
	logger            *logrus.Logger
	validator         *validator.Validate
	clientPool        sync.Pool
	futuresClientPool sync.Pool
	rateLimiter       *RateLimiter
	symbolCache       *SymbolCache
	mu                sync.RWMutex
}

// RateLimiter 速率限制器
type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.Mutex
	limit    int
	window   time.Duration
}

// SymbolCache 交易对缓存
type SymbolCache struct {
	symbols    map[string]*SymbolInfo
	lastUpdate time.Time
	mu         sync.RWMutex
	ttl        time.Duration
}

// SymbolInfo 交易对信息
type SymbolInfo struct {
	Symbol              string
	BaseAsset           string
	QuoteAsset          string
	MinQty              float64
	MaxQty              float64
	StepSize            float64
	MinNotional         float64
	PricePrecision      int
	QuantityPrecision   int
	BaseAssetPrecision  int
	QuoteAssetPrecision int
}

// NewBinanceService 创建币安服务实例
func NewBinanceService(apiKey, secretKey string) (*BinanceService, error) {
	if apiKey == "" || secretKey == "" {
		return nil, errors.New("API credentials cannot be empty")
	}

	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})

	// 直接使用传入的密钥（已经是解密后的）
	bs := &BinanceService{
		apiKey:    apiKey,    // 解密后的API密钥
		secretKey: secretKey, // 解密后的Secret密钥
		testNet:   config.AppConfig.Binance.TestNet,
		logger:    logger,
		validator: validator.New(),
		rateLimiter: &RateLimiter{
			requests: make(map[string][]time.Time),
			limit:    1200, // 币安默认限制
			window:   time.Minute,
		},
		symbolCache: &SymbolCache{
			symbols: make(map[string]*SymbolInfo),
			ttl:     24 * time.Hour,
		},
	}

	// 初始化连接池
	bs.clientPool = sync.Pool{
		New: func() interface{} {
			apiKey, secretKey, err := bs.decryptCredentials()
			if err != nil {
				logger.WithError(err).Error("Failed to decrypt credentials")
				return nil
			}

			// 验证API密钥格式
			if len(apiKey) < 10 || len(secretKey) < 10 {
				logger.Error("Invalid API credentials: keys too short")
				return nil
			}

			// 打印调试信息
			logger.WithFields(logrus.Fields{
				"api_key_full":     apiKey,
				"api_key_len":      len(apiKey),
				"secret_key_len":   len(secretKey),
				"api_key_first_10": apiKey[:min(10, len(apiKey))],
				"api_key_last_10":  apiKey[max(0, len(apiKey)-10):],
			}).Info("API Key debug info")

			client := binance.NewClient(apiKey, secretKey)
			if bs.testNet {
				client.BaseURL = "https://testnet.binance.vision"
				logger.WithField("base_url", client.BaseURL).Debug("Using testnet API")
			} else {
				// 确保使用正确的现货API base URL
				client.BaseURL = "https://api.binance.com"
				logger.WithField("base_url", client.BaseURL).Debug("Using mainnet API")
			}

			// 创建自定义的HTTP客户端，添加apiAgentCode
			customTransport := NewBinanceTransport(config.AppConfig.Binance.SpotAgentCode, false)
			client.HTTPClient = &http.Client{
				Timeout:   30 * time.Second,
				Transport: customTransport,
			}
			client.Debug = false

			// 记录API密钥前缀用于调试
			logger.WithField("api_key_prefix", apiKey[:6]+"...").Debug("Spot client initialized")
			return client
		},
	}

	bs.futuresClientPool = sync.Pool{
		New: func() interface{} {
			apiKey, secretKey, err := bs.decryptCredentials()
			if err != nil {
				logger.WithError(err).Error("Failed to decrypt credentials")
				return nil
			}

			// 验证API密钥格式
			if len(apiKey) < 10 || len(secretKey) < 10 {
				logger.Error("Invalid API credentials: keys too short")
				return nil
			}

			client := futures.NewClient(apiKey, secretKey)
			if bs.testNet {
				client.BaseURL = "https://testnet.binancefuture.com"
				logger.WithField("base_url", client.BaseURL).Debug("Using futures testnet API")
			} else {
				// 确保使用正确的期货API base URL
				client.BaseURL = "https://fapi.binance.com"
				logger.WithField("base_url", client.BaseURL).Debug("Using futures mainnet API")
			}

			// 创建自定义的HTTP客户端，添加apiAgentCode
			customTransport := NewBinanceTransport(config.AppConfig.Binance.FuturesAgentCode, true)
			client.HTTPClient = &http.Client{
				Timeout:   30 * time.Second,
				Transport: customTransport,
			}

			// 记录API密钥前缀用于调试
			logger.WithField("api_key_prefix", apiKey[:6]+"...").Debug("Futures client initialized")
			return client
		},
	}

	return bs, nil
}

// min 返回两个整数中的最小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max 返回两个整数中的最大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// cleanAPIKey 清理API密钥，移除控制字符和空白
func cleanAPIKey(key string) string {
	// 移除前后空白
	key = strings.TrimSpace(key)

	// 移除所有控制字符，但保留可打印字符
	var cleaned strings.Builder
	for _, ch := range key {
		// 保留所有可打印的ASCII字符（32-126）
		if ch >= 32 && ch <= 126 {
			cleaned.WriteRune(ch)
		}
	}

	return cleaned.String()
}

// decryptCredentials 获取API凭证（已经是解密后的）
func (bs *BinanceService) decryptCredentials() (apiKey, secretKey string, err error) {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	// 检查API密钥是否为空
	if bs.apiKey == "" || bs.secretKey == "" {
		return "", "", fmt.Errorf("API credentials are empty")
	}

	// 直接返回已经解密的密钥
	apiKey = bs.apiKey
	secretKey = bs.secretKey

	// 验证密钥
	if len(apiKey) == 0 || len(secretKey) == 0 {
		return "", "", fmt.Errorf("API credentials are empty")
	}

	// 彻底清理API密钥（移除可能的控制字符）
	apiKey = cleanAPIKey(apiKey)
	secretKey = cleanAPIKey(secretKey)

	bs.logger.WithFields(logrus.Fields{
		"api_key_len":    len(apiKey),
		"secret_key_len": len(secretKey),
		"api_key_prefix": apiKey[:min(6, len(apiKey))] + "...",
		"is_testnet":     bs.testNet,
	}).Debug("Credentials retrieved successfully")

	return apiKey, secretKey, nil
}

// checkRateLimit 检查速率限制
func (bs *BinanceService) checkRateLimit(identifier string) error {
	bs.rateLimiter.mu.Lock()
	defer bs.rateLimiter.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-bs.rateLimiter.window)

	// 清理过期的请求记录
	if requests, exists := bs.rateLimiter.requests[identifier]; exists {
		validRequests := []time.Time{}
		for _, reqTime := range requests {
			if reqTime.After(windowStart) {
				validRequests = append(validRequests, reqTime)
			}
		}
		bs.rateLimiter.requests[identifier] = validRequests
	}

	// 检查是否超过限制
	if len(bs.rateLimiter.requests[identifier]) >= bs.rateLimiter.limit {
		return ErrRateLimitExceeded
	}

	// 记录新请求
	bs.rateLimiter.requests[identifier] = append(bs.rateLimiter.requests[identifier], now)
	return nil
}

// GetSpotClient 获取现货客户端
func (bs *BinanceService) GetSpotClient() (*binance.Client, error) {
	client := bs.clientPool.Get()
	if client == nil {
		return nil, errors.New("failed to get spot client from pool")
	}
	return client.(*binance.Client), nil
}

// GetFuturesClient 获取期货客户端
func (bs *BinanceService) GetFuturesClient() (*futures.Client, error) {
	client := bs.futuresClientPool.Get()
	if client == nil {
		return nil, errors.New("failed to get futures client from pool")
	}
	return client.(*futures.Client), nil
}

// GetAccountInfo 获取现货账户信息
func (bs *BinanceService) GetAccountInfo(ctx context.Context) (*binance.Account, error) {
	if err := bs.checkRateLimit("account_info"); err != nil {
		return nil, err
	}

	client, err := bs.GetSpotClient()
	if err != nil {
		return nil, err
	}
	defer bs.clientPool.Put(client)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 记录请求详情
	bs.logger.WithFields(logrus.Fields{
		"base_url":    client.BaseURL,
		"testnet":     bs.testNet,
		"recv_window": config.AppConfig.Binance.RecvWindow,
	}).Debug("Attempting to get account info")

	// 使用配置的recvWindow来处理时间同步问题
	account, err := client.NewGetAccountService().Do(ctx, binance.WithRecvWindow(config.AppConfig.Binance.RecvWindow))
	if err != nil {
		bs.logger.WithError(err).WithFields(logrus.Fields{
			"base_url": client.BaseURL,
			"testnet":  bs.testNet,
		}).Error("Failed to get account info")
		return nil, bs.handleBinanceError(err)
	}

	bs.logger.WithFields(logrus.Fields{
		"balances":     len(account.Balances),
		"can_trade":    account.CanTrade,
		"can_withdraw": account.CanWithdraw,
		"can_deposit":  account.CanDeposit,
	}).Info("Account info retrieved")

	return account, nil
}

// GetFuturesAccountInfo 获取期货账户信息
func (bs *BinanceService) GetFuturesAccountInfo(ctx context.Context) (*futures.Account, error) {
	if err := bs.checkRateLimit("futures_account_info"); err != nil {
		return nil, err
	}

	client, err := bs.GetFuturesClient()
	if err != nil {
		return nil, err
	}
	defer bs.futuresClientPool.Put(client)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	account, err := client.NewGetAccountService().Do(ctx)
	if err != nil {
		bs.logger.WithError(err).Error("Failed to get futures account info")
		return nil, bs.handleBinanceError(err)
	}

	return account, nil
}

// validateSymbol 验证交易对
func (bs *BinanceService) validateSymbol(symbol string) error {
	if symbol == "" {
		return ErrInvalidSymbol
	}

	// 检查缓存
	bs.symbolCache.mu.RLock()
	_, exists := bs.symbolCache.symbols[symbol]
	bs.symbolCache.mu.RUnlock()

	if !exists {
		// 更新缓存
		ctx := context.Background()
		if _, err := bs.GetTradingSymbols(ctx); err != nil {
			return err
		}

		// 再次检查
		bs.symbolCache.mu.RLock()
		_, exists = bs.symbolCache.symbols[symbol]
		bs.symbolCache.mu.RUnlock()

		if !exists {
			return fmt.Errorf("%w: %s", ErrInvalidSymbol, symbol)
		}
	}

	return nil
}

// GetPrice 获取现货价格
func (bs *BinanceService) GetPrice(ctx context.Context, symbol string) (float64, error) {
	if err := bs.validateSymbol(symbol); err != nil {
		return 0, err
	}

	if err := bs.checkRateLimit("price_" + symbol); err != nil {
		return 0, err
	}

	client, err := bs.GetSpotClient()
	if err != nil {
		return 0, err
	}
	defer bs.clientPool.Put(client)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	prices, err := client.NewListPricesService().Symbol(symbol).Do(ctx)
	if err != nil {
		bs.logger.WithError(err).WithField("symbol", symbol).Error("Failed to get price")
		return 0, bs.handleBinanceError(err)
	}

	if len(prices) == 0 {
		return 0, errors.New("no price data available")
	}

	price, err := strconv.ParseFloat(prices[0].Price, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse price: %w", err)
	}

	bs.logger.WithFields(logrus.Fields{
		"symbol": symbol,
		"price":  price,
	}).Debug("Price retrieved")

	return price, nil
}

// GetFuturesPrice 获取期货价格
func (bs *BinanceService) GetFuturesPrice(ctx context.Context, symbol string) (float64, error) {
	if err := bs.validateSymbol(symbol); err != nil {
		return 0, err
	}

	if err := bs.checkRateLimit("futures_price_" + symbol); err != nil {
		return 0, err
	}

	client, err := bs.GetFuturesClient()
	if err != nil {
		return 0, err
	}
	defer bs.futuresClientPool.Put(client)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	prices, err := client.NewListPricesService().Symbol(symbol).Do(ctx)
	if err != nil {
		bs.logger.WithError(err).WithField("symbol", symbol).Error("Failed to get futures price")
		return 0, bs.handleBinanceError(err)
	}

	if len(prices) == 0 {
		return 0, errors.New("no price data available")
	}

	price, err := strconv.ParseFloat(prices[0].Price, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse price: %w", err)
	}

	return price, nil
}

// validateOrder 验证订单参数
func (bs *BinanceService) validateOrder(order interface{}) error {
	if order == nil {
		return errors.New("order cannot be nil")
	}

	switch o := order.(type) {
	case *models.Order:
		if err := bs.validateSymbol(o.Symbol); err != nil {
			return err
		}
		if o.Quantity <= 0 {
			return ErrInvalidQuantity
		}
		if o.Type == models.OrderTypeLimit && o.Price <= 0 {
			return ErrInvalidPrice
		}
		if o.Side != models.OrderSideBuy && o.Side != models.OrderSideSell {
			return ErrInvalidOrderSide
		}

	case *models.FuturesOrder:
		if err := bs.validateSymbol(o.Symbol); err != nil {
			return err
		}
		if o.Quantity <= 0 {
			return ErrInvalidQuantity
		}
		if o.Type == models.OrderTypeLimit && o.Price <= 0 {
			return ErrInvalidPrice
		}
		if o.Side != models.OrderSideBuy && o.Side != models.OrderSideSell {
			return ErrInvalidOrderSide
		}

	default:
		return errors.New("invalid order type")
	}

	return nil
}

// CreateSpotOrder 创建现货订单
func (bs *BinanceService) CreateSpotOrder(ctx context.Context, order *models.Order) (*binance.CreateOrderResponse, error) {
	if err := bs.validateOrder(order); err != nil {
		return nil, err
	}

	if err := bs.checkRateLimit("create_order"); err != nil {
		return nil, err
	}

	client, err := bs.GetSpotClient()
	if err != nil {
		return nil, err
	}
	defer bs.clientPool.Put(client)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 获取交易对信息
	symbolInfo, err := bs.getSymbolInfo(order.Symbol)
	if err != nil {
		return nil, err
	}

	// 格式化数量和价格
	formattedQuantity := utils.FormatFloat(order.Quantity, symbolInfo.QuantityPrecision)

	service := client.NewCreateOrderService().
		Symbol(order.Symbol).
		Side(binance.SideType(order.Side)).
		Type(binance.OrderType(order.Type)).
		Quantity(formattedQuantity)

	if order.Price > 0 {
		formattedPrice := utils.FormatFloat(order.Price, symbolInfo.PricePrecision)
		service = service.Price(formattedPrice)
	}

	if order.StopPrice > 0 {
		formattedStopPrice := utils.FormatFloat(order.StopPrice, symbolInfo.PricePrecision)
		service = service.StopPrice(formattedStopPrice)
	}

	if order.TimeInForce != "" {
		service = service.TimeInForce(binance.TimeInForceType(order.TimeInForce))
	}

	if order.ClientOrderID != "" {
		service = service.NewClientOrderID(order.ClientOrderID)
	}

	response, err := service.Do(ctx)
	if err != nil {
		bs.logger.WithError(err).WithFields(logrus.Fields{
			"symbol":   order.Symbol,
			"side":     order.Side,
			"type":     order.Type,
			"quantity": order.Quantity,
			"price":    order.Price,
		}).Error("Failed to create spot order")
		return nil, bs.handleBinanceError(err)
	}

	bs.logger.WithFields(logrus.Fields{
		"order_id":        response.OrderID,
		"client_order_id": response.ClientOrderID,
		"symbol":          response.Symbol,
		"status":          response.Status,
	}).Info("Spot order created")

	return response, nil
}

// CreateFuturesOrder 创建期货订单
func (bs *BinanceService) CreateFuturesOrder(ctx context.Context, order *models.FuturesOrder) (*futures.CreateOrderResponse, error) {
	if err := bs.validateOrder(order); err != nil {
		return nil, err
	}

	if err := bs.checkRateLimit("create_futures_order"); err != nil {
		return nil, err
	}

	client, err := bs.GetFuturesClient()
	if err != nil {
		return nil, err
	}
	defer bs.futuresClientPool.Put(client)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 获取交易对信息
	symbolInfo, err := bs.getSymbolInfo(order.Symbol)
	if err != nil {
		return nil, err
	}

	// 格式化数量和价格
	formattedQuantity := utils.FormatFloat(order.Quantity, symbolInfo.QuantityPrecision)

	service := client.NewCreateOrderService().
		Symbol(order.Symbol).
		Side(futures.SideType(order.Side)).
		Type(futures.OrderType(order.Type)).
		PositionSide(futures.PositionSideType(order.PositionSide)).
		Quantity(formattedQuantity)

	if order.Price > 0 {
		formattedPrice := utils.FormatFloat(order.Price, symbolInfo.PricePrecision)
		service = service.Price(formattedPrice)
	}

	if order.StopPrice > 0 {
		formattedStopPrice := utils.FormatFloat(order.StopPrice, symbolInfo.PricePrecision)
		service = service.StopPrice(formattedStopPrice)
	}

	if order.TimeInForce != "" {
		service = service.TimeInForce(futures.TimeInForceType(order.TimeInForce))
	}

	if order.ClientOrderID != "" {
		service = service.NewClientOrderID(order.ClientOrderID)
	}

	if order.ReduceOnly {
		service = service.ReduceOnly(order.ReduceOnly)
	}

	if order.WorkingType != "" {
		service = service.WorkingType(futures.WorkingType(order.WorkingType))
	}

	response, err := service.Do(ctx)
	if err != nil {
		bs.logger.WithError(err).WithFields(logrus.Fields{
			"symbol":        order.Symbol,
			"side":          order.Side,
			"type":          order.Type,
			"position_side": order.PositionSide,
			"quantity":      order.Quantity,
			"price":         order.Price,
		}).Error("Failed to create futures order")
		return nil, bs.handleBinanceError(err)
	}

	bs.logger.WithFields(logrus.Fields{
		"order_id":        response.OrderID,
		"client_order_id": response.ClientOrderID,
		"symbol":          response.Symbol,
		"status":          response.Status,
	}).Info("Futures order created")

	return response, nil
}

// CancelSpotOrder 取消现货订单
func (bs *BinanceService) CancelSpotOrder(ctx context.Context, symbol, orderID string) (*binance.CancelOrderResponse, error) {
	if err := bs.validateSymbol(symbol); err != nil {
		return nil, err
	}

	if err := bs.checkRateLimit("cancel_order"); err != nil {
		return nil, err
	}

	orderIDInt, err := strconv.ParseInt(orderID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid order ID: %w", err)
	}

	client, err := bs.GetSpotClient()
	if err != nil {
		return nil, err
	}
	defer bs.clientPool.Put(client)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	response, err := client.NewCancelOrderService().Symbol(symbol).OrderID(orderIDInt).Do(ctx)
	if err != nil {
		bs.logger.WithError(err).WithFields(logrus.Fields{
			"symbol":   symbol,
			"order_id": orderID,
		}).Error("Failed to cancel spot order")
		return nil, bs.handleBinanceError(err)
	}

	bs.logger.WithFields(logrus.Fields{
		"order_id": orderID,
		"symbol":   symbol,
		"status":   response.Status,
	}).Info("Spot order cancelled")

	return response, nil
}

// CancelFuturesOrder 取消期货订单
func (bs *BinanceService) CancelFuturesOrder(ctx context.Context, symbol, orderID string) (*futures.CancelOrderResponse, error) {
	if err := bs.validateSymbol(symbol); err != nil {
		return nil, err
	}

	if err := bs.checkRateLimit("cancel_futures_order"); err != nil {
		return nil, err
	}

	orderIDInt, err := strconv.ParseInt(orderID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid order ID: %w", err)
	}

	client, err := bs.GetFuturesClient()
	if err != nil {
		return nil, err
	}
	defer bs.futuresClientPool.Put(client)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	response, err := client.NewCancelOrderService().Symbol(symbol).OrderID(orderIDInt).Do(ctx)
	if err != nil {
		bs.logger.WithError(err).WithFields(logrus.Fields{
			"symbol":   symbol,
			"order_id": orderID,
		}).Error("Failed to cancel futures order")
		return nil, bs.handleBinanceError(err)
	}

	return response, nil
}

// GetSpotOrderStatus 获取现货订单状态
func (bs *BinanceService) GetSpotOrderStatus(ctx context.Context, symbol, orderID string) (*binance.Order, error) {
	if err := bs.validateSymbol(symbol); err != nil {
		return nil, err
	}

	if err := bs.checkRateLimit("order_status"); err != nil {
		return nil, err
	}

	orderIDInt, err := strconv.ParseInt(orderID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid order ID: %w", err)
	}

	client, err := bs.GetSpotClient()
	if err != nil {
		return nil, err
	}
	defer bs.clientPool.Put(client)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	order, err := client.NewGetOrderService().Symbol(symbol).OrderID(orderIDInt).Do(ctx)
	if err != nil {
		return nil, bs.handleBinanceError(err)
	}

	return order, nil
}

// GetFuturesOrderStatus 获取期货订单状态
func (bs *BinanceService) GetFuturesOrderStatus(ctx context.Context, symbol, orderID string) (*futures.Order, error) {
	if err := bs.validateSymbol(symbol); err != nil {
		return nil, err
	}

	if err := bs.checkRateLimit("futures_order_status"); err != nil {
		return nil, err
	}

	orderIDInt, err := strconv.ParseInt(orderID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid order ID: %w", err)
	}

	client, err := bs.GetFuturesClient()
	if err != nil {
		return nil, err
	}
	defer bs.futuresClientPool.Put(client)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	order, err := client.NewGetOrderService().Symbol(symbol).OrderID(orderIDInt).Do(ctx)
	if err != nil {
		return nil, bs.handleBinanceError(err)
	}

	return order, nil
}

// GetFuturesPositions 获取期货持仓
func (bs *BinanceService) GetFuturesPositions(ctx context.Context) ([]*futures.PositionRisk, error) {
	if err := bs.checkRateLimit("futures_positions"); err != nil {
		return nil, err
	}

	client, err := bs.GetFuturesClient()
	if err != nil {
		return nil, err
	}
	defer bs.futuresClientPool.Put(client)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	positions, err := client.NewGetPositionRiskService().Do(ctx)
	if err != nil {
		return nil, bs.handleBinanceError(err)
	}

	// 过滤掉零持仓
	activePositions := make([]*futures.PositionRisk, 0)
	for _, pos := range positions {
		posAmt, _ := strconv.ParseFloat(pos.PositionAmt, 64)
		if posAmt != 0 {
			activePositions = append(activePositions, pos)
		}
	}

	return activePositions, nil
}

// SetFuturesLeverage 设置期货杠杆
func (bs *BinanceService) SetFuturesLeverage(ctx context.Context, symbol string, leverage int) error {
	if err := bs.validateSymbol(symbol); err != nil {
		return err
	}

	if leverage < 1 || leverage > 125 {
		return errors.New("invalid leverage: must be between 1 and 125")
	}

	if err := bs.checkRateLimit("set_leverage"); err != nil {
		return err
	}

	client, err := bs.GetFuturesClient()
	if err != nil {
		return err
	}
	defer bs.futuresClientPool.Put(client)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	_, err = client.NewChangeLeverageService().Symbol(symbol).Leverage(leverage).Do(ctx)
	if err != nil {
		bs.logger.WithError(err).WithFields(logrus.Fields{
			"symbol":   symbol,
			"leverage": leverage,
		}).Error("Failed to set leverage")
		return bs.handleBinanceError(err)
	}

	bs.logger.WithFields(logrus.Fields{
		"symbol":   symbol,
		"leverage": leverage,
	}).Info("Leverage updated")

	return nil
}

// SetFuturesMarginType 设置期货保证金模式
func (bs *BinanceService) SetFuturesMarginType(ctx context.Context, symbol string, marginType models.MarginType) error {
	if err := bs.validateSymbol(symbol); err != nil {
		return err
	}

	if marginType != models.MarginTypeCross && marginType != models.MarginTypeIsolated {
		return errors.New("invalid margin type")
	}

	if err := bs.checkRateLimit("set_margin_type"); err != nil {
		return err
	}

	client, err := bs.GetFuturesClient()
	if err != nil {
		return err
	}
	defer bs.futuresClientPool.Put(client)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err = client.NewChangeMarginTypeService().Symbol(symbol).MarginType(futures.MarginType(marginType)).Do(ctx)
	if err != nil {
		// 如果已经是该保证金模式，币安会返回错误，但这不应该是错误
		if err.Error() == "No need to change margin type." {
			return nil
		}
		return bs.handleBinanceError(err)
	}

	bs.logger.WithFields(logrus.Fields{
		"symbol":      symbol,
		"margin_type": marginType,
	}).Info("Margin type updated")

	return nil
}

// GetWithdrawHistory 获取提现历史
func (bs *BinanceService) GetWithdrawHistory(ctx context.Context, asset string, limit int) ([]*binance.Withdraw, error) {
	if err := bs.checkRateLimit("withdraw_history"); err != nil {
		return nil, err
	}

	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	client, err := bs.GetSpotClient()
	if err != nil {
		return nil, err
	}
	defer bs.clientPool.Put(client)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	service := client.NewListWithdrawsService()
	if asset != "" {
		service = service.Coin(asset)
	}
	service = service.Limit(limit)

	withdrawals, err := service.Do(ctx)
	if err != nil {
		return nil, bs.handleBinanceError(err)
	}

	return withdrawals, nil
}

// Withdraw 提现
func (bs *BinanceService) Withdraw(ctx context.Context, asset, address, network string, amount float64, addressTag string) (*binance.CreateWithdrawResponse, error) {
	if asset == "" || address == "" || amount <= 0 {
		return nil, errors.New("invalid withdrawal parameters")
	}

	if err := bs.checkRateLimit("withdraw"); err != nil {
		return nil, err
	}

	client, err := bs.GetSpotClient()
	if err != nil {
		return nil, err
	}
	defer bs.clientPool.Put(client)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	service := client.NewCreateWithdrawService().
		Coin(asset).
		Address(address).
		Amount(utils.FormatFloat(amount, 8))

	if network != "" {
		service = service.Network(network)
	}

	if addressTag != "" {
		service = service.AddressTag(addressTag)
	}

	response, err := service.Do(ctx)
	if err != nil {
		bs.logger.WithError(err).WithFields(logrus.Fields{
			"asset":   asset,
			"address": address,
			"amount":  amount,
			"network": network,
		}).Error("Failed to create withdrawal")
		return nil, bs.handleBinanceError(err)
	}

	bs.logger.WithFields(logrus.Fields{
		"withdrawal_id": response.ID,
		"asset":         asset,
		"amount":        amount,
	}).Info("Withdrawal created")

	return response, nil
}

// GetKlines 获取K线数据
func (bs *BinanceService) GetKlines(symbol string, interval string, limit int) ([]KlineData, error) {
	if err := bs.checkRateLimit("klines"); err != nil {
		return nil, err
	}

	client, err := bs.GetSpotClient()
	if err != nil {
		return nil, err
	}
	defer bs.clientPool.Put(client)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	klines, err := client.NewKlinesService().
		Symbol(symbol).
		Interval(interval).
		Limit(limit).
		Do(ctx)
	if err != nil {
		return nil, bs.handleBinanceError(err)
	}

	var result []KlineData
	for _, k := range klines {
		openPrice, _ := strconv.ParseFloat(k.Open, 64)
		highPrice, _ := strconv.ParseFloat(k.High, 64)
		lowPrice, _ := strconv.ParseFloat(k.Low, 64)
		closePrice, _ := strconv.ParseFloat(k.Close, 64)
		volume, _ := strconv.ParseFloat(k.Volume, 64)

		result = append(result, KlineData{
			OpenTime:  k.OpenTime,
			Open:      openPrice,
			High:      highPrice,
			Low:       lowPrice,
			Close:     closePrice,
			Volume:    volume,
			CloseTime: k.CloseTime,
		})
	}

	return result, nil
}

// Get24hrTicker 获取24小时ticker数据
func (bs *BinanceService) Get24hrTicker(symbol string) (*TickerData, error) {
	if err := bs.checkRateLimit("ticker"); err != nil {
		return nil, err
	}

	client, err := bs.GetSpotClient()
	if err != nil {
		return nil, err
	}
	defer bs.clientPool.Put(client)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ticker, err := client.NewListPriceChangeStatsService().
		Symbol(symbol).
		Do(ctx)
	if err != nil {
		return nil, bs.handleBinanceError(err)
	}

	if len(ticker) == 0 {
		return nil, errors.New("no ticker data found")
	}

	t := ticker[0]
	lastPrice, _ := strconv.ParseFloat(t.LastPrice, 64)
	volume, _ := strconv.ParseFloat(t.Volume, 64)
	priceChangePercent, _ := strconv.ParseFloat(t.PriceChangePercent, 64)
	highPrice, _ := strconv.ParseFloat(t.HighPrice, 64)
	lowPrice, _ := strconv.ParseFloat(t.LowPrice, 64)

	return &TickerData{
		Symbol:             t.Symbol,
		LastPrice:          lastPrice,
		Volume:             volume,
		PriceChangePercent: priceChangePercent,
		HighPrice:          highPrice,
		LowPrice:           lowPrice,
	}, nil
}

// TickerData ticker数据结构
type TickerData struct {
	Symbol             string
	LastPrice          float64
	Volume             float64
	PriceChangePercent float64
	HighPrice          float64
	LowPrice           float64
}

// GetTradingSymbols 获取交易对列表
func (bs *BinanceService) GetTradingSymbols(ctx context.Context) ([]binance.Symbol, error) {
	// 检查缓存
	bs.symbolCache.mu.RLock()
	if time.Since(bs.symbolCache.lastUpdate) < bs.symbolCache.ttl && len(bs.symbolCache.symbols) > 0 {
		bs.symbolCache.mu.RUnlock()
		// 从缓存返回
		symbols := make([]binance.Symbol, 0, len(bs.symbolCache.symbols))
		for _, info := range bs.symbolCache.symbols {
			// 这里简化处理，实际应该存储完整的Symbol结构
			symbols = append(symbols, binance.Symbol{
				Symbol:     info.Symbol,
				BaseAsset:  info.BaseAsset,
				QuoteAsset: info.QuoteAsset,
			})
		}
		return symbols, nil
	}
	bs.symbolCache.mu.RUnlock()

	if err := bs.checkRateLimit("exchange_info"); err != nil {
		return nil, err
	}

	client, err := bs.GetSpotClient()
	if err != nil {
		return nil, err
	}
	defer bs.clientPool.Put(client)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	exchangeInfo, err := client.NewExchangeInfoService().Do(ctx)
	if err != nil {
		return nil, bs.handleBinanceError(err)
	}

	// 更新缓存
	bs.symbolCache.mu.Lock()
	bs.symbolCache.symbols = make(map[string]*SymbolInfo)
	for _, symbol := range exchangeInfo.Symbols {
		if symbol.Status == "TRADING" {
			info := &SymbolInfo{
				Symbol:              symbol.Symbol,
				BaseAsset:           symbol.BaseAsset,
				QuoteAsset:          symbol.QuoteAsset,
				BaseAssetPrecision:  symbol.BaseAssetPrecision,
				QuoteAssetPrecision: symbol.QuoteAssetPrecision,
			}

			// 解析过滤器
			for _, filter := range symbol.Filters {
				switch filter["filterType"] {
				case "LOT_SIZE":
					info.MinQty, _ = strconv.ParseFloat(filter["minQty"].(string), 64)
					info.MaxQty, _ = strconv.ParseFloat(filter["maxQty"].(string), 64)
					info.StepSize, _ = strconv.ParseFloat(filter["stepSize"].(string), 64)
				case "MIN_NOTIONAL":
					info.MinNotional, _ = strconv.ParseFloat(filter["minNotional"].(string), 64)
				case "PRICE_FILTER":
					// 可以添加价格精度等信息
				}
			}

			bs.symbolCache.symbols[symbol.Symbol] = info
		}
	}
	bs.symbolCache.lastUpdate = time.Now()
	bs.symbolCache.mu.Unlock()

	return exchangeInfo.Symbols, nil
}

// GetFuturesTradingSymbols 获取期货交易对列表
func (bs *BinanceService) GetFuturesTradingSymbols(ctx context.Context) ([]futures.Symbol, error) {
	if err := bs.checkRateLimit("futures_exchange_info"); err != nil {
		return nil, err
	}

	client, err := bs.GetFuturesClient()
	if err != nil {
		return nil, err
	}
	defer bs.futuresClientPool.Put(client)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	exchangeInfo, err := client.NewExchangeInfoService().Do(ctx)
	if err != nil {
		return nil, bs.handleBinanceError(err)
	}

	return exchangeInfo.Symbols, nil
}

// ValidateAPICredentials 验证API凭证
func (bs *BinanceService) ValidateAPICredentials(ctx context.Context) error {
	// 测试现货API
	_, err := bs.GetAccountInfo(ctx)
	if err != nil {
		return fmt.Errorf("spot API validation failed: %w", err)
	}

	// 测试期货API
	_, err = bs.GetFuturesAccountInfo(ctx)
	if err != nil {
		// 在测试网环境下，期货API可能不可用，只记录警告
		if bs.testNet {
			bs.logger.WithError(err).Warn("Futures API validation failed on testnet, this is expected")
		} else {
			return fmt.Errorf("futures API validation failed: %w", err)
		}
	}

	bs.logger.Info("API credentials validated successfully")
	return nil
}

// DiagnoseAPIConnection 诊断API连接问题
func (bs *BinanceService) DiagnoseAPIConnection(ctx context.Context) (map[string]interface{}, error) {
	diagnosis := make(map[string]interface{})

	// 1. 检查加密密钥
	diagnosis["encryption_key_length"] = len(config.AppConfig.Security.EncryptionKey)
	diagnosis["encryption_key_valid"] = len(config.AppConfig.Security.EncryptionKey) == 32

	// 2. 检查API密钥状态
	apiKey, secretKey, err := bs.decryptCredentials()
	if err != nil {
		diagnosis["credential_decrypt_error"] = err.Error()
		diagnosis["api_key_present"] = false
		diagnosis["secret_key_present"] = false
	} else {
		diagnosis["api_key_present"] = len(apiKey) > 0
		diagnosis["secret_key_present"] = len(secretKey) > 0
		diagnosis["api_key_length"] = len(apiKey)
		diagnosis["secret_key_length"] = len(secretKey)
		if len(apiKey) >= 6 {
			diagnosis["api_key_prefix"] = apiKey[:6] + "..."
		}
		// 验证API密钥格式（币安API密钥通常是64个字符）
		diagnosis["api_key_format_valid"] = len(apiKey) == 64
		diagnosis["secret_key_format_valid"] = len(secretKey) == 64
	}

	// 3. 检查网络配置
	diagnosis["testnet_enabled"] = bs.testNet
	client, err := bs.GetSpotClient()
	if err != nil {
		diagnosis["client_init_error"] = err.Error()
	} else if client != nil {
		diagnosis["base_url"] = client.BaseURL
		diagnosis["timeout"] = client.HTTPClient.Timeout.String()
		bs.clientPool.Put(client)
	}

	// 4. 尝试获取服务器时间（不需要认证）
	if client != nil {
		serverTime, err := client.NewServerTimeService().Do(ctx)
		if err != nil {
			diagnosis["server_time_error"] = err.Error()
			diagnosis["network_accessible"] = false
		} else {
			diagnosis["server_time"] = time.Unix(serverTime/1000, 0).Format(time.RFC3339)
			diagnosis["local_time"] = time.Now().Format(time.RFC3339)
			diagnosis["time_diff_ms"] = time.Now().UnixMilli() - serverTime
			diagnosis["network_accessible"] = true
		}
	}

	// 5. 如果基本网络可达，测试认证
	if diagnosis["network_accessible"] == true {
		// 创建一个新的客户端用于测试
		testClient := binance.NewClient(apiKey, secretKey)
		if bs.testNet {
			testClient.BaseURL = "https://testnet.binance.vision"
		}

		// 尝试不同的recvWindow值
		recvWindows := []int64{5000, 30000, config.AppConfig.Binance.RecvWindow}
		for _, window := range recvWindows {
			_, err := testClient.NewGetAccountService().Do(ctx, binance.WithRecvWindow(window))
			diagnosis[fmt.Sprintf("auth_test_recv_window_%d", window)] = err == nil
			if err != nil {
				diagnosis[fmt.Sprintf("auth_error_recv_window_%d", window)] = err.Error()
			} else {
				diagnosis["working_recv_window"] = window
				break
			}
		}
	}

	// 6. 记录诊断结果
	bs.logger.WithField("diagnosis", diagnosis).Info("API connection diagnosis completed")

	return diagnosis, nil
}

// getSymbolInfo 获取交易对信息
func (bs *BinanceService) getSymbolInfo(symbol string) (*SymbolInfo, error) {
	bs.symbolCache.mu.RLock()
	info, exists := bs.symbolCache.symbols[symbol]
	bs.symbolCache.mu.RUnlock()

	if !exists {
		// 更新缓存并重试
		ctx := context.Background()
		if _, err := bs.GetTradingSymbols(ctx); err != nil {
			return nil, err
		}

		bs.symbolCache.mu.RLock()
		info, exists = bs.symbolCache.symbols[symbol]
		bs.symbolCache.mu.RUnlock()

		if !exists {
			return nil, fmt.Errorf("symbol info not found: %s", symbol)
		}
	}

	return info, nil
}

// handleBinanceError 处理币安API错误
func (bs *BinanceService) handleBinanceError(err error) error {
	if err == nil {
		return nil
	}

	// 记录原始错误信息用于调试
	bs.logger.WithError(err).WithField("raw_error", err.Error()).Debug("Handling Binance API error")

	// 尝试解析为 common.APIError
	var apiErr *common.APIError
	if errors.As(err, &apiErr) {
		bs.logger.WithFields(logrus.Fields{
			"code":    apiErr.Code,
			"message": apiErr.Message,
		}).Error("Binance API error details")

		// 根据错误码返回更详细的错误信息
		switch apiErr.Code {
		case -1021:
			return fmt.Errorf("timestamp error: request timestamp outside recvWindow (server time diff too large)")
		case -1022:
			return fmt.Errorf("invalid signature: API secret key may be incorrect")
		case -2008:
			return fmt.Errorf("invalid API key: API key format is incorrect")
		case -2014:
			return fmt.Errorf("invalid API key: API key does not exist")
		case -2015:
			return fmt.Errorf("invalid API key: API key is not activated or IP not whitelisted")
		case -1000:
			return fmt.Errorf("unknown error: please check Binance API status")
		case -1003:
			return ErrRateLimitExceeded
		case -2010:
			return ErrInsufficientBalance
		case -2011:
			return ErrOrderNotFound
		default:
			return fmt.Errorf("binance API error: %s (code: %d)", apiErr.Message, apiErr.Code)
		}
	}

	// 解析币安错误码字符串
	errStr := err.Error()

	// 常见错误映射
	switch {
	case contains(errStr, "<APIError>"):
		// 处理空响应或格式异常的API错误
		if contains(errStr, "rsp=") && len(errStr) < 20 {
			return fmt.Errorf("币安API返回空响应，请检查：1) API密钥是否正确 2) API密钥是否已激活 3) IP是否在白名单中 4) 网络连接是否正常")
		}
		// 尝试从错误字符串中提取更多信息
		if contains(errStr, "code") {
			return fmt.Errorf("binance API error: %s (check API credentials and permissions)", errStr)
		}
		return fmt.Errorf("binance API error: %s", errStr)
	case contains(errStr, "-1021"):
		return errors.New("timestamp for this request is outside of the recvWindow")
	case contains(errStr, "-1022"):
		return errors.New("signature for this request is not valid")
	case contains(errStr, "-2010"):
		return ErrInsufficientBalance
	case contains(errStr, "-2011"):
		return ErrOrderNotFound
	case contains(errStr, "-1003"):
		return ErrRateLimitExceeded
	case contains(errStr, "-2015"):
		return ErrAPIPermissionDenied
	case contains(errStr, "-1100"):
		return errors.New("illegal characters found in parameter")
	case contains(errStr, "-1101"):
		return errors.New("too many parameters sent")
	case contains(errStr, "-1102"):
		return errors.New("mandatory parameter was not sent")
	case contains(errStr, "-1103"):
		return errors.New("unknown parameter was sent")
	case contains(errStr, "-1104"):
		return errors.New("not all sent parameters were read")
	case contains(errStr, "-1105"):
		return errors.New("parameter was empty")
	case contains(errStr, "-1106"):
		return errors.New("parameter was not required")
	case contains(errStr, "-1111"):
		return errors.New("precision is over the maximum")
	case contains(errStr, "-1112"):
		return errors.New("no orders on book")
	case contains(errStr, "-1121"):
		return errors.New("invalid symbol")
	case contains(errStr, "-1125"):
		return errors.New("this listenKey does not exist")
	case contains(errStr, "-2013"):
		return errors.New("order does not exist")
	case contains(errStr, "-2014"):
		return errors.New("API-key format invalid")
	case contains(errStr, "-2016"):
		return errors.New("no trading window could be found")
	default:
		return fmt.Errorf("binance API error: %v", err)
	}
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// GetOrderBook 获取订单簿
func (bs *BinanceService) GetOrderBook(ctx context.Context, symbol string, limit int) (*binance.DepthResponse, error) {
	if symbol == "" {
		return nil, ErrInvalidSymbol
	}

	if err := bs.checkRateLimit("depth"); err != nil {
		return nil, err
	}

	client, err := bs.GetSpotClient()
	if err != nil {
		return nil, err
	}
	defer bs.clientPool.Put(client)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	service := client.NewDepthService().Symbol(symbol)
	if limit > 0 {
		service = service.Limit(limit)
	}

	depth, err := service.Do(ctx)
	if err != nil {
		return nil, bs.handleBinanceError(err)
	}

	return depth, nil
}

// GetFuturesDepth 获取期货订单簿深度
func (bs *BinanceService) GetFuturesDepth(ctx context.Context, symbol string, limit int) (*futures.DepthResponse, error) {
	if symbol == "" {
		return nil, ErrInvalidSymbol
	}

	if err := bs.checkRateLimit("futures_depth"); err != nil {
		return nil, err
	}

	client, err := bs.GetFuturesClient()
	if err != nil {
		return nil, err
	}
	defer bs.futuresClientPool.Put(client)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	service := client.NewDepthService().Symbol(symbol)
	if limit > 0 {
		service = service.Limit(limit)
	}

	depth, err := service.Do(ctx)
	if err != nil {
		return nil, bs.handleBinanceError(err)
	}

	return depth, nil
}

// Close 关闭服务
func (bs *BinanceService) Close() error {
	bs.logger.Info("Closing Binance service")
	// 清理资源
	return nil
}

// GetTopSymbols 获取交易量最大的交易对
func (bs *BinanceService) GetTopSymbols(limit int) ([]string, error) {
	client, err := bs.GetSpotClient()
	if err != nil {
		return nil, err
	}
	defer bs.clientPool.Put(client)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 获取24小时行情数据
	tickers, err := client.NewListPriceChangeStatsService().Do(ctx)
	if err != nil {
		return nil, bs.handleBinanceError(err)
	}

	// 按交易量排序
	type symbolVolume struct {
		Symbol string
		Volume float64
	}

	var volumes []symbolVolume
	for _, ticker := range tickers {
		// 只选择USDT交易对
		if strings.HasSuffix(ticker.Symbol, "USDT") {
			volume, _ := strconv.ParseFloat(ticker.QuoteVolume, 64)
			volumes = append(volumes, symbolVolume{
				Symbol: ticker.Symbol,
				Volume: volume,
			})
		}
	}

	// 按交易量降序排序
	sort.Slice(volumes, func(i, j int) bool {
		return volumes[i].Volume > volumes[j].Volume
	})

	// 返回前N个交易对
	var symbols []string
	for i := 0; i < limit && i < len(volumes); i++ {
		symbols = append(symbols, volumes[i].Symbol)
	}

	return symbols, nil
}

// GetSymbol24hVolume 获取交易对24小时交易量
func (bs *BinanceService) GetSymbol24hVolume(symbol string) float64 {
	client, err := bs.GetSpotClient()
	if err != nil {
		return 0
	}
	defer bs.clientPool.Put(client)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stats, err := client.NewListPriceChangeStatsService().Symbol(symbol).Do(ctx)
	if err != nil || len(stats) == 0 {
		return 0
	}

	volume, _ := strconv.ParseFloat(stats[0].QuoteVolume, 64)
	return volume
}

// GetCurrentPrice 获取当前价格（decimal类型）
func (bs *BinanceService) GetCurrentPrice(symbol string) (decimal.Decimal, error) {
	price, err := bs.GetPrice(context.Background(), symbol)
	if err != nil {
		return decimal.Zero, err
	}
	return decimal.NewFromFloat(price), nil
}

// GetSymbolInfo 获取交易对信息
func (bs *BinanceService) GetSymbolInfo(symbol string) (*SymbolInfo, error) {
	info, err := bs.getSymbolInfo(symbol)
	if err != nil {
		return nil, err
	}

	return &SymbolInfo{
		Symbol:              info.Symbol,
		BaseAssetPrecision:  info.BaseAssetPrecision,
		QuoteAssetPrecision: info.QuoteAssetPrecision,
		MinQty:              info.MinQty,
		MaxQty:              info.MaxQty,
		StepSize:            info.StepSize,
		MinNotional:         info.MinNotional,
	}, nil
}

// GetSpotClientDirect 获取现货客户端（公开方法）
func (bs *BinanceService) GetSpotClientDirect() *binance.Client {
	client, _ := bs.GetSpotClient()
	return client
}

// GetFuturesClientDirect 获取期货客户端（公开方法）
func (bs *BinanceService) GetFuturesClientDirect() *futures.Client {
	client, _ := bs.GetFuturesClient()
	return client
}
