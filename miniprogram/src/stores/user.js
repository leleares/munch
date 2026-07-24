import { defineStore } from 'pinia'
import { api } from '../api'
import { getToken, setToken, clearToken } from '../api/request'
import { USE_CLOUD_CONTAINER } from '../config'

/**
 * 本地开发用的稳定 openid：每个安装/设备生成一个并持久化。
 * 这样 H5 浏览器、微信开发者工具、不同真机各自是独立用户——
 * 一个客户端建情侣空间拿邀请码，另一个客户端加入，即可本地联调两个角色。
 * （清缓存/清数据缓存会重置成新用户，回到绑定页。）
 */
function devOpenId() {
  const KEY = 'munch_dev_openid'
  let id = ''
  try {
    id = uni.getStorageSync(KEY)
  } catch (e) {}
  if (!id) {
    id = 'dev-' + Date.now() + '-' + Math.floor(Math.random() * 1e6)
    try {
      uni.setStorageSync(KEY, id)
    } catch (e) {}
  }
  return id
}

export const useUserStore = defineStore('user', {
  state: () => ({
    user: null, // { id, nickname, role, coupleId, ... }
    ready: false, // bootstrap 是否完成
  }),
  getters: {
    hasCouple: (s) => !!(s.user && s.user.coupleId),
    isCook: (s) => s.user && s.user.role === 'cook',
  },
  actions: {
    /**
     * 启动流程：确保拿到用户身份。
     * - 已接微信云托管（USE_CLOUD_CONTAINER=true）：openid 由平台注入，直接拉 /profile。
     * - 本地开发（H5，或未接云托管的微信开发者工具）：用稳定的 dev openid 登录，
     *   免 appid/secret 与云环境，直接把功能跑通。
     */
    async bootstrap() {
      try {
        // #ifdef MP-WEIXIN
        if (USE_CLOUD_CONTAINER) {
          this.user = await api.profile()
          this.ready = true
          return
        }
        // #endif

        // 本地开发：有 token 就续用，否则用 dev openid 登录换 token
        if (getToken()) {
          this.user = await api.profile()
        } else {
          const { token, user } = await api.login({ openid: devOpenId(), nickname: '亲爱的' })
          setToken(token)
          this.user = user
        }
        this.ready = true
      } catch (e) {
        this.ready = true
        console.warn('[bootstrap]', e.message)
      }
    },

    async refreshProfile() {
      this.user = await api.profile()
    },

    async createCouple(payload) {
      await api.createCouple(payload)
      await this.refreshProfile()
    },

    async joinCouple(payload) {
      await api.joinCouple(payload)
      await this.refreshProfile()
    },

    logout() {
      clearToken()
      this.user = null
    },
  },
})
