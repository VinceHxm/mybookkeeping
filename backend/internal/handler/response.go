package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"ok": true, "data": data})
}

func Fail(c *gin.Context, code int, msg string) {
	c.JSON(code, gin.H{"ok": false, "error": msg})
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
