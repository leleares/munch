import { defineStore } from 'pinia'
import { api } from '../api'
import { getToken, setToken, clearToken } from '../api/request'
import { USE_CLOUD_CONTAINER } from '../config'

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
     * - 微信端（callContainer）：openid 由平台注入，直接拉 /profile。
     * - H5 / 本地：用 token 或 dev openid 登录后拉 /profile。
     */
    async bootstrap() {
      try {
        // #ifdef MP-WEIXIN
        if (USE_CLOUD_CONTAINER) {
          this.user = await api.profile()
          this.ready = true
          return
        }
        // 非云调用的微信端：走 wx.login 换 code 登录
        await this.wxLogin()
        this.ready = true
        return
        // #endif

        // #ifndef MP-WEIXIN
        if (!getToken()) {
          // H5 本地联调：用一个固定 dev openid 造用户，免真机微信环境
          const { token, user } = await api.login({ openid: 'h5-dev-openid', nickname: '亲爱的' })
          setToken(token)
          this.user = user
        } else {
          this.user = await api.profile()
        }
        this.ready = true
        // #endif
      } catch (e) {
        this.ready = true
        console.warn('[bootstrap]', e.message)
      }
    },

    // #ifdef MP-WEIXIN
    wxLogin() {
      return new Promise((resolve, reject) => {
        uni.login({
          provider: 'weixin',
          success: async ({ code }) => {
            try {
              const { token, user } = await api.login({ code })
              setToken(token)
              this.user = user
              resolve()
            } catch (e) {
              reject(e)
            }
          },
          fail: reject,
        })
      })
    },
    // #endif

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
