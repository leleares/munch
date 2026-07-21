package model

import (
	"time"

	"gorm.io/gorm"
)

// Base 内嵌到各模型，提供自增主键与时间戳。
type Base struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // 软删除
}

// Role 区分情侣空间里的两个角色。
type Role string

const (
	RoleCook    Role = "cook"    // 大厨端（做菜的一方）
	RoleOrderer Role = "orderer" // 点菜端（点菜的一方）
)

// User 微信用户。首次调用后端时按 openid 自动创建。
type User struct {
	Base
	OpenID   string `gorm:"uniqueIndex;size:64" json:"openid"`
	Nickname string `gorm:"size:64" json:"nickname"`
	Avatar   string `gorm:"size:512" json:"avatar"`
	Role     Role   `gorm:"size:16;default:orderer" json:"role"`
	CoupleID *uint  `gorm:"index" json:"coupleId"` // 未绑定情侣空间时为 nil
}

// Couple 情侣空间，是所有业务数据的归属单位。
type Couple struct {
	Base
	Name       string `gorm:"size:64" json:"name"`
	InviteCode string `gorm:"uniqueIndex;size:16" json:"inviteCode"` // 邀请另一半加入用
}

// Group 菜品分组。
type Group struct {
	Base
	CoupleID uint   `gorm:"index" json:"coupleId"`
	Name     string `gorm:"size:32" json:"name"`
	Sort     int    `json:"sort"`
}

// Dish 菜品。
type Dish struct {
	Base
	CoupleID  uint   `gorm:"index" json:"coupleId"`
	Name      string `gorm:"size:64" json:"name"`
	GroupID   uint   `gorm:"index" json:"groupId"`
	Spice     int    `json:"spice"` // 0 不辣 / 1 微辣 / 2 中辣 / 3 重辣
	Desc      string `gorm:"size:255" json:"desc"`
	ImageURL  string `gorm:"size:512" json:"imageUrl"`  // 上传照片时
	IconEmoji string `gorm:"size:16" json:"iconEmoji"`  // 选 emoji 图标时
	IsFav     bool   `gorm:"default:false" json:"isFav"` // 是否「常点」
	CreatedBy uint   `json:"createdBy"`
}

// Order 一次下单。
type Order struct {
	Base
	CoupleID  uint        `gorm:"index" json:"coupleId"`
	CreatedBy uint        `json:"createdBy"`
	Status    OrderStatus `gorm:"size:16;default:pending" json:"status"`
	Message   string      `gorm:"size:255" json:"message"`      // 给大厨的留言
	Items     []OrderItem `gorm:"foreignKey:OrderID" json:"items"`
}

// OrderStatus 订单状态机。
type OrderStatus string

const (
	StatusPending OrderStatus = "pending" // 待接单
	StatusCooking OrderStatus = "cooking" // 备菜中
	StatusServed  OrderStatus = "served"  // 已上菜
)

// Next 返回状态机的下一个状态；到终态返回空串。
func (s OrderStatus) Next() OrderStatus {
	switch s {
	case StatusPending:
		return StatusCooking
	case StatusCooking:
		return StatusServed
	default:
		return ""
	}
}

// OrderItem 下单快照：存名称/备注，避免菜品后续被改/删影响历史。
type OrderItem struct {
	Base
	OrderID    uint   `gorm:"index" json:"orderId"`
	DishID     uint   `json:"dishId"`
	Name       string `gorm:"size:64" json:"name"`
	Qty        int    `json:"qty"`
	SpiceLabel string `gorm:"size:32" json:"spiceLabel"` // 如「微辣 🌶」
	Forbid     string `gorm:"size:255" json:"forbid"`    // 忌口/悄悄话
}

// ShopItem 买菜清单条目。
type ShopItem struct {
	Base
	CoupleID uint   `gorm:"index" json:"coupleId"`
	Text     string `gorm:"size:128" json:"text"`
	Done     bool   `gorm:"default:false" json:"done"`
}

// AllModels 供 AutoMigrate 一次性建表。
func AllModels() []interface{} {
	return []interface{}{
		&User{}, &Couple{}, &Group{}, &Dish{}, &Order{}, &OrderItem{}, &ShopItem{},
	}
}
