package router

import (
	"munch/server/internal/config"
	"munch/server/internal/handler"
	"munch/server/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Setup 装配所有路由。
func Setup(db *gorm.DB, cfg *config.Config, h *handler.Handler) *gin.Engine {
	r := gin.Default()

	// 本地开发跨域放开；线上走 callContainer 同源，不受影响。
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"Origin", "Content-Type", "Authorization", "X-WX-OPENID"},
	}))

	// 健康检查（微信云托管探活）
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	// 本地磁盘存储时，图片通过 /static 提供访问
	r.Static("/static", cfg.StaticDir)

	// 内置静态资源（霞鹜文楷子集字体等），供小程序 loadFontFace 拉取。
	// 上线后建议把字体传到 COS/CDN，前端 FONT_URL 改指过去，减轻服务压力。
	r.Static("/assets", "./assets")

	api := r.Group("/api")
	{
		// 登录无需鉴权
		api.POST("/login", h.Login)

		// 其余接口统一走鉴权中间件（X-WX-OPENID 或 JWT）
		auth := api.Group("")
		auth.Use(middleware.Auth(db, cfg))
		{
			auth.GET("/profile", h.Profile)

			// 情侣空间
			auth.POST("/couple", h.CreateCouple)
			auth.POST("/couple/join", h.JoinCouple)
			auth.GET("/couple", h.GetCouple)

			// 菜品
			auth.GET("/dishes", h.ListDishes)
			auth.POST("/dishes", h.CreateDish)
			auth.PATCH("/dishes/:id", h.UpdateDish)
			auth.DELETE("/dishes/:id", h.DeleteDish)

			// 分组
			auth.GET("/groups", h.ListGroups)
			auth.POST("/groups", h.CreateGroup)
			auth.PATCH("/groups/:id", h.UpdateGroup)
			auth.DELETE("/groups/:id", h.DeleteGroup)

			// 订单
			auth.GET("/orders", h.ListOrders)
			auth.POST("/orders", h.CreateOrder)
			auth.PATCH("/orders/:id/status", h.AdvanceOrder)

			// 买菜清单
			auth.GET("/shop-items", h.ListShopItems)
			auth.POST("/shop-items", h.CreateShopItem)
			auth.PATCH("/shop-items/:id", h.UpdateShopItem)
			auth.DELETE("/shop-items/:id", h.DeleteShopItem)

			// 图片上传
			auth.POST("/upload", h.Upload)             // multipart（H5/本地）
			auth.POST("/upload-base64", h.UploadBase64) // base64（云托管 callContainer）
		}
	}

	return r
}
