package middleware

import (
	"finance-dashboard/config"
	"finance-dashboard/store"
	"finance-dashboard/utils"
	"log"
	"strings"
	"sync"
	"time"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	ContextUserID = "userId"
	ContextRole   = "role"
	ContextJTI    = "jti"
)

func AuthMiddleware(userStore store.UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.SendError(c, utils.NewUnauthorizedError("authorization token required"))
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			utils.SendError(c, utils.NewUnauthorizedError("invalid authorization format, use Bearer token"))
			c.Abort()
			return
		}

		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(config.App.JWTSecret), nil
		})

		if err != nil {
			switch {
			case errors.Is(err, jwt.ErrTokenExpired):
				utils.SendError(c, utils.NewUnauthorizedError("token has expired"))
			case errors.Is(err, jwt.ErrTokenSignatureInvalid):
				utils.SendError(c, utils.NewUnauthorizedError("invalid token"))
			case errors.Is(err, jwt.ErrTokenMalformed):
				utils.SendError(c, utils.NewUnauthorizedError("malformed token"))
			default:
				utils.SendError(c, utils.NewUnauthorizedError("invalid token"))
			}
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			utils.SendError(c, utils.NewUnauthorizedError("invalid token claims"))
			c.Abort()
			return
		}

		userID, ok := claims[ContextUserID].(string)
		if !ok || userID == "" {
			utils.SendError(c, utils.NewUnauthorizedError("invalid token payload"))
			c.Abort()
			return
		}

		user := userStore.GetByID(userID)
		if user == nil {
			utils.SendError(c, utils.NewUnauthorizedError("user not found"))
			c.Abort()
			return
		}

		if !user.IsActive {
			utils.SendError(c, utils.NewUnauthorizedError("account has been deactivated"))
			c.Abort()
			return
		}

		jti, _ := claims[ContextJTI].(string)

		c.Set(ContextUserID, userID)
		c.Set(ContextRole, string(user.Role))
		c.Set(ContextJTI, jti)

		c.Next()
	}
}

func GetUserID(c *gin.Context) string {
	id, _ := c.Get(ContextUserID)
	str, _ := id.(string)
	return str
}

func GetRole(c *gin.Context) string {
	role, _ := c.Get(ContextRole)
	str, _ := role.(string)
	return str
}

// rateLimitEntry tracks request count and window start for one IP
type rateLimitEntry struct {
	count       int
	windowStart time.Time
}

type RateLimiter struct {
	mu      sync.Mutex // plain Mutex — reads always accompany writes here
	entries map[string]*rateLimitEntry
	limit   int           // max requests per window
	window  time.Duration // window size
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		entries: make(map[string]*rateLimitEntry),
		limit:   limit,
		window:  window,
	}
	go rl.cleanupLoop()
	return rl
}
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.cleanup()
	}
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	for ip, entry := range rl.entries {
		if now.Sub(entry.windowStart) >= rl.window {
			delete(rl.entries, ip)
		}
	}
}

// RateLimit returns gin middleware using this limiter
func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {

		ip := c.ClientIP()

		rl.mu.Lock()

		entry, exists := rl.entries[ip]
		now := time.Now()

		if !exists || now.Sub(entry.windowStart) >= rl.window {
			// New IP or window expired — reset
			rl.entries[ip] = &rateLimitEntry{
				count:       1,
				windowStart: now,
			}
			rl.mu.Unlock()
			c.Next()
			return
		}

		entry.count++

		if entry.count > rl.limit {
			rl.mu.Unlock()

			c.Header("Retry-After", "60")
			utils.SendError(c, &utils.AppError{
				Code:       "RATE_LIMIT_EXCEEDED",
				Message:    "Too many requests",
				Details:    "rate limit exceeded, please retry after 60 seconds",
				StatusCode: 429,
			})
			c.Abort()
			log.Printf("rate limit hit: ip=%s count=%d", ip, entry.count)
			return
		}

		rl.mu.Unlock()
		c.Next()
	}
}
