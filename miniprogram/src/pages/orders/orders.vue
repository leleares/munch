<script setup>
import { ref } from 'vue'
import { onShow, onHide } from '@dcloudio/uni-app'
import { useMenuStore } from '../../stores/menu'
import { useCartStore } from '../../stores/cart'
import { useOrderStore, statusMeta } from '../../stores/order'
import { useUserStore } from '../../stores/user'
import { api } from '../../api'

const menu = useMenuStore()
const cart = useCartStore()
const order = useOrderStore()
const user = useUserStore()

const tab = ref('fav') // fav | shop | hist
const shopItems = ref([])
const shopText = ref('')

onShow(async () => {
  if (!user.hasCouple) { uni.reLaunch({ url: '/pages/bind/bind' }); return }
  if (!menu.dishes.length) await menu.loadAll()
  await loadShop()
  order.startPolling() // 记录页停留期间轮询订单状态变化
})
onHide(() => order.stopPolling())

async function loadShop() {
  shopItems.value = (await api.listShopItems()) || []
}
async function addShop() {
  const t = shopText.value.trim()
  if (!t) return uni.showToast({ title: '写点要买的吧～', icon: 'none' })
  const item = await api.createShopItem({ text: t })
  shopItems.value.unshift(item)
  shopText.value = ''
}
async function toggleShop(item) {
  const updated = await api.updateShopItem(item.id, { done: !item.done })
  item.done = updated.done
}
async function delShop(item) {
  await api.deleteShopItem(item.id)
  shopItems.value = shopItems.value.filter((i) => i.id !== item.id)
}

function addFav(d) {
  cart.inc(d.id)
  uni.showToast({ title: `「${d.name}」已加入清单 🌿`, icon: 'none' })
}

function fmtTime(t) {
  // 后端返回 ISO 时间，简单格式化成 M月D日 HH:mm
  const d = new Date(t)
  const p = (n) => String(n).padStart(2, '0')
  return `${d.getMonth() + 1}月${d.getDate()}日 ${p(d.getHours())}:${p(d.getMinutes())}`
}
</script>

<template>
  <view class="page">
    <view class="seg">
      <text class="seg-item" :class="{ on: tab === 'fav' }" @tap="tab = 'fav'">常点 ⭐</text>
      <text class="seg-item" :class="{ on: tab === 'shop' }" @tap="tab = 'shop'">买菜清单 🛒</text>
      <text class="seg-item" :class="{ on: tab === 'hist' }" @tap="tab = 'hist'">点菜记录 📖</text>
    </view>

    <!-- 常点 -->
    <view v-if="tab === 'fav'" class="pane">
      <view class="fav-chips">
        <text v-for="d in menu.favDishes" :key="d.id" class="mc-chip fav" @tap="addFav(d)">{{ d.name }}</text>
      </view>
      <view v-if="!menu.favDishes.length" class="empty">还没有常点的菜～</view>
    </view>

    <!-- 买菜清单 -->
    <view v-else-if="tab === 'shop'" class="pane">
      <view class="shop-add">
        <input class="mc-input" v-model="shopText" placeholder="要买点什么？" @confirm="addShop" />
        <view class="mc-btn add-btn" @tap="addShop">加</view>
      </view>
      <view class="shop-list">
        <view v-for="item in shopItems" :key="item.id" class="shop-item">
          <view class="check" :class="{ on: item.done }" @tap="toggleShop(item)">
            <text v-if="item.done">✓</text>
          </view>
          <text class="shop-text" :class="{ done: item.done }">{{ item.text }}</text>
          <text class="del" @tap="delShop(item)">×</text>
        </view>
        <view v-if="!shopItems.length" class="empty">清单空空，加点要买的吧 🛒</view>
      </view>
    </view>

    <!-- 点菜记录 -->
    <view v-else class="pane">
      <view v-for="o in order.orders" :key="o.id" class="mc-card order">
        <view class="order-head">
          <text class="order-time">{{ fmtTime(o.createdAt) }}</text>
          <text class="mc-status" :class="statusMeta(o.status).cls">{{ statusMeta(o.status).label }}</text>
        </view>
        <view class="order-items">
          <text v-for="(it, idx) in o.items" :key="idx" class="order-item">
            {{ it.name }} <text class="oi-note">{{ it.spiceLabel }}{{ it.forbid ? ' · ' + it.forbid : '' }}</text> ×{{ it.qty }}
          </text>
        </view>
        <text v-if="o.message" class="order-msg">💌 {{ o.message }}</text>
      </view>
      <view v-if="!order.orders.length" class="empty">还没有点过菜，去点一桌吧 🍚</view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.page { padding: 24rpx $page-pad; }
.seg { display: flex; gap: 14rpx; margin-bottom: 24rpx; }
.seg-item {
  flex: 1; text-align: center; padding: 18rpx 0; border-radius: $radius-chip;
  font-size: 26rpx; color: $text-sub; background: $card-bg; border: 2rpx solid $card-border;
  &.on { color: #fff; background: $sage; border-color: $sage; }
}
.pane { min-height: 60vh; }
.empty { text-align: center; color: $text-weak; font-size: 26rpx; padding: 80rpx 0; }

.fav-chips { display: flex; flex-wrap: wrap; gap: 16rpx; }
.mc-chip.fav { background: $sage-soft-bg; color: $sage-deep; border-color: $sage-soft-bd; }

.shop-add { display: flex; gap: 16rpx; margin-bottom: 24rpx; }
.add-btn { width: 96rpx; height: 84rpx; flex-shrink: 0; }
.shop-list { display: flex; flex-direction: column; gap: 8rpx; }
.shop-item { display: flex; align-items: center; padding: 20rpx 4rpx; border-bottom: 2rpx solid $divider; }
.check { width: 44rpx; height: 44rpx; border-radius: 12rpx; border: 3rpx solid $input-border; display: flex; align-items: center; justify-content: center; color: #fff; font-size: 28rpx; }
.check.on { background: $sage; border-color: $sage; }
.shop-text { flex: 1; margin: 0 20rpx; font-size: 28rpx; color: $text-main; }
.shop-text.done { color: $text-weak; text-decoration: line-through; }
.del { font-size: 40rpx; color: $text-weak; padding: 0 10rpx; }

.order { padding: $card-pad; margin-bottom: 20rpx; }
.order-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 14rpx; }
.order-time { font-size: 24rpx; color: $text-sub; }
.order-items { display: flex; flex-direction: column; gap: 8rpx; }
.order-item { font-size: 28rpx; color: $text-main; }
.oi-note { font-size: 22rpx; color: $text-weak; }
.order-msg { display: block; margin-top: 14rpx; font-size: 24rpx; color: $text-sub; }
</style>
