package handler

import (
	"fmt"

	"munch/server/pkg/response"

	"github.com/gin-gonic/gin"
	sts "github.com/tencentyun/qcloud-cos-sts-sdk/go"
)

// COSCredential 签发 COS 临时密钥（STS），供小程序前端直传图片用。
// 走现有鉴权中间件：云托管注入 X-WX-OPENID，这里只是发凭证、不碰图片。
// 临时密钥仅允许 PutObject 到本桶的 munch/ 前缀，15 分钟过期，泄露也无大碍。
func (h *Handler) COSCredential(c *gin.Context) {
	if _, ok := h.coupleID(c); !ok {
		return
	}
	cfg := h.Cfg
	if cfg.COSSecretID == "" || cfg.COSSecretKey == "" || cfg.COSBucket == "" ||
		cfg.COSRegion == "" || cfg.COSAppID == "" {
		response.Fail(c, response.CodeServer, "COS 未正确配置")
		return
	}

	// 只授权上传到 munch/ 前缀，资源 ARN 形如：
	// qcs::cos:ap-beijing:uid/1330007488:ares1-1330007488/munch/*
	resource := fmt.Sprintf("qcs::cos:%s:uid/%s:%s/munch/*",
		cfg.COSRegion, cfg.COSAppID, cfg.COSBucket)

	client := sts.NewClient(cfg.COSSecretID, cfg.COSSecretKey, nil)
	res, err := client.GetCredential(&sts.CredentialOptions{
		DurationSeconds: 900, // 15 分钟
		Region:          cfg.COSRegion,
		Policy: &sts.CredentialPolicy{
			Version: "2.0",
			Statement: []sts.CredentialPolicyStatement{{
				Effect: "allow",
				Action: []string{
					"name/cos:PutObject",     // 简单上传
					"name/cos:PostObject",    // 表单上传（小程序 SDK 用）
					"name/cos:InitiateMultipartUpload",
					"name/cos:ListMultipartUploads",
					"name/cos:ListParts",
					"name/cos:UploadPart",
					"name/cos:CompleteMultipartUpload",
				},
				Resource: []string{resource},
			}},
		},
	})
	if err != nil || res.Credentials == nil {
		response.Fail(c, response.CodeServer, "获取临时密钥失败")
		return
	}

	response.OK(c, gin.H{
		"tmpSecretId":  res.Credentials.TmpSecretID,
		"tmpSecretKey": res.Credentials.TmpSecretKey,
		"sessionToken": res.Credentials.SessionToken,
		"startTime":    res.StartTime,
		"expiredTime":  res.ExpiredTime,
		"bucket":       cfg.COSBucket,
		"region":       cfg.COSRegion,
		"prefix":       "munch/",
	})
}
