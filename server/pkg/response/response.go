package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Body 是所有接口统一的返回结构：code=0 表示成功，非 0 为业务错误码。
type Body struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// OK 返回成功数据。
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Body{Code: 0, Msg: "ok", Data: data})
}

// Fail 返回业务错误（HTTP 仍为 200，靠 code 区分，前端处理更统一）。
func Fail(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Body{Code: code, Msg: msg})
}

// Unauthorized 返回 401，用于鉴权失败。
func Unauthorized(c *gin.Context, msg string) {
	c.JSON(http.StatusUnauthorized, Body{Code: 401, Msg: msg})
}

// 常见业务错误码
const (
	CodeParamError = 1001 // 参数错误
	CodeNoCouple   = 1002 // 尚未加入情侣空间
	CodeNotFound   = 1003 // 资源不存在
	CodeForbidden  = 1004 // 无权限
	CodeConflict   = 1005 // 状态冲突（如订单状态机顺序不对）
	CodeServer     = 1500 // 服务端错误
)
