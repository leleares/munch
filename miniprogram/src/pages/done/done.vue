<script setup>
import { computed } from 'vue'
import { useOrderStore } from '../../stores/order'

const order = useOrderStore()
const last = computed(() => order.orders[0]) // 刚下的单排在最前

function again() {
  uni.switchTab({ url: '/pages/menu/menu' })
}
function goHistory() {
  uni.switchTab({ url: '/pages/orders/orders' })
}
</script>

<template>
  <view class="page">
    <text class="chef">🧑‍🍳</text>
    <text class="title">订单已送到大厨手机啦</text>
    <text class="sub">大厨已接单，安心等上菜～</text>

    <view v-if="last" class="mc-card summary">
      <text v-for="(it, idx) in last.items" :key="idx" class="sum-item">
        {{ it.name }} <text class="note">{{ it.spiceLabel }}{{ it.forbid ? ' · ' + it.forbid : '' }}</text> ×{{ it.qty }}
      </text>
      <text v-if="last.message" class="sum-msg">💌 {{ last.message }}</text>
    </view>

    <view class="btns">
      <view class="mc-btn ghost" @tap="goHistory">看看点过的菜 ›</view>
      <view class="mc-btn" @tap="again">再点一桌 🍚</view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.page { min-height: 100vh; padding: 80rpx $page-pad; display: flex; flex-direction: column; align-items: center; }
.chef { font-size: 120rpx; }
.title { font-size: 44rpx; font-weight: 700; color: $text-title; margin-top: 20rpx; }
.sub { font-size: 26rpx; color: $text-sub; margin-top: 12rpx; }
.summary { width: 100%; padding: 28rpx; margin-top: 48rpx; display: flex; flex-direction: column; gap: 12rpx; }
.sum-item { font-size: 28rpx; color: $text-main; }
.note { font-size: 22rpx; color: $text-weak; }
.sum-msg { margin-top: 12rpx; font-size: 24rpx; color: $text-sub; }
.btns { width: 100%; display: flex; gap: 20rpx; margin-top: 60rpx; }
.btns .mc-btn { flex: 1; }
</style>
