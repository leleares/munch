<script setup>
import { onLaunch } from '@dcloudio/uni-app'
import { useUserStore } from './stores/user'

onLaunch(async () => {
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
