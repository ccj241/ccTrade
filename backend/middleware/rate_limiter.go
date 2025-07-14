package middleware

import (
	"github.com/ccj241/cctrade/config"
	"github.com/ccj241/cctrade/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"net/http"
	"sync"
	"time"
)

type IPRateLimiter struct {
	ips             map[string]*ipLimiterEntry
	mu              *sync.RWMutex
	r               rate.Limit
	b               int
	cleanupInterval time.Duration
	inactiveTimeout time.Duration
}

type ipLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	i := &IPRateLimiter{
		ips:             make(map[string]*ipLimiterEntry),
		mu:              &sync.RWMutex{},
		r:               r,
		b:               b,
		cleanupInterval: 1 * time.Hour,
		inactiveTimeout: 3 * time.Hour,
	}

	// 启动清理goroutine
	go i.cleanupRoutine()

	return i
}

func (i *IPRateLimiter) AddIP(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter := rate.NewLimiter(i.r, i.b)
	i.ips[ip] = &ipLimiterEntry{
		limiter:  limiter,
		lastSeen: time.Now(),
	}

	return limiter
}

func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.RLock()
	entry, exists := i.ips[ip]
	i.mu.RUnlock()

	if !exists {
		return i.AddIP(ip)
	}

	// 更新最后访问时间
	i.mu.Lock()
	entry.lastSeen = time.Now()
	i.mu.Unlock()

	return entry.limiter
}

var ipLimiter *IPRateLimiter

func init() {
	if config.AppConfig != nil {
		rateLimit := float64(config.AppConfig.Security.RateLimitRPM) / 60.0
		ipLimiter = NewIPRateLimiter(rate.Limit(rateLimit), config.AppConfig.Security.RateLimitRPM)
	} else {
		ipLimiter = NewIPRateLimiter(rate.Limit(1), 60)
	}
}

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		limiter := ipLimiter.GetLimiter(c.ClientIP())
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, utils.Response{
				Code:    429,
				Message: "请求频率过高，请稍后再试",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// cleanupRoutine 定期清理不活跃的IP限流器
func (i *IPRateLimiter) cleanupRoutine() {
	ticker := time.NewTicker(i.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		i.mu.Lock()
		for ip, entry := range i.ips {
			if now.Sub(entry.lastSeen) > i.inactiveTimeout {
				delete(i.ips, ip)
			}
		}
		i.mu.Unlock()
	}
}

type userLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type UserRateLimiter struct {
	users           map[uint]*userLimiterEntry
	mu              sync.RWMutex
	cleanupInterval time.Duration
	inactiveTimeout time.Duration
}

func NewUserRateLimiter() *UserRateLimiter {
	u := &UserRateLimiter{
		users:           make(map[uint]*userLimiterEntry),
		cleanupInterval: 1 * time.Hour,
		inactiveTimeout: 3 * time.Hour,
	}
	go u.cleanupRoutine()
	return u
}

func (u *UserRateLimiter) cleanupRoutine() {
	ticker := time.NewTicker(u.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		u.mu.Lock()
		for uid, entry := range u.users {
			if now.Sub(entry.lastSeen) > u.inactiveTimeout {
				delete(u.users, uid)
			}
		}
		u.mu.Unlock()
	}
}

func UserRateLimitMiddleware(requests int, window time.Duration) gin.HandlerFunc {
	userLimiter := NewUserRateLimiter()

	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		uid := userID.(uint)

		userLimiter.mu.RLock()
		entry, exists := userLimiter.users[uid]
		userLimiter.mu.RUnlock()

		var limiter *rate.Limiter
		if !exists {
			limiter = rate.NewLimiter(rate.Every(window/time.Duration(requests)), requests)
			userLimiter.mu.Lock()
			userLimiter.users[uid] = &userLimiterEntry{
				limiter:  limiter,
				lastSeen: time.Now(),
			}
			userLimiter.mu.Unlock()
		} else {
			limiter = entry.limiter
			// 更新最后访问时间
			userLimiter.mu.Lock()
			entry.lastSeen = time.Now()
			userLimiter.mu.Unlock()
		}

		if !limiter.Allow() {
			utils.ErrorResponse(c, 429, "用户请求频率过高")
			c.Abort()
			return
		}

		c.Next()
	}
}
