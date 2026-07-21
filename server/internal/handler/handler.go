package handler

import (
	"munch/server/internal/config"
	"munch/server/internal/middleware"
	"munch/server/internal/model"
	"munch/server/internal/storage"
	"munch/server/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler 持有所有 handler 依赖，方法即路由处理函数。
type Handler struct {
	DB      *gorm.DB
	Cfg     *config.Config
	Storage storage.Storage
}

func New(db *gorm.DB, cfg *config.Config, st storage.Storage) *Handler {
	return &Handler{DB: db, Cfg: cfg, Storage: st}
}

// coupleID 取当前用户的情侣空间 ID；未绑定则写错误响应并返回 ok=false。
// 除登录与情侣空间管理接口外，其余业务接口都以它做数据隔离。
func (h *Handler) coupleID(c *gin.Context) (uint, bool) {
	u := middleware.CurrentUser(c)
	if u == nil {
		response.Unauthorized(c, "未登录")
		return 0, false
	}
	if u.CoupleID == nil {
		response.Fail(c, response.CodeNoCouple, "还没有情侣空间，先创建或加入一个吧")
		return 0, false
	}
	return *u.CoupleID, true
}

func (h *Handler) user(c *gin.Context) *model.User {
	return middleware.CurrentUser(c)
}
