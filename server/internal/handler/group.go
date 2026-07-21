package handler

import (
	"strconv"

	"munch/server/internal/model"
	"munch/server/pkg/response"

	"github.com/gin-gonic/gin"
)

// ListGroups 列出情侣空间的所有分组。
func (h *Handler) ListGroups(c *gin.Context) {
	cid, ok := h.coupleID(c)
	if !ok {
		return
	}
	var groups []model.Group
	h.DB.Where("couple_id = ?", cid).Order("sort asc, id asc").Find(&groups)
	response.OK(c, groups)
}

// GroupReq 分组增改入参。
type GroupReq struct {
	Name string `json:"name"`
}

// CreateGroup 新建分组。
func (h *Handler) CreateGroup(c *gin.Context) {
	cid, ok := h.coupleID(c)
	if !ok {
		return
	}
	var req GroupReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		response.Fail(c, response.CodeParamError, "分组名不能为空")
		return
	}
	var count int64
	h.DB.Model(&model.Group{}).Where("couple_id = ?", cid).Count(&count)
	g := model.Group{CoupleID: cid, Name: req.Name, Sort: int(count)}
	if err := h.DB.Create(&g).Error; err != nil {
		response.Fail(c, response.CodeServer, "创建失败")
		return
	}
	response.OK(c, g)
}

// UpdateGroup 改分组名。
func (h *Handler) UpdateGroup(c *gin.Context) {
	cid, ok := h.coupleID(c)
	if !ok {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	var req GroupReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		response.Fail(c, response.CodeParamError, "分组名不能为空")
		return
	}
	var g model.Group
	if err := h.DB.Where("id = ? AND couple_id = ?", id, cid).First(&g).Error; err != nil {
		response.Fail(c, response.CodeNotFound, "分组不存在")
		return
	}
	h.DB.Model(&g).Update("name", req.Name)
	response.OK(c, g)
}

// DeleteGroup 删除分组。该组菜品迁到剩余的第一个分组；禁止删到 0 组。
func (h *Handler) DeleteGroup(c *gin.Context) {
	cid, ok := h.coupleID(c)
	if !ok {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))

	var groups []model.Group
	h.DB.Where("couple_id = ?", cid).Order("sort asc, id asc").Find(&groups)
	if len(groups) <= 1 {
		response.Fail(c, response.CodeConflict, "至少保留一个分组")
		return
	}

	// 找 fallback：删除目标之外的第一个分组
	var fallback *model.Group
	for i := range groups {
		if groups[i].ID != uint(id) {
			fallback = &groups[i]
			break
		}
	}
	if fallback == nil {
		response.Fail(c, response.CodeConflict, "至少保留一个分组")
		return
	}

	tx := h.DB.Begin()
	tx.Model(&model.Dish{}).Where("couple_id = ? AND group_id = ?", cid, id).Update("group_id", fallback.ID)
	tx.Where("id = ? AND couple_id = ?", id, cid).Delete(&model.Group{})
	tx.Commit()

	response.OK(c, gin.H{"id": id, "fallbackGroupId": fallback.ID})
}
