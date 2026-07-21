import { request } from './request'

// 登录 & 情侣空间
export const api = {
  login: (data) => request('/login', { method: 'POST', data }),
  profile: () => request('/profile'),

  createCouple: (data) => request('/couple', { method: 'POST', data }),
  joinCouple: (data) => request('/couple/join', { method: 'POST', data }),
  getCouple: () => request('/couple'),

  // 菜品
  listDishes: (query = {}) => {
    const qs = Object.entries(query)
      .filter(([, v]) => v !== undefined && v !== '')
      .map(([k, v]) => `${k}=${encodeURIComponent(v)}`)
      .join('&')
    return request('/dishes' + (qs ? '?' + qs : ''))
  },
  createDish: (data) => request('/dishes', { method: 'POST', data }),
  updateDish: (id, data) => request(`/dishes/${id}`, { method: 'PATCH', data }),
  deleteDish: (id) => request(`/dishes/${id}`, { method: 'DELETE' }),

  // 分组
  listGroups: () => request('/groups'),
  createGroup: (data) => request('/groups', { method: 'POST', data }),
  updateGroup: (id, data) => request(`/groups/${id}`, { method: 'PATCH', data }),
  deleteGroup: (id) => request(`/groups/${id}`, { method: 'DELETE' }),

  // 订单
  listOrders: () => request('/orders'),
  createOrder: (data) => request('/orders', { method: 'POST', data }),
  advanceOrder: (id, status) => request(`/orders/${id}/status`, { method: 'PATCH', data: { status } }),

  // 买菜清单
  listShopItems: () => request('/shop-items'),
  createShopItem: (data) => request('/shop-items', { method: 'POST', data }),
  updateShopItem: (id, data) => request(`/shop-items/${id}`, { method: 'PATCH', data }),
  deleteShopItem: (id) => request(`/shop-items/${id}`, { method: 'DELETE' }),
}
