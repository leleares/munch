import { defineStore } from 'pinia'
import { SPICE_LABELS } from './menu'

export const useCartStore = defineStore('cart', {
  state: () => ({
    items: {}, // { dishId: qty }
    notes: {}, // { dishId: { spice: 0-3, forbid: '' } }
    msg: '', // 给大厨的留言草稿
  }),
  getters: {
    count: (s) => Object.values(s.items).reduce((a, b) => a + b, 0),
    isEmpty: (s) => Object.keys(s.items).length === 0,
    qtyOf: (s) => (id) => s.items[id] || 0,
    // 某道菜的备注展示文案：辣度 · 忌口
    noteText: (s) => (id, dishSpice) => {
      const n = s.notes[id] || {}
      const spice = n.spice != null ? n.spice : dishSpice
      const parts = [SPICE_LABELS[spice] || SPICE_LABELS[0]]
      if (n.forbid) parts.push(n.forbid)
      return parts.join(' · ')
    },
  },
  actions: {
    inc(id) {
      this.items[id] = (this.items[id] || 0) + 1
    },
    dec(id) {
      const q = (this.items[id] || 0) - 1
      if (q <= 0) delete this.items[id]
      else this.items[id] = q
    },
    setNote(id, note) {
      this.notes[id] = { ...(this.notes[id] || {}), ...note }
    },
    clear() {
      this.items = {}
      this.notes = {}
      this.msg = ''
    },
    // 组装下单 payload（items 带辣度/忌口快照）
    buildOrderPayload(menuStore) {
      const items = Object.entries(this.items).map(([id, qty]) => {
        const dish = menuStore.dishById(Number(id))
        const n = this.notes[id] || {}
        const spice = n.spice != null ? n.spice : dish ? dish.spice : 0
        return {
          dishId: Number(id),
          qty,
          spiceLabel: SPICE_LABELS[spice] || SPICE_LABELS[0],
          forbid: n.forbid || '',
        }
      })
      return { items, message: this.msg }
    },
  },
})
