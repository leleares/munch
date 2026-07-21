package handler

import (
	"strconv"

	"munch/server/internal/model"
	"munch/server/pkg/response"

	"github.com/gin-gonic/gin"
)

// ListShopItems 买菜清单列表。
func (h *Handler) ListShopItems(c *gin.Context) {
	cid, ok := h.coupleID(c)
	if !ok {
		return
	}
	var items []model.ShopItem
	h.DB.Where("couple_id = ?", cid).Order("done asc, id desc").Find(&items)
	response.OK(c, items)
}

// ShopItemReq 买菜清单入参。
type ShopItemReq struct {
	Text string `json:"text"`
	Done *bool  `json:"done"`
}

// CreateShopItem 新增买菜条目。
func (h *Handler) CreateShopItem(c *gin.Context) {
	cid, ok := h.coupleID(c)
	if !ok {
		return
	}
	var req ShopItemReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Text == "" {
		response.Fail(c, response.CodeParamError, "写点要买的吧")
		return
	}
	item := model.ShopItem{CoupleID: cid, Text: req.Text}
	if err := h.DB.Create(&item).Error; err != nil {
		response.Fail(c, response.CodeServer, "新增失败")
		return
	}
	response.OK(c, item)
}

// UpdateShopItem 更新买菜条目（勾选/文案）。
func (h *Handler) UpdateShopItem(c *gin.Context) {
	cid, ok := h.coupleID(c)
	if !ok {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	var item model.ShopItem
	if err := h.DB.Where("id = ? AND couple_id = ?", id, cid).First(&item).Error; err != nil {
		response.Fail(c, response.CodeNotFound, "条目不存在")
		return
	}
	var req ShopItemReq
	_ = c.ShouldBindJSON(&req)
	updates := map[string]interface{}{}
	if req.Text != "" {
		updates["text"] = req.Text
	}
	if req.Done != nil {
		updates["done"] = *req.Done
	}
	h.DB.Model(&item).Updates(updates)
	h.DB.First(&item, item.ID)
	response.OK(c, item)
}

// DeleteShopItem 删除买菜条目。
func (h *Handler) DeleteShopItem(c *gin.Context) {
	cid, ok := h.coupleID(c)
	if !ok {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	h.DB.Where("id = ? AND couple_id = ?", id, cid).Delete(&model.ShopItem{})
	response.OK(c, gin.H{"id": id})
}
