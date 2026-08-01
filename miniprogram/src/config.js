/**
 * 运行时配置。部署微信云托管后回来把 CLOUD_ENV / CLOUD_SERVICE 填上。
 */

// 是否走微信云托管容器调用（wx.cloud.callContainer）。
// true：微信端通过云调用后端，平台自动注入 X-WX-OPENID，免备案（推荐）。
// false：走普通 HTTP 请求（H5 / 本地联调用）。
export const USE_CLOUD_CONTAINER = true;

// 微信云托管环境 ID（控制台 → 环境 → 环境ID）
export const CLOUD_ENV = "prod-d3gauzkf662e0a7bc";

// 微信云托管服务名（部署 server 时起的服务名），部署后填。
export const CLOUD_SERVICE = "munch-server";

// H5 / 本地联调时的后端地址（USE_CLOUD_CONTAINER=false 时生效）。
// 本地：go run / docker compose 都映射到宿主机 18090（8080 被 nginx 占用）。
export const API_BASE_URL = "http://127.0.0.1:18090";

// 所有接口的统一前缀
export const API_PREFIX = "/api";

// ---- 字体：霞鹜文楷（设计稿指定）----
// 小程序不能用 CSS @font-face 加载网络字体，必须走 loadFontFace，且只支持 woff/woff2/ttf。
// 这里用的是子集化后的一级常用字（约 971KB），由后端 /assets 提供。
// 上线后建议把字体传到 COS/CDN，把 FONT_URL 换成那个 https 地址，
// 并在小程序后台把该域名加进「downloadFile 合法域名」，否则真机加载不到。
export const FONT_FAMILY = "LXGW WenKai TC";
// 线上：指向微信云托管服务的公网域名（该域名需加进小程序后台「downloadFile 合法域名」）。
// 本地开发想切回去，改成 API_BASE_URL + "/assets/fonts/lxgw-wenkai-tc-subset.woff2" 即可。
export const FONT_URL =
  "https://munch-server-286616-10-1457919674.sh.run.tcloudbase.com/assets/fonts/lxgw-wenkai-tc-subset.woff2";

// ---- COS 图片直传（非机密信息，密钥仍由后端 STS 签发）----
// 桶名和地域是公开信息，放前端只为直传时拼参数；SecretId/Key 绝不放这里。
// 域名 ares1-1330007488.cos.ap-beijing.myqcloud.com 需加进小程序后台
// 「uploadFile 合法域名」（直传）和「downloadFile 合法域名」（显示图）。
export const COS_BUCKET = "ares1-1330007488";
export const COS_REGION = "ap-beijing";
