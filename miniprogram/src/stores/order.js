import { defineStore } from 'pinia'
import { api } from '../api'

const STATUS = {
  pending: { label: '待接单', cls: 'pending', next: '接单开做 🍳' },
  cooking: { label: '备菜中', cls: 'cooking', next: '标记上菜 🍽' },
  served: { label: '已上菜', cls: 'served', next: '' },
}

export function statusMeta(status) {
  return STATUS[status] || STATUS.pending
}

export const useOrderStore = defineStore('order', {
  state: () => ({
    orders: [],
    _timer: null,
  }),
  getters: {
    // 大厨端待处理数量
    pendingCount: (s) => s.orders.filter((o) => o.status !== 'served').length,
  },
  actions: {
    async load() {
      this.orders = (await api.listOrders()) || []
    },
    async placeOrder(payload) {
      const order = await api.createOrder(payload)
      this.orders.unshift(order)
      return order
    },
    async advance(id) {
      const updated = await api.advanceOrder(id)
      const i = this.orders.findIndex((o) => o.id === id)
      if (i >= 0) this.orders[i] = { ...this.orders[i], status: updated.status }
      return updated
    },

    // 轮询：记录页 / 大厨端进入时开启，离开时关闭。实时性靠它（不上订阅消息）。
    startPolling(intervalMs = 4000) {
      this.stopPolling()
      this.load()
      this._timer = setInterval(() => this.load(), intervalMs)
    },
    stopPolling() {
      if (this._timer) {
        clearInterval(this._timer)
        this._timer = null
      }
    },
  },
})
