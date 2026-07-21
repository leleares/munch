import {
  USE_CLOUD_CONTAINER,
  CLOUD_ENV,
  CLOUD_SERVICE,
  API_BASE_URL,
  API_PREFIX,
} from '../config'

const TOKEN_KEY = 'munch_token'

export function getToken() {
  try {
    return uni.getStorageSync(TOKEN_KEY) || ''
  } catch (e) {
    return ''
  }
}

export function setToken(token) {
  uni.setStorageSync(TOKEN_KEY, token)
}

export function clearToken() {
  uni.removeStorageSync(TOKEN_KEY)
}

let cloudInited = false

// 微信端首次调用前初始化云能力
function ensureCloudInit() {
  // #ifdef MP-WEIXIN
  if (!cloudInited && typeof wx !== 'undefined' && wx.cloud) {
    wx.cloud.init({ env: CLOUD_ENV })
    cloudInited = true
  }
  // #endif
}

/**
 * 统一请求。返回后端 Body.data（code=0 时），否则 reject 一个带 msg 的错误。
 * @param {string} path   如 '/dishes'（会自动拼 API_PREFIX）
 * @param {object} opts   { method, data, header }
 */
export function request(path, { method = 'GET', data = {}, header = {} } = {}) {
  const fullPath = API_PREFIX + path

  // ---- 微信云托管容器调用（免备案，自动注入 X-WX-OPENID）----
  // #ifdef MP-WEIXIN
  if (USE_CLOUD_CONTAINER) {
    ensureCloudInit()
    return new Promise((resolve, reject) => {
      wx.cloud.callContainer({
        config: { env: CLOUD_ENV },
        path: fullPath,
        method,
        header: { 'X-WX-SERVICE': CLOUD_SERVICE, 'content-type': 'application/json', ...header },
        data,
        success: (res) => handleBody(res.data, resolve, reject),
        fail: (err) => reject(new Error(err.errMsg || '网络异常')),
      })
    })
  }
  // #endif

  // ---- 普通 HTTP 请求（H5 / 本地联调，带 JWT）----
  return new Promise((resolve, reject) => {
    const token = getToken()
    uni.request({
      url: API_BASE_URL + fullPath,
      method,
      data,
      header: {
        'content-type': 'application/json',
        ...(token ? { Authorization: 'Bearer ' + token } : {}),
        ...header,
      },
      success: (res) => handleBody(res.data, resolve, reject),
      fail: (err) => reject(new Error(err.errMsg || '网络异常')),
    })
  })
}

function handleBody(body, resolve, reject) {
  if (!body || typeof body !== 'object') {
    reject(new Error('返回格式异常'))
    return
  }
  if (body.code === 0) {
    resolve(body.data)
  } else if (body.code === 401) {
    clearToken()
    reject(new Error(body.msg || '登录失效'))
  } else {
    const err = new Error(body.msg || '请求失败')
    err.code = body.code
    reject(err)
  }
}
