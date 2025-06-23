package H

import (
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

const (
	postLimit    = 40
	getLimit     = 50
	defaultLimit = 20
	interval     = time.Minute
)

type methodLimiter map[string]*tokenLimit
type rateLimiter struct {
	mu      sync.Mutex
	limits  map[string]*methodLimiter
	cleanup time.Duration
}

type tokenLimit struct {
	count   int
	resetAt time.Time
}

func NewRateLimiter(cleanup time.Duration) *rateLimiter {
	rl := &rateLimiter{
		limits:  make(map[string]*methodLimiter),
		cleanup: cleanup,
	}
	go rl.cleanupExpiredTokens()
	return rl
}

func (rl *rateLimiter) Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := normalizeToken(c.Request().Header.Get("Authorization"))
		if IsEmpty(token) {
			token = normalizeToken(c.Request().Header.Get("Token"))
		}
		if IsEmpty(token) {
			token = c.Request().URL.Query().Get("access_token")
		}
		if IsEmpty(token) {
			return next(c)
		}

		method := c.Request().Method
		var limit int
		switch method {
		case http.MethodPost:
			limit = postLimit
		case http.MethodGet:
			limit = getLimit
		default:
			limit = defaultLimit
		}

		allowed := rl.Allow(token, method, limit)
		remainingPost, _ := rl.GetRemaining(token, http.MethodPost, postLimit)
		remainingGet, _ := rl.GetRemaining(token, http.MethodGet, getLimit)
		remainingOther, _ := rl.GetRemaining(token, "OTHER", defaultLimit)

		c.Response().Header().Set("X-RateLimit-Post-Remaining", IntToString(remainingPost))
		c.Response().Header().Set("X-RateLimit-Get-Remaining", IntToString(remainingGet))
		c.Response().Header().Set("X-RateLimit-Other-Remaining", IntToString(remainingOther))

		if !allowed {
			return c.JSON(http.StatusTooManyRequests, map[string]string{
				"error": "Rate limit exceeded",
			})
		}

		return next(c)
	}
}

func (rl *rateLimiter) Allow(token, method string, limit int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	limiter, ok := rl.limits[token]
	if !ok || limiter == nil {
		rl.limits[token] = &methodLimiter{}
		limiter = rl.limits[token]
	}
	if _, ok := (*limiter)[method]; !ok {
		(*limiter)[method] = &tokenLimit{
			count:   0,
			resetAt: time.Now().Add(interval + time.Duration(rand.Intn(10))*time.Second),
		}
	}

	tokenLimit := (*limiter)[method]
	now := time.Now()

	if now.After(tokenLimit.resetAt) {
		tokenLimit.count = 0
		tokenLimit.resetAt = now.Add(interval + time.Duration(rand.Intn(10))*time.Second)
	}

	if tokenLimit.count >= limit {
		return false
	}

	tokenLimit.count++
	return true
}

func (rl *rateLimiter) GetRemaining(token, method string, limit int) (int, bool) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, ok := rl.limits[token]
	if !ok {
		return limit, true
	}
	if _, ok := (*limiter)[method]; !ok {
		return limit, true
	}

	tokenLimit := (*limiter)[method]
	now := time.Now()

	if now.After(tokenLimit.resetAt) {
		delete((*limiter), method)
		if len(*limiter) == 0 {
			delete(rl.limits, token)
		}
		return limit, true
	}

	return limit - tokenLimit.count, tokenLimit.count < limit
}

func (rl *rateLimiter) cleanupExpiredTokens() {
	for {
		time.Sleep(rl.cleanup)
		rl.mu.Lock()
		for token, methods := range rl.limits {
			for method, limit := range *methods {
				if time.Now().After(limit.resetAt) {
					delete(*methods, method)
				}
			}
			if len(*methods) == 0 {
				delete(rl.limits, token)
			}
		}
		rl.mu.Unlock()
	}
}

func normalizeToken(token string) string {
	token = strings.TrimSpace(token)
	if IsEmpty(token) {
		return token
	}

	parts := strings.Split(token, " ")
	lastPart := parts[len(parts)-1]

	jsonFromJWT, _, err := jwt.NewParser().ParseUnverified(lastPart, jwt.MapClaims{})
	if err != nil {
		return token
	}

	claims, ok := jsonFromJWT.Claims.(jwt.MapClaims)
	if !ok {
		return token
	}

	if uuid, ok := claims["uuid"].(string); ok && !IsEmpty(uuid) {
		return uuid
	}

	return token
}
