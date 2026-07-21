package handler

import (
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
