<script setup>
import { onShow, onHide } from '@dcloudio/uni-app'
import { useOrderStore, statusMeta } from '../../stores/order'

const order = useOrderStore()

onShow(() => order.startPolling(3000)) // 大厨端刷得勤一点
onHide(() => order.stopPolling())

async function advance(o) {
  try {
    await order.advance(o.id)
  } catch (e) {
    uni.showToast({ title: e.message, icon: 'none' })
  }
}

function fmtTime(t) {
  const d = new Date(t)
  const p = (n) => String(n).padStart(2, '0')
  return `${d.getMonth() + 1}月${d.getDate()}日 ${p(d.getHours())}:${p(d.getMinutes())}`
}

// 切回点菜端（menu 是 tab 页，用 switchTab）
function goGuest() {
  uni.switchTab({ url: '/pages/menu/menu' })
}
</script>

<template>
  <view class="page">
    <!-- 切回点菜端 FAB -->
    <view class="fab" @tap="goGuest">
      <image class="fab-icon" src="/static/icons/fab-bowl-white.png" mode="aspectFit" />
      <text class="fab-label">点菜端</text>
    </view>

    <view class="head">
      <text class="title">收到的点单</text>
      <text v-if="order.pendingCount" class="badge">{{ order.pendingCount }} 桌待处理</text>
    </view>

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

      <view v-if="statusMeta(o.status).next" class="mc-btn full" @tap="advance(o)">
        {{ statusMeta(o.status).next }}
      </view>
      <view v-else class="done-tip">已上菜 ✓ 慢慢吃～</view>
    </view>

    <view v-if="!order.orders.length" class="empty">☕ 还没有人点单，先歇会儿～</view>
  </view>
</template>

<style lang="scss" scoped>
.page { padding: 24rpx $page-pad; position: relative; }

/* 切端悬浮按钮（与点菜端同款） */
.fab {
  position: fixed; top: 20rpx; right: 24rpx; z-index: 50;
  width: 104rpx; height: 104rpx; border-radius: 28rpx;
  background: $sage; box-shadow: $shadow-fab;
  display: flex; flex-direction: column; align-items: center; justify-content: center;
}
.fab-icon { width: 46rpx; height: 46rpx; }
.fab-label { font-size: 18rpx; color: #fff; margin-top: 4rpx; }

.head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 24rpx; }
.title { font-size: 40rpx; font-weight: 700; color: $text-title; }
.badge { font-size: 22rpx; color: $status-pending-text; background: $status-pending-bg; padding: 8rpx 20rpx; border-radius: 24rpx; }

.order { padding: $card-pad; margin-bottom: 20rpx; }
.order-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 14rpx; }
.order-time { font-size: 24rpx; color: $text-sub; }
.order-items { display: flex; flex-direction: column; gap: 8rpx; }
.order-item { font-size: 28rpx; color: $text-main; }
.oi-note { font-size: 22rpx; color: $text-weak; }
.order-msg { display: block; margin: 14rpx 0; font-size: 24rpx; color: $text-sub; }
.mc-btn.full { width: 100%; margin-top: 18rpx; }
.done-tip { text-align: center; color: $status-served-text; font-size: 26rpx; margin-top: 16rpx; }
.empty { text-align: center; color: $text-weak; font-size: 28rpx; padding: 120rpx 0; }
</style>
