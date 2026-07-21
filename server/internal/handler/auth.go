package handler

import (
	"math/rand"
	"strings"

	"munch/server/internal/middleware"
	"munch/server/internal/model"
	"munch/server/internal/service"
	"munch/server/pkg/response"

	"github.com/gin-gonic/gin"
)

// LoginReq 登录入参。
// - 正式：小程序端 wx.login 拿 code 传上来，后端 code2session 换 openid。
// - 本地联调：微信凭证未配置时，可直接传 openid 造一个用户，方便无微信环境开发。
type LoginReq struct {
	Code     string `json:"code"`
	OpenID   string `json:"openid"` // 仅本地联调用
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// Login 登录：换取 openid → 找到/创建用户 → 签发 JWT。
func (h *Handler) Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "参数错误")
		return
	}

	openid := req.OpenID
	if req.Code != "" && h.Cfg.WechatAppID != "" {
		sess, err := service.Code2Session(h.Cfg.WechatAppID, h.Cfg.WechatAppSecret, req.Code)
		if err != nil {
			response.Fail(c, response.CodeServer, "微信登录失败："+err.Error())
			return
		}
		openid = sess.OpenID
	}
	if openid == "" {
		response.Fail(c, response.CodeParamError, "缺少 code（或本地联调用的 openid）")
		return
	}

	var user model.User
	if err := h.DB.Where(model.User{OpenID: openid}).
		Attrs(model.User{Role: model.RoleOrderer}).
		FirstOrCreate(&user).Error; err != nil {
		response.Fail(c, response.CodeServer, "用户创建失败")
		return
	}
	// 有传昵称/头像则更新
	if req.Nickname != "" || req.Avatar != "" {
		h.DB.Model(&user).Updates(map[string]interface{}{"nickname": req.Nickname, "avatar": req.Avatar})
	}

	token, err := service.IssueToken(h.Cfg.JWTSecret, user.ID, user.OpenID)
	if err != nil {
		response.Fail(c, response.CodeServer, "签发 token 失败")
		return
	}
	response.OK(c, gin.H{"token": token, "user": user})
}

// Profile 返回当前登录用户。
func (h *Handler) Profile(c *gin.Context) {
	response.OK(c, middleware.CurrentUser(c))
}

// CoupleReq 创建/加入情侣空间入参。
type CoupleReq struct {
	Name       string     `json:"name"`
	Role       model.Role `json:"role"`       // cook | orderer
	InviteCode string     `json:"inviteCode"` // 加入时必填
}

// CreateCouple 创建情侣空间，生成邀请码，并把当前用户设为成员。
func (h *Handler) CreateCouple(c *gin.Context) {
	var req CoupleReq
	_ = c.ShouldBindJSON(&req)
	u := middleware.CurrentUser(c)

	couple := model.Couple{Name: orDefault(req.Name, "我们的小食记"), InviteCode: genInviteCode()}
	if err := h.DB.Create(&couple).Error; err != nil {
		response.Fail(c, response.CodeServer, "创建失败")
		return
	}
	h.DB.Model(u).Updates(map[string]interface{}{"couple_id": couple.ID, "role": orDefaultRole(req.Role, model.RoleCook)})

	// 新空间铺一份默认分组，让菜单不至于空着
	seedGroups(h, couple.ID)

	response.OK(c, couple)
}

// JoinCouple 用邀请码加入已有的情侣空间。
func (h *Handler) JoinCouple(c *gin.Context) {
	var req CoupleReq
	if err := c.ShouldBindJSON(&req); err != nil || req.InviteCode == "" {
		response.Fail(c, response.CodeParamError, "请填写邀请码")
		return
	}
	var couple model.Couple
	if err := h.DB.Where("invite_code = ?", strings.ToUpper(req.InviteCode)).First(&couple).Error; err != nil {
		response.Fail(c, response.CodeNotFound, "邀请码无效")
		return
	}
	u := middleware.CurrentUser(c)
	h.DB.Model(u).Updates(map[string]interface{}{"couple_id": couple.ID, "role": orDefaultRole(req.Role, model.RoleOrderer)})
	response.OK(c, couple)
}

// GetCouple 返回当前情侣空间信息及成员。
func (h *Handler) GetCouple(c *gin.Context) {
	cid, ok := h.coupleID(c)
	if !ok {
		return
	}
	var couple model.Couple
	if err := h.DB.First(&couple, cid).Error; err != nil {
		response.Fail(c, response.CodeNotFound, "空间不存在")
		return
	}
	var members []model.User
	h.DB.Where("couple_id = ?", cid).Find(&members)
	response.OK(c, gin.H{"couple": couple, "members": members})
}

// genInviteCode 生成 6 位大写字母数字邀请码（去掉易混字符）。
func genInviteCode() string {
	const charset = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func orDefault(v, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}

func orDefaultRole(v, def model.Role) model.Role {
	if v == model.RoleCook || v == model.RoleOrderer {
		return v
	}
	return def
}

// seedGroups 给新情侣空间铺默认分组。
func seedGroups(h *Handler, coupleID uint) {
	defaults := []string{"下饭菜", "汤汤水水", "解馋", "素菜"}
	for i, name := range defaults {
		h.DB.Create(&model.Group{CoupleID: coupleID, Name: name, Sort: i})
	}
}
