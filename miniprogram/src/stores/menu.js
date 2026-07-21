import { defineStore } from 'pinia'
import { api } from '../api'

export const SPICE_LABELS = ['不辣', '微辣 🌶', '中辣 🌶🌶', '重辣 🌶🌶🌶']

export const useMenuStore = defineStore('menu', {
  state: () => ({
    dishes: [],
    groups: [],
    cat: '全部', // 当前分类名，'全部' 表示不过滤
    loading: false,
    editingId: null, // 长按菜品进入「加新菜」tab 编辑时，用它传递被编辑的菜 id
  }),
  getters: {
    // 按当前分类过滤后的菜品
    visibleDishes: (s) => {
      if (s.cat === '全部') return s.dishes
      const g = s.groups.find((x) => x.name === s.cat)
      if (!g) return s.dishes
      return s.dishes.filter((d) => d.groupId === g.id)
    },
    favDishes: (s) => s.dishes.filter((d) => d.isFav),
    dishById: (s) => (id) => s.dishes.find((d) => d.id === id),
    groupById: (s) => (id) => s.groups.find((g) => g.id === id),
  },
  actions: {
    async loadAll() {
      this.loading = true
      try {
        const [dishes, groups] = await Promise.all([api.listDishes(), api.listGroups()])
        this.dishes = dishes || []
        this.groups = groups || []
      } finally {
        this.loading = false
      }
    },
    setCat(name) {
      this.cat = name
    },
    async addDish(payload) {
      const dish = await api.createDish(payload)
      this.dishes.unshift(dish)
      return dish
    },
    async updateDish(id, payload) {
      const dish = await api.updateDish(id, payload)
      const i = this.dishes.findIndex((d) => d.id === id)
      if (i >= 0) this.dishes[i] = dish
      return dish
    },
    async removeDish(id) {
      await api.deleteDish(id)
      this.dishes = this.dishes.filter((d) => d.id !== id)
    },
    async addGroup(name) {
      const g = await api.createGroup({ name })
      this.groups.push(g)
      return g
    },
    async renameGroup(id, name) {
      const old = this.groupById(id)
      const oldName = old && old.name
      await api.updateGroup(id, { name })
      if (old) old.name = name
      if (this.cat === oldName) this.cat = name
    },
    async removeGroup(id) {
      const res = await api.deleteGroup(id)
      const removed = this.groupById(id)
      this.groups = this.groups.filter((g) => g.id !== id)
      // 该组菜品被后端迁到 fallback，前端同步一下
      if (res && res.fallbackGroupId) {
        this.dishes.forEach((d) => {
          if (d.groupId === id) d.groupId = res.fallbackGroupId
        })
      }
      if (removed && this.cat === removed.name) this.cat = '全部'
    },
  },
})
