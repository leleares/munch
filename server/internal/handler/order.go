package handler

import (
	"strconv"

	"munch/server/internal/model"
	"munch/server/pkg/response"

	"github.com/gin-gonic/gin"
)

// OrderItemReq 下单单项入参。
type OrderItemReq struct {
	DishID     uint   `json:"dishId"`
	Qty        int    `json:"qty"`
	SpiceLabel string `json:"spiceLabel"`
	Forbid     string `json:"forbid"`
}

// CreateOrderReq 下单入参。
type CreateOrderReq struct {
	Items   []OrderItemReq `json:"items"`
	Message string         `json:"message"`
}

// CreateOrder 下单：把当前购物车固化为一张订单，item 存名称快照。
func (h *Handler) CreateOrder(c *gin.Context) {
	cid, ok := h.coupleID(c)
	if !ok {
		return
	}
	var req CreateOrderReq
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Items) == 0 {
		response.Fail(c, response.CodeParamError, "先点一道菜嘛")
		return
	}

	order := model.Order{
		CoupleID:  cid,
		CreatedBy: h.user(c).ID,
		Status:    model.StatusPending,
		Message:   req.Message,
	}
	// 为每个 item 补一份菜名快照
	for _, it := range req.Items {
		if it.Qty <= 0 {
			continue
		}
		name := ""
		var dish model.Dish
		if err := h.DB.Where("id = ? AND couple_id = ?", it.DishID, cid).First(&dish).Error; err == nil {
			name = dish.Name
		}
		order.Items = append(order.Items, model.OrderItem{
			DishID: it.DishID, Name: name, Qty: it.Qty,
			SpiceLabel: it.SpiceLabel, Forbid: it.Forbid,
		})
	}
	if len(order.Items) == 0 {
		response.Fail(c, response.CodeParamError, "先点一道菜嘛")
		return
	}

	if err := h.DB.Create(&order).Error; err != nil {
		response.Fail(c, response.CodeServer, "下单失败")
		return
	}
	response.OK(c, order)
}

// ListOrders 订单列表，按时间倒序，预加载 items。
func (h *Handler) ListOrders(c *gin.Context) {
	cid, ok := h.coupleID(c)
	if !ok {
		return
	}
	var orders []model.Order
	h.DB.Preload("Items").Where("couple_id = ?", cid).Order("id desc").Find(&orders)
	response.OK(c, orders)
}

// AdvanceStatusReq 推进订单状态入参（可选，传了就校验目标是否等于自然下一态）。
type AdvanceStatusReq struct {
	Status model.OrderStatus `json:"status"`
}

// AdvanceOrder 大厨推进订单状态：待接单 → 备菜中 → 已上菜（服务端强制状态机顺序）。
func (h *Handler) AdvanceOrder(c *gin.Context) {
	cid, ok := h.coupleID(c)
	if !ok {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	var order model.Order
	if err := h.DB.Where("id = ? AND couple_id = ?", id, cid).First(&order).Error; err != nil {
		response.Fail(c, response.CodeNotFound, "订单不存在")
		return
	}
	next := order.Status.Next()
	if next == "" {
		response.Fail(c, response.CodeConflict, "订单已完成，无需推进")
		return
	}
	// 若前端指定了目标状态，必须与自然下一态一致，防止跳态
	var req AdvanceStatusReq
	if err := c.ShouldBindJSON(&req); err == nil && req.Status != "" && req.Status != next {
		response.Fail(c, response.CodeConflict, "状态流转顺序不对")
		return
	}
	h.DB.Model(&order).Update("status", next)
	order.Status = next
	response.OK(c, order)
}
