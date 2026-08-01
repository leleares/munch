package handler

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"munch/server/pkg/response"

	"github.com/gin-gonic/gin"
)

// Upload 接收 multipart 文件字段 file，存入存储层，返回 imageUrl。
// 微信云托管里建议把 STORAGE_DRIVER 切到 cos；本地默认落磁盘并通过 /static 提供访问。
func (h *Handler) Upload(c *gin.Context) {
	if _, ok := h.coupleID(c); !ok {
		return
	}
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, response.CodeParamError, "缺少文件字段 file")
		return
	}
	f, err := fileHeader.Open()
	if err != nil {
		response.Fail(c, response.CodeServer, "读取文件失败")
		return
	}
	defer f.Close()

	// 用时间戳 + 用户 ID 生成文件名，避免碰撞（Save 由存储层落盘/上云）
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	name := fmt.Sprintf("%d_%d%s", h.user(c).ID, time.Now().UnixNano(), ext)

	url, err := h.Storage.Save(name, f)
	if err != nil {
		response.Fail(c, response.CodeServer, "保存失败："+err.Error())
		return
	}
	response.OK(c, gin.H{"imageUrl": url})
}

// UploadBase64Req base64 上传入参。
// 微信云托管走 callContainer（JSON 通道，不支持 multipart），故图片以 base64 传。
type UploadBase64Req struct {
	Ext  string `json:"ext"`  // 扩展名，如 ".jpg"，可空
	Data string `json:"data"` // base64，可带 data:image/xxx;base64, 前缀
}

// UploadBase64 接收 base64 图片，存入存储层，返回 imageUrl。
func (h *Handler) UploadBase64(c *gin.Context) {
	if _, ok := h.coupleID(c); !ok {
		return
	}
	var req UploadBase64Req
	if err := c.ShouldBindJSON(&req); err != nil || req.Data == "" {
		response.Fail(c, response.CodeParamError, "缺少图片数据")
		return
	}

	// 去掉 data URL 前缀，并按前缀推断扩展名（前端没传 ext 时）
	raw := req.Data
	ext := strings.ToLower(strings.TrimSpace(req.Ext))
	if i := strings.Index(raw, ","); strings.HasPrefix(raw, "data:") && i > 0 {
		if ext == "" {
			if strings.Contains(raw[:i], "png") {
				ext = ".png"
			} else if strings.Contains(raw[:i], "webp") {
				ext = ".webp"
			} else if strings.Contains(raw[:i], "gif") {
				ext = ".gif"
			}
		}
		raw = raw[i+1:]
	}
	if ext == "" {
		ext = ".jpg"
	}

	data, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		response.Fail(c, response.CodeParamError, "图片解码失败")
		return
	}

	name := fmt.Sprintf("%d_%d%s", h.user(c).ID, time.Now().UnixNano(), ext)
	url, err := h.Storage.Save(name, bytes.NewReader(data))
	if err != nil {
		response.Fail(c, response.CodeServer, "保存失败："+err.Error())
		return
	}
	response.OK(c, gin.H{"imageUrl": url})
}
