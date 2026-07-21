<script setup>
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { useMenuStore } from '../../stores/menu'
import { useCartStore } from '../../stores/cart'

const menu = useMenuStore()
const cart = useCartStore()

const dishId = ref(null)
const spice = ref(0)
const forbid = ref('')

const dish = computed(() => menu.dishById(dishId.value))

onLoad((q) => {
  dishId.value = Number(q.id)
  const d = menu.dishById(dishId.value)
  const n = cart.notes[dishId.value] || {}
  spice.value = n.spice != null ? n.spice : d ? d.spice : 0
  forbid.value = n.forbid || ''
})

const spices = ['不辣', '微辣', '中辣', '重辣']

function thumbStyle(d) {
  if (!d) return ''
  if (d.imageUrl) return `background-image:url(${d.imageUrl});background-size:cover;background-position:center;`
  if (d.iconEmoji) return 'background:#eef0e4;'
  return 'background:repeating-linear-gradient(45deg,#e8e6d8,#e8e6d8 12rpx,#dedbc9 12rpx,#dedbc9 24rpx);'
}

function addToCart() {
  cart.setNote(dishId.value, { spice: spice.value, forbid: forbid.value })
  cart.inc(dishId.value)
  uni.showToast({ title: '加好啦 🌿', icon: 'none' })
  setTimeout(() => uni.navigateBack(), 400)
}
</script>

<template>
  <view class="page" v-if="dish">
    <view class="hero" :style="thumbStyle(dish)">
      <text v-if="dish.iconEmoji" class="hero-emoji">{{ dish.iconEmoji }}</text>
    </view>

    <view class="body">
      <text class="name">{{ dish.name }}</text>
      <text class="desc">{{ dish.desc }}</text>

      <text class="label">辣度</text>
      <view class="chips">
        <text v-for="(s, i) in spices" :key="i" class="mc-chip" :class="{ on: spice === i }" @tap="spice = i">{{ s }}</text>
      </view>

      <text class="label">忌口 / 悄悄话</text>
      <textarea class="mc-input area" v-model="forbid" placeholder="不要香菜、少放盐、多给点辣…写给大厨看 🌿" />
    </view>

    <view class="bar">
      <view class="mc-btn full" @tap="addToCart">加入这道菜 🌿</view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.page { min-height: 100vh; display: flex; flex-direction: column; }
.hero { height: 380rpx; display: flex; align-items: center; justify-content: center; }
.hero-emoji { font-size: 140rpx; }
.body { flex: 1; padding: 36rpx $page-pad; }
.name { display: block; font-size: 44rpx; font-weight: 700; color: $text-title; }
.desc { display: block; font-size: 26rpx; color: $text-sub; margin-top: 12rpx; }
.label { display: block; font-size: 28rpx; color: $text-title; font-weight: 700; margin: 36rpx 0 16rpx; }
.chips { display: flex; gap: 16rpx; flex-wrap: wrap; }
.area { height: 180rpx; }
.bar { padding: 20rpx $page-pad calc(20rpx + env(safe-area-inset-bottom)); }
.mc-btn.full { width: 100%; }
</style>
