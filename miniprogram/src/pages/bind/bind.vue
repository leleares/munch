<script setup>
import { ref } from 'vue'
import { useUserStore } from '../../stores/user'

const user = useUserStore()
const tab = ref('create') // create | join
const coupleName = ref('我们的小食记')
const role = ref('cook') // 默认自己是大厨
const inviteCode = ref('')
const submitting = ref(false)

async function onCreate() {
  submitting.value = true
  try {
    await user.createCouple({ name: coupleName.value, role: role.value })
    uni.showToast({ title: '空间创建好啦 🌿', icon: 'none' })
    setTimeout(() => uni.switchTab({ url: '/pages/menu/menu' }), 600)
  } catch (e) {
    uni.showToast({ title: e.message, icon: 'none' })
  } finally {
    submitting.value = false
  }
}

async function onJoin() {
  if (!inviteCode.value.trim()) return uni.showToast({ title: '填一下邀请码', icon: 'none' })
  submitting.value = true
  try {
    await user.joinCouple({ inviteCode: inviteCode.value.trim(), role: role.value })
    uni.showToast({ title: '已加入，一起吃饭啦 🍚', icon: 'none' })
    setTimeout(() => uni.switchTab({ url: '/pages/menu/menu' }), 600)
  } catch (e) {
    uni.showToast({ title: e.message, icon: 'none' })
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <view class="wrap">
    <view class="hero">
      <text class="emoji">🌿</text>
      <text class="title">小食记</text>
      <text class="sub">你点菜，我下厨——先建一个只属于你俩的小空间</text>
    </view>

    <view class="seg">
      <text class="seg-item" :class="{ on: tab === 'create' }" @tap="tab = 'create'">我来创建</text>
      <text class="seg-item" :class="{ on: tab === 'join' }" @tap="tab = 'join'">我有邀请码</text>
    </view>

    <view class="mc-card panel">
      <template v-if="tab === 'create'">
        <text class="label">给你们的小空间起个名字</text>
        <input class="mc-input" :value="coupleName" @input="coupleName = $event.detail.value" placeholder="我们的小食记" />
      </template>
      <template v-else>
        <text class="label">输入对方给你的邀请码</text>
        <input class="mc-input code" :value="inviteCode" @input="inviteCode = $event.detail.value" placeholder="6 位邀请码" maxlength="6" />
      </template>

      <text class="label mt">我是</text>
      <view class="roles">
        <text class="mc-chip" :class="{ on: role === 'cook' }" @tap="role = 'cook'">大厨 🍳</text>
        <text class="mc-chip" :class="{ on: role === 'orderer' }" @tap="role = 'orderer'">点菜的 🥢</text>
      </view>

      <view class="mc-btn full" :class="{ disabled: submitting }" @tap="tab === 'create' ? onCreate() : onJoin()">
        {{ tab === 'create' ? '创建我们的空间 🌿' : '加入 TA 的空间 🍚' }}
      </view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.wrap { min-height: 100vh; padding: $page-pad; box-sizing: border-box; }
.hero { display: flex; flex-direction: column; align-items: center; padding: 60rpx 0 40rpx; }
.hero .emoji { font-size: 88rpx; }
.hero .title { font-size: 52rpx; font-weight: 700; color: $text-title; margin-top: 12rpx; }
.hero .sub { font-size: 26rpx; color: $text-sub; margin-top: 12rpx; text-align: center; }

.seg { display: flex; gap: 16rpx; margin-bottom: 24rpx; }
.seg-item {
  flex: 1; text-align: center; padding: 20rpx 0; border-radius: $radius-chip;
  font-size: 28rpx; color: $text-sub; background: $card-bg; border: 2rpx solid $card-border;
  &.on { color: #fff; background: $sage; border-color: $sage; }
}

.panel { padding: 32rpx; }
.label { display: block; font-size: 26rpx; color: $text-sub; margin-bottom: 14rpx; }
.label.mt { margin-top: 28rpx; }
.code { letter-spacing: 8rpx; text-align: center; font-size: 36rpx; }
.roles { display: flex; gap: 16rpx; }
.mc-btn.full { margin-top: 40rpx; }
.mc-btn.disabled { opacity: 0.6; }
</style>
