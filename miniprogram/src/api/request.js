import {
  USE_CLOUD_CONTAINER,
  CLOUD_ENV,
  CLOUD_SERVICE,
  API_BASE_URL,
  API_PREFIX,
} from "../config";

const TOKEN_KEY = "munch_token";

export function getToken() {
  try {
    return uni.getStorageSync(TOKEN_KEY) || "";
  } catch (e) {
    return "";
  }
}

export function setToken(token) {
  uni.setStorageSync(TOKEN_KEY, token);
}

export function clearToken() {
  uni.removeStorageSync(TOKEN_KEY);
}

let cloudInited = false;

// 微信端首次调用前初始化云能力
function ensureCloudInit() {
  // #ifdef MP-WEIXIN
  if (!cloudInited && typeof wx !== "undefined" && wx.cloud) {
    wx.cloud.init({ env: CLOUD_ENV });
    cloudInited = true;
  }
  // #endif
}

/**
 * 统一请求。返回后端 Body.data（code=0 时），否则 reject 一个带 msg 的错误。
 * @param {string} path   如 '/dishes'（会自动拼 API_PREFIX）
 * @param {object} opts   { method, data, header }
 */
// 🐶 请求/响应日志（含方法、路径、通道、数据），方便在控制台排查
function logReq(method, fullPath, channel, data) {
  console.log(`🐶 请求 ${method} ${fullPath}  [${channel}]`, data);
}
function logRes(method, fullPath, body) {
  console.log(`🐶 响应 ${method} ${fullPath}`, body);
}
function logErr(method, fullPath, err) {
  console.warn(`🐶 错误 ${method} ${fullPath}`, err);
}

export function request(path, { method = "GET", data = {}, header = {} } = {}) {
  const fullPath = API_PREFIX + path;

  // ---- 微信云托管容器调用（免备案，自动注入 X-WX-OPENID）----
  // #ifdef MP-WEIXIN
  if (USE_CLOUD_CONTAINER) {
    ensureCloudInit();
    logReq(method, fullPath, "callContainer", data);
    return new Promise((resolve, reject) => {
      wx.cloud.callContainer({
        config: { env: CLOUD_ENV },
        path: fullPath,
        method,
        header: {
          "X-WX-SERVICE": CLOUD_SERVICE,
          "content-type": "application/json",
          ...header,
        },
        data,
        success: (res) => {
          logRes(method, fullPath, res.data);
          handleBody(res.data, resolve, reject);
        },
        fail: (err) => {
          logErr(method, fullPath, err.errMsg || err);
          reject(new Error(err.errMsg || "网络异常"));
        },
      });
    });
  }
  // #endif

  // ---- 普通 HTTP 请求（H5 / 本地联调，带 JWT）----
  logReq(method, fullPath, "http", data);
  return new Promise((resolve, reject) => {
    const token = getToken();
    uni.request({
      url: API_BASE_URL + fullPath,
      method,
      data,
      header: {
        "content-type": "application/json",
        ...(token ? { Authorization: "Bearer " + token } : {}),
        ...header,
      },
      success: (res) => {
        logRes(method, fullPath, res.data);
        handleBody(res.data, resolve, reject);
      },
      fail: (err) => {
        logErr(method, fullPath, err.errMsg || err);
        reject(new Error(err.errMsg || "网络异常"));
      },
    });
  });
}

function handleBody(body, resolve, reject) {
  if (!body || typeof body !== "object") {
    reject(new Error("返回格式异常"));
    return;
  }
  if (body.code === 0) {
    resolve(body.data);
  } else if (body.code === 401) {
    clearToken();
    reject(new Error(body.msg || "登录失效"));
  } else {
    const err = new Error(body.msg || "请求失败");
    err.code = body.code;
    reject(err);
  }
}
