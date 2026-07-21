package handler

import (
	"strconv"

	"munch/server/internal/model"
	"munch/server/pkg/response"

	"github.com/gin-gonic/gin"
)

// ListDishes 列出菜品，可按 groupId、fav 过滤（软删除自动排除）。
func (h *Handler) ListDishes(c *gin.Context) {
	cid, ok := h.coupleID(c)
	if !ok {
		return
	}
	q := h.DB.Where("couple_id = ?", cid)
	if gid := c.Query("groupId"); gid != "" {
		q = q.Where("group_id = ?", gid)
	}
	if c.Query("fav") == "1" || c.Query("fav") == "true" {
		q = q.Where("is_fav = ?", true)
	}
	var dishes []model.Dish
	q.Order("id desc").Find(&dishes)
	response.OK(c, dishes)
}

// DishReq 新增/编辑菜品入参。
type DishReq struct {
	Name      string `json:"name"`
	GroupID   uint   `json:"groupId"`
	Spice     int    `json:"spice"`
	Desc      string `json:"desc"`
	ImageURL  string `json:"imageUrl"`
	IconEmoji string `json:"iconEmoji"`
	IsFav     *bool  `json:"isFav"`
}

// CreateDish 新增菜品。
func (h *Handler) CreateDish(c *gin.Context) {
	cid, ok := h.coupleID(c)
	if !ok {
		return
	}
	var req DishReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		response.Fail(c, response.CodeParamError, "菜名不能为空")
		return
	}
	dish := model.Dish{
		CoupleID: cid, Name: req.Name, GroupID: req.GroupID, Spice: req.Spice,
		Desc: req.Desc, ImageURL: req.ImageURL, IconEmoji: req.IconEmoji,
		CreatedBy: h.user(c).ID,
	}
	if req.IsFav != nil {
		dish.IsFav = *req.IsFav
	}
	if err := h.DB.Create(&dish).Error; err != nil {
		response.Fail(c, response.CodeServer, "新增失败")
		return
	}
	response.OK(c, dish)
}

// UpdateDish 编辑菜品（仅更新传入字段）。
func (h *Handler) UpdateDish(c *gin.Context) {
	cid, ok := h.coupleID(c)
	if !ok {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	var dish model.Dish
	if err := h.DB.Where("id = ? AND couple_id = ?", id, cid).First(&dish).Error; err != nil {
		response.Fail(c, response.CodeNotFound, "菜品不存在")
		return
	}
	var req DishReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "参数错误")
		return
	}
	updates := map[string]interface{}{
		"name": req.Name, "group_id": req.GroupID, "spice": req.Spice,
		"desc": req.Desc, "image_url": req.ImageURL, "icon_emoji": req.IconEmoji,
	}
	if req.IsFav != nil {
		updates["is_fav"] = *req.IsFav
	}
	h.DB.Model(&dish).Updates(updates)
	h.DB.First(&dish, dish.ID)
	response.OK(c, dish)
}

// DeleteDish 软删除菜品。
func (h *Handler) DeleteDish(c *gin.Context) {
	cid, ok := h.coupleID(c)
	if !ok {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.DB.Where("id = ? AND couple_id = ?", id, cid).Delete(&model.Dish{}).Error; err != nil {
		response.Fail(c, response.CodeServer, "删除失败")
		return
	}
	response.OK(c, gin.H{"id": id})
}
