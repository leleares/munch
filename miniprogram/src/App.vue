<script setup>
import { onLaunch } from '@dcloudio/uni-app'
import { useUserStore } from './stores/user'
import { FONT_FAMILY, FONT_URL } from './config'

// 加载霞鹜文楷（设计稿指定字体）。
// 小程序无法用 CSS @font-face 加载网络字体，必须用 loadFontFace + global:true。
// 字体到达前页面先用回退字体渲染，加载完成后会自动替换，不阻塞首屏。
function loadDesignFont() {
  uni.loadFontFace({
    global: true,
    family: FONT_FAMILY,
    source: `url("${FONT_URL}")`,
    success: () => console.log('🐶 字体已加载', FONT_FAMILY),
    fail: (e) => console.warn('🐶 字体加载失败（将回退系统字体）', e),
  })
}

onLaunch(async () => {
  loadDesignFont()
  // 启动即静默登录：拿 openid / 恢复本地 token，决定进点菜端还是绑定页
  const user = useUserStore()
  await user.bootstrap()
})
</script>

<template>
  <!-- uni-app 约定：App.vue 不写页面结构，仅放全局逻辑与全局样式 -->
</template>

<style lang="scss">
/* 全局基础样式（设计 token 见 uni.scss / styles/tokens.scss） */
@import './styles/common.scss';

page {
  background: $screen-bg;
  color: $text-main;
  font-family: 'LXGW WenKai TC', 'Noto Sans SC', system-ui, sans-serif;
  font-size: 28rpx;
}
</style>
