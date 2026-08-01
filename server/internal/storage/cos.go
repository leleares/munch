package storage

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strings"

	"munch/server/internal/config"

	"github.com/tencentyun/cos-go-sdk-v5"
)

// cosStorage 把文件上传到腾讯云 COS，返回桶的公开访问地址。
// 跨端（web / native / 小程序）都能直接用这个 URL 引用图片。
type cosStorage struct {
	client  *cos.Client
	baseURL string // 桶访问域名，如 https://ares1-1330007488.cos.ap-beijing.myqcloud.com
}

// newCOS 用账号密钥初始化 COS 客户端。密钥只从环境变量读，绝不落代码/前端。
func newCOS(cfg *config.Config) (*cosStorage, error) {
	u, err := url.Parse(cfg.COSBucketURL)
	if err != nil {
		return nil, fmt.Errorf("COS_BUCKET_URL 解析失败: %w", err)
	}
	client := cos.NewClient(&cos.BaseURL{BucketURL: u}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cfg.COSSecretID,
			SecretKey: cfg.COSSecretKey,
		},
	})
	return &cosStorage{client: client, baseURL: strings.TrimRight(cfg.COSBucketURL, "/")}, nil
}

func (s *cosStorage) Save(filename string, r io.Reader) (string, error) {
	// 统一放到 munch/ 前缀下，和桶里其他业务隔开
	key := "munch/" + filename
	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: contentTypeByName(filename),
		},
	}
	if _, err := s.client.Object.Put(context.Background(), key, r, opt); err != nil {
		return "", err
	}
	return s.baseURL + "/" + key, nil
}

// contentTypeByName 按扩展名猜 MIME，让浏览器能直接内联预览图片。
func contentTypeByName(name string) string {
	if t := mime.TypeByExtension(path.Ext(name)); t != "" {
		return t
	}
	return "application/octet-stream"
}
