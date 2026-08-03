<script setup>
import { ref, watch, nextTick } from "vue";

const props = defineProps({
  visible: { type: Boolean, default: false },
  title: { type: String, default: "" },
  maxHeight: { type: String, default: "80vh" }, // 抽屉最大高度
});

const emit = defineEmits(["update:visible", "close"]);

// 动画状态：null(未挂载) | 'entering' | 'visible' | 'leaving'
const animState = ref(null);
const startY = ref(0);
const currentY = ref(0);
const dragging = ref(false);

watch(
  () => props.visible,
  (val) => {
    if (val) {
      animState.value = "entering";
      nextTick(() => {
        setTimeout(() => (animState.value = "visible"), 10);
      });
    } else if (animState.value) {
      animState.value = "leaving";
      setTimeout(() => (animState.value = null), 300); // 等动画完再卸载
    }
  },
  { immediate: true }
);

function close() {
  emit("update:visible", false);
  emit("close");
}

// 遮罩点击关闭
function onMaskTap() {
  close();
}

// 下拉关闭手势
function onHandleTouchStart(e) {
  startY.value = e.touches[0].clientY;
  currentY.value = 0;
  dragging.value = true;
}
function onHandleTouchMove(e) {
  if (!dragging.value) return;
  const delta = e.touches[0].clientY - startY.value;
  if (delta > 0) {
    // 只允许向下拖
    currentY.value = delta;
    e.preventDefault(); // 阻止页面滚动
  }
}
function onHandleTouchEnd() {
  if (!dragging.value) return;
  dragging.value = false;
  if (currentY.value > 120) {
    // 下拉超过 120px 则关闭
    close();
  }
  currentY.value = 0;
}

function sheetStyle() {
  if (dragging.value && currentY.value > 0) {
    return `transform: translateY(${currentY.value}px); transition: none;`;
  }
  return "";
}
</script>

<template>
  <view v-if="animState" class="bottom-sheet-mask" :class="animState" @tap="onMaskTap">
    <view
      class="bottom-sheet"
      :class="animState"
      :style="sheetStyle()"
      @tap.stop
      catchtouchmove
    >
      <!-- 顶部 handle，可下拉关闭 -->
      <view
        class="sheet-handle"
        @touchstart="onHandleTouchStart"
        @touchmove="onHandleTouchMove"
        @touchend="onHandleTouchEnd"
      />
      <!-- 标题栏 -->
      <view v-if="title || $slots.header" class="sheet-head">
        <slot name="header">
          <text class="sheet-title">{{ title }}</text>
        </slot>
        <text class="sheet-close" @tap="close">×</text>
      </view>
      <!-- 内容插槽 -->
      <view class="sheet-body" :style="{ maxHeight }">
        <slot />
      </view>
      <!-- 底部操作区插槽 -->
      <view v-if="$slots.footer" class="sheet-footer">
        <slot name="footer" />
      </view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
/* 遮罩 */
.bottom-sheet-mask {
  position: fixed;
  inset: 0;
  background: $mask;
  z-index: 100;
  display: flex;
  align-items: flex-end;
  opacity: 0;
  transition: opacity 0.3s ease-out;
}
.bottom-sheet-mask.entering,
.bottom-sheet-mask.visible {
  opacity: 1;
}
.bottom-sheet-mask.leaving {
  opacity: 0;
}

/* 抽屉主体 */
.bottom-sheet {
  width: 100%;
  background: $screen-bg;
  border-radius: 40rpx 40rpx 0 0;
  padding: 24rpx $page-pad 40rpx;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  transform: translateY(100%);
  transition: transform 0.3s ease-out;
}
.bottom-sheet.entering {
  transform: translateY(100%);
}
.bottom-sheet.visible {
  transform: translateY(0);
}
.bottom-sheet.leaving {
  transform: translateY(100%);
}

/* 顶部 handle */
.sheet-handle {
  width: 72rpx;
  height: 8rpx;
  border-radius: 8rpx;
  background: #d8d3c4;
  margin: 0 auto 20rpx;
  cursor: grab;
}

/* 标题栏 */
.sheet-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20rpx;
}
.sheet-title {
  font-size: 32rpx;
  font-weight: 700;
  color: $text-title;
}
.sheet-close {
  font-size: 48rpx;
  color: $text-weak;
  line-height: 1;
  padding: 8rpx;
}

/* 内容区 */
.sheet-body {
  flex: 1;
  overflow-y: auto;
  min-height: 0;
}

/* 底部操作区 */
.sheet-footer {
  margin-top: 24rpx;
  flex-shrink: 0;
}
</style>
