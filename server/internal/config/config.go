package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config 汇总所有运行期配置，全部来自环境变量。
// 微信云托管会自动注入 MYSQL_ADDRESS / MYSQL_USERNAME / MYSQL_PASSWORD，
// 本地开发则从 server/.env 读取（见 .env.example）。
type Config struct {
	Port string // 服务监听端口，微信云托管默认要求监听 80

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	JWTSecret string // 自签 JWT 的密钥

	WechatAppID     string // 小程序 appid（本地 code2session 登录用）
	WechatAppSecret string // 小程序 secret（本地 code2session 登录用）

	// 是否允许「裸 openid 登录」——本地开发免微信环境的后门。
	// 生产环境必须为 false，否则任何人都能用任意 openid 伪造身份。
	AllowDevLogin bool

	// 图片存储：默认 local（落磁盘并通过 /static 提供），可切 cos
	StorageDriver string
	StaticDir     string // local 驱动的落盘目录
	PublicBaseURL string // 拼接图片可访问地址的前缀，如 https://xxx.com

	// 腾讯云 COS（StorageDriver=cos 时使用）
	COSSecretID  string
	COSSecretKey string
	COSBucketURL string
	COSRegion    string // 如 ap-beijing；不配则从 BucketURL 自动解析
	COSAppID     string // 如 1330007488；不配则从 bucket 名后缀自动解析
	COSBucket    string // 完整桶名，如 ares1-1330007488；从 BucketURL 自动解析
}

// Load 读取 .env（若存在）后从环境变量装配配置。
func Load() *Config {
	// 本地存在 .env 时加载；线上（微信云托管）没有该文件会安静跳过。
	_ = godotenv.Load()

	c := &Config{
		Port: env("PORT", "80"),

		// 优先用微信云托管注入的 MYSQL_* 变量，其次回退到通用 DB_* 变量。
		DBHost:     firstNonEmpty(splitHost(os.Getenv("MYSQL_ADDRESS")), env("DB_HOST", "127.0.0.1")),
		DBPort:     firstNonEmpty(splitPort(os.Getenv("MYSQL_ADDRESS")), env("DB_PORT", "3306")),
		DBUser:     firstNonEmpty(os.Getenv("MYSQL_USERNAME"), env("DB_USER", "root")),
		DBPassword: firstNonEmpty(os.Getenv("MYSQL_PASSWORD"), env("DB_PASSWORD", "")),
		DBName:     env("DB_NAME", "munch"),

		JWTSecret: env("JWT_SECRET", "munch-dev-secret-change-me"),

		WechatAppID:     os.Getenv("WECHAT_APPID"),
		WechatAppSecret: os.Getenv("WECHAT_SECRET"),

		// 默认关闭：线上不显式打开就一定是安全的
		AllowDevLogin: os.Getenv("ALLOW_DEV_LOGIN") == "true",

		StorageDriver: env("STORAGE_DRIVER", "local"),
		StaticDir:     env("STATIC_DIR", "./data/uploads"),
		PublicBaseURL: env("PUBLIC_BASE_URL", ""),

		COSSecretID:  os.Getenv("COS_SECRET_ID"),
		COSSecretKey: os.Getenv("COS_SECRET_KEY"),
		COSBucketURL: os.Getenv("COS_BUCKET_URL"),
		COSRegion:    os.Getenv("COS_REGION"),
		COSAppID:     os.Getenv("COS_APPID"),
	}

	// 容错：不少人会把 "host:port" 整个填进 DB_HOST，
	// 那样会和 DBPort 拼成 "host:port:port" 导致解析失败。这里拆开。
	if host, port, ok := splitHostPort(c.DBHost); ok {
		c.DBHost = host
		c.DBPort = port
	}

	// 从 COS_BUCKET_URL 自动解析 bucket / region / appid，省得再单独配。
	// 形如 https://ares1-1330007488.cos.ap-beijing.myqcloud.com
	//   bucket = ares1-1330007488, region = ap-beijing, appid = 1330007488
	if bucket, region := parseCOSBucketURL(c.COSBucketURL); bucket != "" {
		c.COSBucket = bucket
		if c.COSRegion == "" {
			c.COSRegion = region
		}
		if c.COSAppID == "" {
			if i := strings.LastIndex(bucket, "-"); i >= 0 {
				c.COSAppID = bucket[i+1:]
			}
		}
	}
	return c
}

// parseCOSBucketURL 从桶访问域名里拆出 bucket 与 region。
// 期望域名形如 <bucket>.cos.<region>.myqcloud.com，解析不出则返回空串。
func parseCOSBucketURL(raw string) (bucket, region string) {
	if raw == "" {
		return "", ""
	}
	host := raw
	if i := strings.Index(host, "://"); i >= 0 {
		host = host[i+3:]
	}
	if i := strings.IndexAny(host, "/:"); i >= 0 {
		host = host[:i]
	}
	// host = ares1-1330007488.cos.ap-beijing.myqcloud.com
	parts := strings.Split(host, ".")
	if len(parts) >= 3 && parts[1] == "cos" {
		return parts[0], parts[2]
	}
	return "", ""
}

// splitHostPort 识别 "host:port" 形式，端口必须是纯数字才认。
func splitHostPort(v string) (host, port string, ok bool) {
	i := strings.LastIndex(v, ":")
	if i <= 0 || i == len(v)-1 {
		return "", "", false
	}
	host, port = v[:i], v[i+1:]
	if strings.Contains(host, ":") {
		return "", "", false // IPv6 之类，不动
	}
	for _, r := range port {
		if r < '0' || r > '9' {
			return "", "", false
		}
	}
	return host, port, true
}

// DSN 返回 GORM 连接 MySQL 用的 data source name。
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
	)
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// MYSQL_ADDRESS 形如 "10.0.0.1:3306"，拆出 host。
func splitHost(addr string) string {
	if addr == "" {
		return ""
	}
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return addr[:i]
		}
	}
	return addr
}

// MYSQL_ADDRESS 形如 "10.0.0.1:3306"，拆出 port。
func splitPort(addr string) string {
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return addr[i+1:]
		}
	}
	return ""
}
