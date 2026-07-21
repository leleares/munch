package middleware

import (
	"strings"

	"munch/server/internal/config"
	"munch/server/internal/model"
	"munch/server/internal/service"
	"munch/server/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const ctxUserKey = "currentUser"

// Auth 是鉴权中间件工厂。识别调用者身份的两条路径：
//  1. 微信云托管 callContainer：平台校验后在请求头注入 X-WX-OPENID，可信，直接取用（推荐，免 code2session）。
//  2. 自建/本地：读取 Authorization: Bearer <jwt>，解析出 userID。
//
// 命中任一路径后，按 openid/uid 找到（或惰性创建）用户并挂到 gin.Context。
func Auth(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user model.User

		// 路径 1：微信云托管注入的 openid
		if openid := c.GetHeader("X-WX-OPENID"); openid != "" {
			if err := db.Where(model.User{OpenID: openid}).
				Attrs(model.User{Role: model.RoleOrderer}).
				FirstOrCreate(&user).Error; err != nil {
				response.Unauthorized(c, "用户初始化失败")
				c.Abort()
				return
			}
			c.Set(ctxUserKey, &user)
			c.Next()
			return
		}

		// 路径 2：自签 JWT
		token := bearer(c.GetHeader("Authorization"))
		if token == "" {
			response.Unauthorized(c, "未登录")
			c.Abort()
			return
		}
		claims, err := service.ParseToken(cfg.JWTSecret, token)
		if err != nil {
			response.Unauthorized(c, "登录已过期，请重新登录")
			c.Abort()
			return
		}
		if err := db.First(&user, claims.UserID).Error; err != nil {
			response.Unauthorized(c, "用户不存在")
			c.Abort()
			return
		}
		c.Set(ctxUserKey, &user)
		c.Next()
	}
}

// CurrentUser 从 gin.Context 取出当前登录用户。
func CurrentUser(c *gin.Context) *model.User {
	v, ok := c.Get(ctxUserKey)
	if !ok {
		return nil
	}
	u, _ := v.(*model.User)
	return u
}

func bearer(h string) string {
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return h
}
