package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"mybookkeeping/internal/model"
)

const ContextUserID = "userId"

type Auth struct {
	rdb *redis.Client
}

func NewAuth(rdb *redis.Client) *Auth {
	return &Auth{rdb: rdb}
}

func (a *Auth) Required() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}
		ctx := context.Background()
		uid, err := a.rdb.Get(ctx, "sess:"+token).Result()
		if err == redis.Nil || uid == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "登录已过期"})
			return
		}
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "会话校验失败"})
			return
		}
		c.Set(ContextUserID, uid)
		c.Set("sessionToken", token)
		c.Next()
	}
}

// RequireAdmin 须在 Required() 之后；getUser 按 userId 取用户
func RequireAdmin(getUser func(uint64) (*model.User, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := GetUserID(c)
		if uid == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}
		u, err := getUser(uid)
		if err != nil || u == nil || u.Disabled {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "登录已失效"})
			return
		}
		if !u.IsAdmin() {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
			return
		}
		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	if strings.HasPrefix(strings.ToLower(h), "bearer ") {
		return strings.TrimSpace(h[7:])
	}
	return c.Query("token")
}

func GetUserID(c *gin.Context) uint64 {
	v, _ := c.Get(ContextUserID)
	switch t := v.(type) {
	case string:
		var id uint64
		for _, ch := range t {
			if ch < '0' || ch > '9' {
				return 0
			}
			id = id*10 + uint64(ch-'0')
		}
		return id
	case uint64:
		return t
	default:
		return 0
	}
}
