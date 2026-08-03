<script setup>
import { ref, computed, nextTick } from "vue";
import { onShow, onHide } from "@dcloudio/uni-app";
import { useUserStore } from "../../stores/user";
import { useMenuStore } from "../../stores/menu";
import { useCartStore } from "../../stores/cart";
import { useOrderStore } from "../../stores/order";
import BottomSheet from "../../components/BottomSheet/BottomSheet.vue";

const user = useUserStore();
const menu = useMenuStore();
const cart = useCartStore();
const order = useOrderStore();

const drawer = ref(false);

// 转盘状态：rolling=抽奖中，randomId=当前候选菜
const randomVisible = ref(false);
const rolling = ref(false);
const randomId = ref(null);
let rollTimer = null;
const randomDish = computed(() => menu.dishById(randomId.value));

// 飞入购物车：并发飞行块。连点 ＋ 时每次都生成独立的一个，各飞各的、互不打断
const flyers = ref([]);
let flyerSeq = 0;
const flyTimers = new Map(); // flyerId -> timer
const cartBump = ref(false);
let bumpTimer = null;

const greeting = computed(
  () => `${(user.user && user.user.nickname) || "亲爱的"}，慢慢挑～`,
);
const cats = computed(() => ["全部", ...menu.groups.map((g) => g.name)]);

onShow(async () => {
  // 等 bootstrap 完成，无情侣空间则去绑定页
  if (!user.ready) await user.bootstrap();
  if (!user.hasCouple) {
    uni.reLaunch({ url: "/pages/bind/bind" });
    return;
  }
  await menu.loadAll();
});

// ---- 缩略图背景 ----
function thumbStyle(d) {
  if (!d) return "";
  if (d.imageUrl)
    return `background-image:url(${d.imageUrl});background-size:cover;background-position:center;`;
  if (d.iconEmoji) return "background:#eef0e4;";
  return "background:repeating-linear-gradient(45deg,#e8e6d8,#e8e6d8 12rpx,#dedbc9 12rpx,#dedbc9 24rpx);";
}
function spiceText(n) {
  return ["不辣", "微辣 🌶", "中辣 🌶🌶", "重辣 🌶🌶🌶"][n] || "";
}

// ---- 菜品长按：编辑（add 是 tab 页，用 editingId 传参 + switchTab）----
function onDishLongPress(d) {
  menu.editingId = d.id;
  uni.switchTab({ url: "/pages/add/add" });
}
function goDetail(d) {
  uni.navigateTo({ url: `/pages/detail/detail?id=${d.id}` });
}

// ---- 今天吃什么：老虎机式抽奖，每 85ms 换一个候选，跳 13 次后停（≈1.1s）----
function openRandom() {
  if (!menu.dishes.length) return toast("先加几道菜吧～");
  randomVisible.value = true;
  roll();
}
function roll() {
  const pool = menu.dishes;
  if (!pool.length) return;
  const pick = () => pool[Math.floor(Math.random() * pool.length)].id;
  clearInterval(rollTimer); // 防止连点叠加多个定时器
  rolling.value = true;
  randomId.value = pick();
  let n = 0;
  rollTimer = setInterval(() => {
    n++;
    randomId.value = pick();
    if (n >= 13) {
      clearInterval(rollTimer);
      rollTimer = null;
      rolling.value = false; // 停下 → 进入揭晓阶段
    }
  }, 85);
}
function closeRandom() {
  clearInterval(rollTimer);
  rollTimer = null;
  rolling.value = false;
  randomVisible.value = false;
}
function pickRandom() {
  if (rolling.value) return; // 抽奖中不可确认
  const d = randomDish.value;
  if (!d) return;
  cart.inc(d.id);
  closeRandom();
  toast(`「${d.name}」已加入 🌿`);
}

// ---- 加菜：数据立刻 +1，动画并行（沿左上四分之一圆弧飞入购物车）----
async function addDish(dish) {
  cart.inc(dish.id);
  await nextTick(); // 首次加菜时购物车条此刻才渲染出来，等它出现再取坐标
  runFly(dish);
}

function runFly(dish) {
  const q = uni.createSelectorQuery();
  q.select("#cart-fly-target").boundingClientRect();
  q.select("#add-btn-" + dish.id).boundingClientRect();
  q.exec((res) => {
    const t = res && res[0];
    const s = res && res[1];
    if (!t || !s) return; // 拿不到坐标就跳过动画,不影响加购

    const sx = s.left + s.width / 2;
    const sy = s.top + s.height / 2;
    const dx = t.left + t.width / 2 - sx;
    const dy = t.top + t.height / 2 - sy;

    const id = ++flyerSeq;
    flyers.value.push({
      id,
      sx,
      sy,
      x: 0,
      y: 0,
      scale: 1,
      opacity: 1,
      img: dish.imageUrl || "",
      emoji: dish.iconEmoji || "🍽",
    });

    const STEPS = 14;
    const DUR = 700;
    let i = 0;
    const timer = setInterval(() => {
      i++;
      const p = Math.min(i / STEPS, 1);
      const e = p * p; // ease-in 近似
      const th = (e * Math.PI) / 2;
      const f = flyers.value.find((x) => x.id === id);
      if (!f) {
        clearInterval(timer);
        flyTimers.delete(id);
        return;
      }
      f.x = dx * Math.sin(th); // 横向先快后慢
      f.y = dy * (1 - Math.cos(th)); // 纵向先慢后快 → 合成凸向左上的四分之一圆弧
      f.scale = 1 - e * 0.78;
      f.opacity = p >= 1 ? 0.25 : 1;
      if (p >= 1) {
        clearInterval(timer);
        flyTimers.delete(id);
        flyers.value = flyers.value.filter((x) => x.id !== id);
        bumpCart();
      }
    }, DUR / STEPS);
    flyTimers.set(id, timer);
  });
}

function flyerStyle(f) {
  return (
    `left:${f.sx}px;top:${f.sy}px;` +
    `transform:translate(calc(-50% + ${f.x}px), calc(-50% + ${f.y}px)) scale(${f.scale});` +
    `opacity:${f.opacity};` +
    (f.img
      ? `background-image:url(${f.img});background-size:cover;background-position:center;`
      : "")
  );
}

// 购物车图标「接住」回弹 scale 1→1.4→1。
// 连续落地时先摘掉再挂上 class，否则 CSS 动画不会重新播放。
function bumpCart() {
  clearTimeout(bumpTimer);
  cartBump.value = false;
  nextTick(() => {
    cartBump.value = true;
    bumpTimer = setTimeout(() => (cartBump.value = false), 320);
  });
}

// 离开页面清掉全部定时器，避免泄漏
onHide(() => {
  clearInterval(rollTimer);
  rollTimer = null;
  flyTimers.forEach((t) => clearInterval(t));
  flyTimers.clear();
  flyers.value = [];
});

// ---- 下单 ----
async function placeOrder() {
  if (cart.count === 0) return toast("先点一道菜嘛～");
  uni.showLoading({ title: "正在送往大厨…" });
  try {
    await order.placeOrder(cart.buildOrderPayload(menu));
    cart.clear();
    drawer.value = false;
    uni.hideLoading();
    uni.navigateTo({ url: "/pages/done/done" });
  } catch (e) {
    uni.hideLoading();
    toast(e.message);
  }
}

function goChef() {
  uni.navigateTo({ url: "/pages/chef/chef" });
}
function toast(t) {
  uni.showToast({ title: t, icon: "none" });
}
</script>

<template>
  <view class="page">
    <!-- 切大厨端 FAB -->
    <view class="fab" @tap="goChef">
      <image
        class="fab-icon"
        src="/static/icons/fab-pot-white.png"
        mode="aspectFit"
      />
      <text class="fab-label">大厨端</text>
    </view>

    <scroll-view scroll-y class="scroll">
      <view class="scroll-inner">
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
        <image
          class="today-arrow"
          src="/static/icons/chevron.png"
          mode="aspectFit"
        />
      </view>

      <!-- 分类 chip -->
      <scroll-view scroll-x class="cats" :show-scrollbar="false">
        <text
          v-for="c in cats"
          :key="c"
          class="mc-chip cat"
          :class="{ on: menu.cat === c }"
          @tap="menu.setCat(c)"
          >{{ c }}</text
        >
      </scroll-view>
      <!-- 菜品列表 -->
      <view class="list">
        <view v-for="d in menu.visibleDishes" :key="d.id" class="mc-card dish">
          <view
            class="thumb"
            :style="thumbStyle(d)"
            @tap="goDetail(d)"
            @longpress="onDishLongPress(d)"
          >
            <text v-if="d.iconEmoji" class="thumb-emoji">{{
              d.iconEmoji
            }}</text>
          </view>
          <view
            class="dish-mid"
            @tap="goDetail(d)"
            @longpress="onDishLongPress(d)"
          >
            <text class="dish-name">{{ d.name }}</text>
            <text class="dish-desc">{{ d.desc }}</text>
            <text class="dish-spice">{{ spiceText(d.spice) }}</text>
          </view>
          <view class="dish-right">
            <view
              v-if="!cart.qtyOf(d.id)"
              :id="'add-btn-' + d.id"
              class="addbtn"
              @tap="addDish(d)"
              >＋</view
            >
            <view v-else class="stepper">
              <text class="step" @tap="cart.dec(d.id)">−</text>
              <text class="qty">{{ cart.qtyOf(d.id) }}</text>
              <text :id="'add-btn-' + d.id" class="step" @tap="addDish(d)"
                >＋</text
              >
            </view>
          </view>
        </view>
        <view v-if="!menu.visibleDishes.length" class="empty"
          >这个分类还没有菜，去右下角加一道吧 🌱</view
        >
      </view>
      <view style="height: 200rpx" />
      </view>
    </scroll-view>

    <!-- 购物车条 -->
    <view v-if="cart.count" class="cartbar">
      <view class="cartbar-left" @tap="drawer = true">
        <text id="cart-fly-target" class="cart-icon" :class="{ bump: cartBump }"
          >🛒</text
        >
        <text class="cart-text">已点 {{ cart.count }} 道</text>
      </view>
      <view class="cartbar-btn" @tap="placeOrder">去下单</view>
    </view>

    <!-- 购物车抽屉 -->
    <BottomSheet v-model:visible="drawer" title="">
      <template #header>
        <text class="sheet-title">我点的菜 · {{ cart.count }} 道</text>
      </template>
      <scroll-view scroll-y class="cart-list">
        <view
          v-for="d in menu.dishes.filter((x) => cart.qtyOf(x.id))"
          :key="d.id"
          class="cart-item"
        >
          <view class="sheet-thumb" :style="thumbStyle(d)">
            <text v-if="d.iconEmoji" class="thumb-emoji sm">{{
              d.iconEmoji
            }}</text>
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
      <template #footer>
        <textarea
          class="mc-input note"
          :value="cart.msg"
          @input="cart.msg = $event.detail.value"
          placeholder="给大厨留张小纸条 💌"
        />
        <view class="mc-btn full" @tap="placeOrder"
          >下单 · 让大厨开火 🍳（{{ cart.count }} 道）</view
        >
      </template>
    </BottomSheet>

    <!-- 飞入购物车的飞行块（可并发多个） -->
    <view v-for="f in flyers" :key="f.id" class="flyer" :style="flyerStyle(f)">
      <text v-if="!f.img" class="flyer-emoji">{{ f.emoji }}</text>
    </view>

    <!-- 今天吃什么弹窗：rolling（抽奖中）/ reveal（揭晓）两阶段 -->
    <view v-if="randomVisible" class="mask center" @tap="closeRandom">
      <view class="random" @tap.stop>
        <text class="random-title">🎲 今天就吃</text>

        <!-- 抽奖中：骰子旋转 + 背后候选菜模糊虚化 -->
        <template v-if="rolling">
          <view class="roll-box">
            <text class="roll-ghost">{{
              (randomDish && (randomDish.iconEmoji || "🍽")) || "🍽"
            }}</text>
            <text class="roll-dice">🎲</text>
          </view>
          <text class="roll-tip">让我想想…</text>
          <text class="roll-sub">正在为你挑一道 ✨</text>
        </template>

        <!-- 揭晓 -->
        <template v-else>
          <view class="random-thumb reveal" :style="thumbStyle(randomDish)">
            <text
              v-if="randomDish && randomDish.iconEmoji"
              class="thumb-emoji lg"
              >{{ randomDish.iconEmoji }}</text
            >
          </view>
          <text class="random-name reveal">{{
            randomDish && randomDish.name
          }}</text>
          <text class="random-desc">{{ randomDish && randomDish.desc }}</text>
        </template>

        <view class="random-btns">
          <view class="mc-btn ghost flex1" @tap="roll">换一个 🎲</view>
          <view
            class="mc-btn flex1"
            :class="{ disabled: rolling }"
            @tap="pickRandom"
            >就它啦 🌿</view
          >
        </view>
      </view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.page {
  height: 100vh;
  position: relative;
}
.scroll {
  height: 100%;
  box-sizing: border-box;
}
.scroll-inner {
  padding: 0 $page-pad;
}

/* FAB */
.fab {
  position: fixed;
  top: 20rpx;
  right: 24rpx;
  z-index: 50;
  width: 104rpx;
  height: 104rpx;
  border-radius: 28rpx;
  background: $sage;
  box-shadow: $shadow-fab;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}
.fab-icon {
  width: 46rpx;
  height: 46rpx;
}
.fab-label {
  font-size: 18rpx;
  color: #fff;
  margin-top: 4rpx;
}

/* 问候语 */
.greet {
  padding: 32rpx 0 20rpx;
}
.greet-title {
  display: block;
  font-size: 48rpx;
  font-weight: 700;
  color: $text-title;
}
.greet-sub {
  display: block;
  font-size: 26rpx;
  color: $text-sub;
  margin-top: 10rpx;
}

/* 今天吃什么 */
.today {
  display: flex;
  align-items: center;
  padding: 24rpx 28rpx;
  margin-bottom: 24rpx;
}
.today-icon {
  font-size: 48rpx;
  width: 88rpx;
  height: 88rpx;
  line-height: 88rpx;
  text-align: center;
  background: $sage-soft-bg;
  border-radius: 24rpx;
}
.today-mid {
  flex: 1;
  margin-left: 20rpx;
}
.today-title {
  display: block;
  font-size: 30rpx;
  font-weight: 700;
  color: $text-title;
}
.today-desc {
  display: block;
  font-size: 24rpx;
  color: $text-sub;
  margin-top: 6rpx;
}
.today-arrow {
  width: 34rpx;
  height: 34rpx;
  flex-shrink: 0;
}

/* 分类 */
.cats {
  white-space: nowrap;
  margin-bottom: 24rpx;
}
.cats ::-webkit-scrollbar {
  display: none; /* 隐藏横向滚动条 */
}
.cat {
  margin-right: 16rpx;
}
.hint {
  display: block;
  font-size: 22rpx;
  color: $text-weak;
  margin: 4rpx 0 20rpx;
}

/* 菜品行 */
.list {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}
.dish {
  display: flex;
  align-items: center;
  padding: $card-pad;
}
.thumb {
  width: 132rpx;
  height: 132rpx;
  border-radius: $radius-thumb;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}
.thumb-emoji {
  font-size: 60rpx;
}
.thumb-emoji.sm {
  font-size: 44rpx;
}
.thumb-emoji.lg {
  font-size: 96rpx;
}
.dish-mid {
  flex: 1;
  margin: 0 20rpx;
  min-width: 0;
}
.dish-name {
  display: block;
  font-size: 30rpx;
  font-weight: 700;
  color: $text-title;
}
.dish-desc {
  display: block;
  font-size: 24rpx;
  color: $text-sub;
  margin-top: 6rpx;
}
.dish-spice {
  display: block;
  font-size: 22rpx;
  color: $text-weak;
  margin-top: 6rpx;
}
.dish-right {
  flex-shrink: 0;
}
.addbtn {
  width: 60rpx;
  height: 60rpx;
  border-radius: 18rpx;
  border: 3rpx solid $sage-soft-bd;
  color: $sage;
  font-size: 40rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}
.stepper {
  display: flex;
  align-items: center;
  gap: 8rpx;
}
.step {
  width: 56rpx;
  height: 56rpx;
  border-radius: 16rpx;
  background: $sage;
  color: #fff;
  font-size: 34rpx;
  text-align: center;
  line-height: 56rpx;
}
.qty {
  min-width: 44rpx;
  text-align: center;
  font-size: 30rpx;
  color: $text-main;
}
.empty {
  text-align: center;
  color: $text-weak;
  font-size: 26rpx;
  padding: 80rpx 0;
}

/* 购物车条 */
.cartbar {
  position: fixed;
  left: $page-pad;
  right: $page-pad;
  bottom: 24rpx;
  z-index: 40;
  height: 96rpx;
  border-radius: $radius-pill;
  background: $sage;
  box-shadow: $shadow-fab;
  display: flex;
  align-items: center;
  padding: 0 12rpx 0 32rpx;
}
.cartbar-left {
  flex: 1;
  display: flex;
  align-items: center;
}
.cart-icon {
  font-size: 40rpx;
  display: inline-block;
}
.cart-icon.bump {
  animation: cart-bump 0.32s ease-out;
}
.cart-text {
  color: #fff;
  font-size: 28rpx;
  margin-left: 16rpx;
}
.cartbar-btn {
  background: #fff;
  color: $sage-deep;
  padding: 18rpx 40rpx;
  border-radius: $radius-pill;
  font-size: 28rpx;
}

/* 购物车抽屉内容 */
.cart-list {
  max-height: 50vh;
}
.cart-item {
  display: flex;
  align-items: center;
  padding: 16rpx 0;
  border-bottom: 2rpx solid $divider;
}
.sheet-thumb {
  width: 88rpx;
  height: 88rpx;
  border-radius: 18rpx;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}
.sheet-mid {
  flex: 1;
  margin: 0 20rpx;
}
.sheet-title {
  font-size: 32rpx;
  font-weight: 700;
  color: $text-title;
}
.note {
  margin-bottom: 24rpx;
  height: 120rpx;
  padding: 22rpx 26rpx;
  line-height: 1.5;
}
.mc-btn.full {
  width: 100%;
}

/* 随机弹窗 */
.mask {
  position: fixed;
  inset: 0;
  background: $mask;
  z-index: 100;
}
.mask.center {
  display: flex;
  align-items: center;
  justify-content: center;
}
.random {
  width: 80%;
  background: $card-bg;
  border-radius: 40rpx;
  padding: 48rpx 40rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  animation: pop 0.3s ease-out;
}
.random-title {
  font-size: 30rpx;
  color: $text-sub;
}
.random-thumb {
  width: 200rpx;
  height: 200rpx;
  border-radius: 32rpx;
  margin: 24rpx 0;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}
.random-name {
  font-size: 40rpx;
  font-weight: 700;
  color: $text-title;
}
.random-desc {
  font-size: 26rpx;
  color: $text-sub;
  margin-top: 10rpx;
  text-align: center;
}
.random-btns {
  display: flex;
  gap: 20rpx;
  width: 100%;
  margin-top: 36rpx;
}
.flex1 {
  flex: 1;
}
.mc-btn.disabled {
  opacity: 0.45;
}

/* 转盘·抽奖中 */
.roll-box {
  width: 220rpx;
  height: 220rpx;
  border-radius: 36rpx;
  margin: 32rpx auto 0;
  background: $sage-soft-bg;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  position: relative;
}
.roll-ghost {
  position: absolute;
  font-size: 88rpx;
  opacity: 0.3;
  filter: blur(10rpx);
}
.roll-dice {
  font-size: 92rpx;
  animation: spin 0.6s linear infinite;
}
.roll-tip {
  font-size: 44rpx;
  font-weight: 700;
  color: $sage;
  margin-top: 24rpx;
  animation: pulse 1s ease-in-out infinite;
}
.roll-sub {
  font-size: 24rpx;
  color: $text-weak;
  margin-top: 10rpx;
}

/* 揭晓入场 */
.reveal {
  animation: reveal 0.45s ease-out;
}

/* 飞入购物车的飞行块 */
.flyer {
  position: fixed;
  z-index: 999;
  pointer-events: none;
  width: 72rpx;
  height: 72rpx;
  border-radius: 22rpx;
  background: $sage-soft-bg;
  box-shadow: 0 16rpx 36rpx rgba(80, 90, 60, 0.32);
  display: flex;
  align-items: center;
  justify-content: center;
}
.flyer-emoji {
  font-size: 48rpx;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
@keyframes pulse {
  0%,
  100% {
    opacity: 0.5;
    transform: scale(0.97);
  }
  50% {
    opacity: 1;
    transform: scale(1.03);
  }
}
@keyframes reveal {
  0% {
    transform: scale(0.55) rotate(-7deg);
    opacity: 0;
  }
  60% {
    transform: scale(1.08) rotate(3deg);
  }
  100% {
    transform: scale(1) rotate(0);
    opacity: 1;
  }
}
@keyframes pop {
  0% {
    transform: scale(0.7);
    opacity: 0;
  }
  100% {
    transform: scale(1);
    opacity: 1;
  }
}
@keyframes cart-bump {
  0% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.4);
  }
  100% {
    transform: scale(1);
  }
}
</style>
