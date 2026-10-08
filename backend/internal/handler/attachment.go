package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"mybookkeeping/internal/middleware"
	"mybookkeeping/internal/service"
)

func (a *API) UploadAttachment(c *gin.Context) {
	if !a.minioReady || a.attachments == nil {
		Fail(c, http.StatusServiceUnavailable, "附件服务未就绪（请启动 MinIO）")
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, service.MaxAttachmentSize+(1<<20))
	file, err := c.FormFile("file")
	if err != nil {
		BadRequest(c, "请上传文件（上限 10MB）")
		return
	}
	f, err := file.Open()
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	defer f.Close()
	att, err := a.attachments.Upload(c.Request.Context(), middleware.GetUserID(c), file.Header.Get("Content-Type"), f, file.Size)
	if errors.Is(err, service.ErrAttachmentTooLarge) || errors.Is(err, service.ErrAttachmentType) {
		BadRequest(c, err.Error())
		return
	}
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	OK(c, att)
}

// ServeAttachment 凭签名短链读取附件（供 <img> 直接引用，不需要会话令牌）
func (a *API) ServeAttachment(c *gin.Context) {
	if a.attachments == nil {
		NotFound(c, service.ErrAttachmentNotFound.Error())
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	rc, size, ct, err := a.attachments.OpenSigned(c.Request.Context(), id, c.Query("exp"), c.Query("sig"))
	if err != nil {
		NotFound(c, service.ErrAttachmentNotFound.Error())
		return
	}
	defer rc.Close()
	c.Header("Cache-Control", "private, max-age=3600")
	c.Header("Content-Security-Policy", "default-src 'none'; sandbox")
	c.DataFromReader(http.StatusOK, size, ct, rc, nil)
}
