package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"mybookkeeping/internal/middleware"
)

func (a *API) RecognizeText(c *gin.Context) {
	if a.llm == nil || !a.llm.Enabled() {
		Fail(c, http.StatusServiceUnavailable, "未配置 DeepSeek API Key")
		return
	}
	var req struct {
		Text string `json:"text"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	res, err := a.llm.RecognizeText(c.Request.Context(), middleware.GetUserID(c), req.Text)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, res)
}

func (a *API) RecognizeTextBatch(c *gin.Context) {
	if a.llm == nil || !a.llm.Enabled() {
		Fail(c, http.StatusServiceUnavailable, "未配置 DeepSeek API Key")
		return
	}
	var req struct {
		Text string `json:"text"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	list, err := a.llm.RecognizeTextBatch(c.Request.Context(), middleware.GetUserID(c), req.Text)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, gin.H{"items": list})
}

func (a *API) RecognizeImage(c *gin.Context) {
	if a.llm == nil || !a.llm.Enabled() {
		Fail(c, http.StatusServiceUnavailable, "未配置 DeepSeek API Key")
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 13<<20)
	file, err := c.FormFile("file")
	if err != nil {
		BadRequest(c, "请上传图片（字段名 file，上限约 12MB）")
		return
	}
	f, err := file.Open()
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, 12<<20+1))
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	if len(data) > 12<<20 {
		BadRequest(c, "图片过大（上限约 12MB）")
		return
	}
	ct := file.Header.Get("Content-Type")
	hint := c.PostForm("hint")
	res, err := a.llm.RecognizeImage(c.Request.Context(), middleware.GetUserID(c), data, ct, hint)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, res)
}
