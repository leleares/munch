package storage

import (
	"io"
	"log"
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
		if cfg.COSSecretID == "" || cfg.COSSecretKey == "" || cfg.COSBucketURL == "" {
			log.Println("[storage] STORAGE_DRIVER=cos 但 COS_* 环境变量不全，回退到本地磁盘")
			return newLocal(cfg)
		}
		s, err := newCOS(cfg)
		if err != nil {
			log.Printf("[storage] COS 初始化失败，回退到本地磁盘: %v", err)
			return newLocal(cfg)
		}
		log.Println("[storage] 使用腾讯云 COS 存储图片")
		return s
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
