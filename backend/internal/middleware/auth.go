package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"mybookkeeping/internal/i18n"
	"mybookkeeping/internal/kv"
	"mybookkeeping/internal/model"
)

const ContextUserID = "userId"

type Auth struct {
	kv kv.Store
}

func NewAuth(store kv.Store) *Auth {
	return &Auth{kv: store}
}

func (a *Auth) Required() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			abort(c, http.StatusUnauthorized, "未登录")
			return
		}
		ctx := context.Background()
		uid, ok, err := a.kv.Get(ctx, "sess:"+token)
		if err != nil {
			abort(c, http.StatusInternalServerError, "会话校验失败")
			return
		}
		if !ok || uid == "" {
			abort(c, http.StatusUnauthorized, "登录已过期")
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
			abort(c, http.StatusUnauthorized, "未登录")
			return
		}
		u, err := getUser(uid)
		if err != nil || u == nil || u.Disabled {
			abort(c, http.StatusUnauthorized, "登录已失效")
			return
		}
		if !u.IsAdmin() {
			abort(c, http.StatusForbidden, "需要管理员权限")
			return
		}
		c.Next()
	}
}

func abort(c *gin.Context, status int, msg string) {
	c.AbortWithStatusJSON(status, gin.H{"error": i18n.T(i18n.FromRequest(c), msg)})
}

// extractToken 只接受 Authorization 头；不支持 URL 参数，避免令牌写进访问日志 / Referer。
func extractToken(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	if strings.HasPrefix(strings.ToLower(h), "bearer ") {
		return strings.TrimSpace(h[7:])
	}
	return ""
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
