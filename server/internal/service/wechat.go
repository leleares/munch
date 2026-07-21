package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// WechatSession 是 code2session 接口返回的关键字段。
type WechatSession struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

var httpClient = &http.Client{Timeout: 5 * time.Second}

// Code2Session 用小程序端 wx.login 拿到的 code 换取 openid。
// 仅在「自建/本地」登录路径使用；走微信云托管 callContainer 时平台会直接注入 X-WX-OPENID，无需调用它。
func Code2Session(appID, secret, code string) (*WechatSession, error) {
	if appID == "" || secret == "" {
		return nil, errors.New("未配置 WECHAT_APPID / WECHAT_SECRET")
	}
	api := "https://api.weixin.qq.com/sns/jscode2session?" + url.Values{
		"appid":      {appID},
		"secret":     {secret},
		"js_code":    {code},
		"grant_type": {"authorization_code"},
	}.Encode()

	resp, err := httpClient.Get(api)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var s WechatSession
	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return nil, err
	}
	if s.ErrCode != 0 {
		return nil, fmt.Errorf("code2session 失败: %d %s", s.ErrCode, s.ErrMsg)
	}
	return &s, nil
}
