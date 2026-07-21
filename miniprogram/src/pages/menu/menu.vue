<script setup>
import { ref, computed } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useUserStore } from '../../stores/user'
import { useMenuStore } from '../../stores/menu'
import { useCartStore } from '../../stores/cart'
import { useOrderStore } from '../../stores/order'

const user = useUserStore()
const menu = useMenuStore()
const cart = useCartStore()
const order = useOrderStore()

const drawer = ref(false)
const randomVisible = ref(false)
const randomDish = ref(null)

const greeting = computed(() => `${(user.user && user.user.nickname) || '亲爱的'}，慢慢挑～`)
const cats = computed(() => ['全部', ...menu.groups.map((g) => g.name)])

onShow(async () => {
  // 等 bootstrap 完成，无情侣空间则去绑定页
  if (!user.ready) await user.bootstrap()
  if (!user.hasCouple) {
    uni.reLaunch({ url: '/pages/bind/bind' })
    return
  }
  await menu.loadAll()
})

// ---- 缩略图背景 ----
function thumbStyle(d) {
  if (d.imageUrl) return `background-image:url(${d.imageUrl});background-size:cover;background-position:center;`
  if (d.iconEmoji) return 'background:#eef0e4;'
  return 'background:repeating-linear-gradient(45deg,#e8e6d8,#e8e6d8 12rpx,#dedbc9 12rpx,#dedbc9 24rpx);'
}
function spiceText(n) {
  return ['不辣', '微辣 🌶', '中辣 🌶🌶', '重辣 🌶🌶🌶'][n] || ''
}

// ---- 分类长按：改名/删除 ----
function onGroupLongPress(g) {
  uni.showActionSheet({
    itemList: ['改名', '删除分组'],
    success: ({ tapIndex }) => {
      if (tapIndex === 0) renameGroup(g)
      else deleteGroup(g)
    },
  })
}
function renameGroup(g) {
  uni.showModal({
    title: '分组改名', editable: true, content: g.name,
    success: async ({ confirm, content }) => {
      if (confirm && content && content.trim()) {
        try { await menu.renameGroup(g.id, content.trim()) } catch (e) { toast(e.message) }
      }
    },
  })
}
function deleteGroup(g) {
  uni.showModal({
    title: '删除分组', content: `「${g.name}」里的菜会移到其它分组，确定删除？`,
    success: async ({ confirm }) => {
      if (confirm) {
        try { await menu.removeGroup(g.id) } catch (e) { toast(e.message) }
      }
    },
  })
}

// ---- 菜品长按：编辑（add 是 tab 页，用 editingId 传参 + switchTab）----
function onDishLongPress(d) {
  menu.editingId = d.id
  uni.switchTab({ url: '/pages/add/add' })
}
function goDetail(d) {
  uni.navigateTo({ url: `/pages/detail/detail?id=${d.id}` })
}

// ---- 今天吃什么 ----
function openRandom() {
  if (!menu.dishes.length) return toast('先加几道菜吧～')
  roll()
  randomVisible.value = true
}
function roll() {
  const pool = menu.dishes
  randomDish.value = pool[Math.floor(Math.random() * pool.length)]
}
function pickRandom() {
  if (randomDish.value) cart.inc(randomDish.value.id)
  randomVisible.value = false
  toast(`「${randomDish.value.name}」已加入 🌿`)
}

// ---- 下单 ----
async function placeOrder() {
  if (cart.count === 0) return toast('先点一道菜嘛～')
  uni.showLoading({ title: '正在送往大厨…' })
  try {
    await order.placeOrder(cart.buildOrderPayload(menu))
    cart.clear()
    drawer.value = false
    uni.hideLoading()
    uni.navigateTo({ url: '/pages/done/done' })
  } catch (e) {
    uni.hideLoading()
    toast(e.message)
  }
}

function goChef() {
  uni.navigateTo({ url: '/pages/chef/chef' })
}
function toast(t) { uni.showToast({ title: t, icon: 'none' }) }
</script>

<template>
  <view class="page">
    <!-- 切大厨端 FAB -->
    <view class="fab" @tap="goChef">
      <text class="fab-icon">🍳</text>
      <text class="fab-label">大厨端</text>
    </view>

    <scroll-view scroll-y class="scroll">
      <!-- 问候语 -->
      <view class="greet">
        <text class="greet-title">{{ greeting }}</text>
        <text class="greet-sub">想吃什么都记下来，我一样样做给你 🌿</text>
      </view>

      <!-- 今天吃什么 -->
      <view class="mc-card today" @tap="openRandom">
        <text class="today-icon">🎲</text>
        <view class="today-mid">
          <text class="today-title">今天吃什么</text>
          <text class="today-desc">选择困难就交给天意 ✨</text>
        </view>
        <text class="today-arrow">›</text>
      </view>

      <!-- 分类 chip -->
      <scroll-view scroll-x class="cats" :show-scrollbar="false">
        <text
          v-for="c in cats" :key="c"
          class="mc-chip cat" :class="{ on: menu.cat === c }"
          @tap="menu.setCat(c)"
          @longpress="c !== '全部' && onGroupLongPress(menu.groups.find(g => g.name === c))"
        >{{ c }}</text>
      </scroll-view>

      <text class="hint">长按菜品可编辑 · 长按分组可改名/删除</text>

      <!-- 菜品列表 -->
      <view class="list">
        <view v-for="d in menu.visibleDishes" :key="d.id" class="mc-card dish">
          <view class="thumb" :style="thumbStyle(d)" @tap="goDetail(d)" @longpress="onDishLongPress(d)">
            <text v-if="d.iconEmoji" class="thumb-emoji">{{ d.iconEmoji }}</text>
          </view>
          <view class="dish-mid" @tap="goDetail(d)" @longpress="onDishLongPress(d)">
            <text class="dish-name">{{ d.name }}</text>
            <text class="dish-desc">{{ d.desc }}</text>
            <text class="dish-spice">{{ spiceText(d.spice) }}</text>
          </view>
          <view class="dish-right">
            <view v-if="!cart.qtyOf(d.id)" class="addbtn" @tap="cart.inc(d.id)">＋</view>
            <view v-else class="stepper">
              <text class="step" @tap="cart.dec(d.id)">−</text>
              <text class="qty">{{ cart.qtyOf(d.id) }}</text>
              <text class="step" @tap="cart.inc(d.id)">＋</text>
            </view>
          </view>
        </view>
        <view v-if="!menu.visibleDishes.length" class="empty">这个分类还没有菜，去右下角加一道吧 🌱</view>
      </view>
      <view style="height: 200rpx" />
    </scroll-view>

    <!-- 购物车条 -->
    <view v-if="cart.count" class="cartbar">
      <view class="cartbar-left" @tap="drawer = true">
        <text class="cart-icon">🛒</text>
        <text class="cart-text">已点 {{ cart.count }} 道</text>
      </view>
      <view class="cartbar-btn" @tap="placeOrder">去下单</view>
    </view>

    <!-- 购物车抽屉 -->
    <view v-if="drawer" class="mask" @tap="drawer = false">
      <view class="sheet" @tap.stop>
        <view class="sheet-handle" />
        <view class="sheet-head">
          <text class="sheet-title">我点的菜 · {{ cart.count }} 道</text>
          <text class="sheet-close" @tap="drawer = false">×</text>
        </view>
        <scroll-view scroll-y class="sheet-list">
          <view v-for="d in menu.dishes.filter(x => cart.qtyOf(x.id))" :key="d.id" class="sheet-item">
            <view class="sheet-thumb" :style="thumbStyle(d)">
              <text v-if="d.iconEmoji" class="thumb-emoji sm">{{ d.iconEmoji }}</text>
            </view>
            <view class="sheet-mid">
              <text class="dish-name">{{ d.name }}</text>
              <text class="dish-spice">{{ cart.noteText(d.id, d.spice) }}</text>
            </view>
            <view class="stepper">
              <text class="step" @tap="cart.dec(d.id)">−</text>
              <text class="qty">{{ cart.qtyOf(d.id) }}</text>
              <text class="step" @tap="cart.inc(d.id)">＋</text>
            </view>
          </view>
        </scroll-view>
        <textarea class="mc-input note" v-model="cart.msg" placeholder="给大厨留张小纸条 💌" />
        <view class="mc-btn full" @tap="placeOrder">下单 · 让大厨开火 🍳（{{ cart.count }} 道）</view>
      </view>
    </view>

    <!-- 今天吃什么弹窗 -->
    <view v-if="randomVisible" class="mask center" @tap="randomVisible = false">
      <view class="random" @tap.stop>
        <text class="random-title">🎲 今天就吃</text>
        <view class="random-thumb" :style="thumbStyle(randomDish)">
          <text v-if="randomDish && randomDish.iconEmoji" class="thumb-emoji lg">{{ randomDish.iconEmoji }}</text>
        </view>
        <text class="random-name">{{ randomDish && randomDish.name }}</text>
        <text class="random-desc">{{ randomDish && randomDish.desc }}</text>
        <view class="random-btns">
          <view class="mc-btn ghost flex1" @tap="roll">换一个 🎲</view>
          <view class="mc-btn flex1" @tap="pickRandom">就它啦 🌿</view>
        </view>
      </view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.page { height: 100vh; position: relative; }
.scroll { height: 100%; padding: 0 $page-pad; box-sizing: border-box; }

/* FAB */
.fab {
  position: fixed; top: 20rpx; right: 24rpx; z-index: 50;
  width: 104rpx; height: 104rpx; border-radius: 28rpx;
  background: $sage; box-shadow: $shadow-fab;
  display: flex; flex-direction: column; align-items: center; justify-content: center;
}
.fab-icon { font-size: 36rpx; }
.fab-label { font-size: 18rpx; color: #fff; margin-top: 4rpx; }

/* 问候语 */
.greet { padding: 32rpx 0 20rpx; }
.greet-title { display: block; font-size: 48rpx; font-weight: 700; color: $text-title; }
.greet-sub { display: block; font-size: 26rpx; color: $text-sub; margin-top: 10rpx; }

/* 今天吃什么 */
.today { display: flex; align-items: center; padding: 24rpx 28rpx; margin-bottom: 24rpx; }
.today-icon { font-size: 48rpx; width: 88rpx; height: 88rpx; line-height: 88rpx; text-align: center; background: $sage-soft-bg; border-radius: 24rpx; }
.today-mid { flex: 1; margin-left: 20rpx; }
.today-title { display: block; font-size: 30rpx; font-weight: 700; color: $text-title; }
.today-desc { display: block; font-size: 24rpx; color: $text-sub; margin-top: 6rpx; }
.today-arrow { font-size: 44rpx; color: #a8b090; }

/* 分类 */
.cats { white-space: nowrap; margin-bottom: 12rpx; }
.cat { margin-right: 16rpx; }
.hint { display: block; font-size: 22rpx; color: $text-weak; margin: 4rpx 0 20rpx; }

/* 菜品行 */
.list { display: flex; flex-direction: column; gap: 20rpx; }
.dish { display: flex; align-items: center; padding: $card-pad; }
.thumb { width: 132rpx; height: 132rpx; border-radius: $radius-thumb; flex-shrink: 0; display: flex; align-items: center; justify-content: center; overflow: hidden; }
.thumb-emoji { font-size: 60rpx; }
.thumb-emoji.sm { font-size: 44rpx; }
.thumb-emoji.lg { font-size: 96rpx; }
.dish-mid { flex: 1; margin: 0 20rpx; min-width: 0; }
.dish-name { display: block; font-size: 30rpx; font-weight: 700; color: $text-title; }
.dish-desc { display: block; font-size: 24rpx; color: $text-sub; margin-top: 6rpx; }
.dish-spice { display: block; font-size: 22rpx; color: $text-weak; margin-top: 6rpx; }
.dish-right { flex-shrink: 0; }
.addbtn {
  width: 60rpx; height: 60rpx; border-radius: 18rpx; border: 3rpx solid $sage-soft-bd;
  color: $sage; font-size: 40rpx; display: flex; align-items: center; justify-content: center;
}
.stepper { display: flex; align-items: center; gap: 8rpx; }
.step { width: 56rpx; height: 56rpx; border-radius: 16rpx; background: $sage; color: #fff; font-size: 34rpx; text-align: center; line-height: 56rpx; }
.qty { min-width: 44rpx; text-align: center; font-size: 30rpx; color: $text-main; }
.empty { text-align: center; color: $text-weak; font-size: 26rpx; padding: 80rpx 0; }

/* 购物车条 */
.cartbar {
  position: fixed; left: $page-pad; right: $page-pad; bottom: 24rpx; z-index: 40;
  height: 96rpx; border-radius: $radius-pill; background: $sage; box-shadow: $shadow-fab;
  display: flex; align-items: center; padding: 0 12rpx 0 32rpx;
}
.cartbar-left { flex: 1; display: flex; align-items: center; }
.cart-icon { font-size: 40rpx; }
.cart-text { color: #fff; font-size: 28rpx; margin-left: 16rpx; }
.cartbar-btn { background: #fff; color: $sage-deep; padding: 18rpx 40rpx; border-radius: $radius-pill; font-size: 28rpx; }

/* 遮罩 & 抽屉 */
.mask { position: fixed; inset: 0; background: $mask; z-index: 100; display: flex; align-items: flex-end; }
.mask.center { align-items: center; justify-content: center; }
.sheet { width: 100%; background: $screen-bg; border-radius: 40rpx 40rpx 0 0; padding: 24rpx $page-pad 40rpx; box-sizing: border-box; max-height: 80vh; display: flex; flex-direction: column; }
.sheet-handle { width: 72rpx; height: 8rpx; border-radius: 8rpx; background: #d8d3c4; margin: 0 auto 20rpx; }
.sheet-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 20rpx; }
.sheet-title { font-size: 32rpx; font-weight: 700; color: $text-title; }
.sheet-close { font-size: 48rpx; color: $text-weak; }
.sheet-list { max-height: 44vh; }
.sheet-item { display: flex; align-items: center; padding: 16rpx 0; border-bottom: 2rpx solid $divider; }
.sheet-thumb { width: 88rpx; height: 88rpx; border-radius: 18rpx; flex-shrink: 0; display: flex; align-items: center; justify-content: center; overflow: hidden; }
.sheet-mid { flex: 1; margin: 0 20rpx; }
.note { margin: 24rpx 0; height: 120rpx; }
.mc-btn.full { width: 100%; }

/* 随机弹窗 */
.random { width: 80%; background: $card-bg; border-radius: 40rpx; padding: 48rpx 40rpx; display: flex; flex-direction: column; align-items: center; }
.random-title { font-size: 30rpx; color: $text-sub; }
.random-thumb { width: 200rpx; height: 200rpx; border-radius: 32rpx; margin: 24rpx 0; display: flex; align-items: center; justify-content: center; overflow: hidden; }
.random-name { font-size: 40rpx; font-weight: 700; color: $text-title; }
.random-desc { font-size: 26rpx; color: $text-sub; margin-top: 10rpx; text-align: center; }
.random-btns { display: flex; gap: 20rpx; width: 100%; margin-top: 36rpx; }
.flex1 { flex: 1; }
</style>
