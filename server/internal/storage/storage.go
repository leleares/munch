package storage

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"munch/server/internal/config"
)

// Storage 抽象图片存储，方便在本地磁盘与腾讯云 COS 之间切换。
type Storage interface {
	// Save 保存文件内容，返回可公开访问的 URL。
	Save(filename string, r io.Reader) (string, error)
}

// New 按配置选择存储驱动。
func New(cfg *config.Config) Storage {
	switch cfg.StorageDriver {
	case "cos":
		// TODO: 接入腾讯云 COS（github.com/tencentyun/cos-go-sdk-v5）。
		// 微信云托管里推荐用 COS 存图；配好 COS_* 环境变量后在此返回 cosStorage。
		// 暂未实现前回退到本地磁盘，保证服务可跑。
		return newLocal(cfg)
	default:
		return newLocal(cfg)
	}
}

// localStorage 把文件落到 StaticDir，并通过 /static/<file> 提供访问。
type localStorage struct {
	dir     string
	baseURL string // PublicBaseURL，为空则返回相对路径 /static/<file>
}

func newLocal(cfg *config.Config) *localStorage {
	_ = os.MkdirAll(cfg.StaticDir, 0o755)
	return &localStorage{dir: cfg.StaticDir, baseURL: strings.TrimRight(cfg.PublicBaseURL, "/")}
}

func (s *localStorage) Save(filename string, r io.Reader) (string, error) {
	dst := filepath.Join(s.dir, filename)
	f, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return "", err
	}
	path := "/static/" + filename
	if s.baseURL != "" {
		return s.baseURL + path, nil
	}
	return path, nil
}
