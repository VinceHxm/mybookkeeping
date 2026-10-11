package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"mybookkeeping/internal/i18n"
)

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"ok": true, "data": data})
}

// Fail msg 一律传中文原文，这里按请求语言翻译
func Fail(c *gin.Context, code int, msg string) {
	c.JSON(code, gin.H{"ok": false, "error": i18n.T(i18n.FromRequest(c), msg)})
}

// FailCode 附带稳定错误码，供前端按码分支（文案会随语言变化，不能拿来判断）
func FailCode(c *gin.Context, code int, errCode, msg string) {
	c.JSON(code, gin.H{"ok": false, "code": errCode, "error": i18n.T(i18n.FromRequest(c), msg)})
}

func BadRequest(c *gin.Context, msg string) {
	Fail(c, http.StatusBadRequest, msg)
}

func NotFound(c *gin.Context, msg string) {
	Fail(c, http.StatusNotFound, msg)
}

// ServerError 内部错误细节（SQL、Redis、MinIO 报错等）只写日志，不回给客户端。
func ServerError(c *gin.Context, msg string) {
	log.Printf("server error %s %s: %s", c.Request.Method, c.FullPath(), msg)
	Fail(c, http.StatusInternalServerError, "服务器内部错误，请稍后重试")
}
