/**
 * 运行时配置。部署微信云托管后回来把 CLOUD_ENV / CLOUD_SERVICE 填上。
 */

// 是否走微信云托管容器调用（wx.cloud.callContainer）。
// true：微信端通过云调用后端，平台自动注入 X-WX-OPENID，免备案（推荐）。
// false：走普通 HTTP 请求（H5 / 本地联调用）。
export const USE_CLOUD_CONTAINER = false;

// 微信云托管环境 ID（控制台 → 环境 → 环境ID），部署后填。
export const CLOUD_ENV = "";

// 微信云托管服务名（部署 server 时起的服务名），部署后填。
export const CLOUD_SERVICE = "munch-server";

// H5 / 本地联调时的后端地址（USE_CLOUD_CONTAINER=false 时生效）。
// 本地：go run 起在 8080；或 docker compose 映射的 8080。
export const API_BASE_URL = "http://127.0.0.1:8080";

// 所有接口的统一前缀
export const API_PREFIX = "/api";
