<script setup>
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useMenuStore } from '../../stores/menu'
import { useUserStore } from '../../stores/user'
import { API_BASE_URL, API_PREFIX, USE_CLOUD_CONTAINER } from '../../config'
import { getToken } from '../../api/request'

const menu = useMenuStore()
const user = useUserStore()

const ICONS = ['🍚', '🍜', '🍲', '🥘', '🍳', '🥟', '🥬', '🍗', '🦐', '🐟', '🍅', '🌶']

const editId = ref(null)
const name = ref('')
const desc = ref('')
const groupId = ref(null)
const spice = ref(1)
const iconEmoji = ref('')
const imageUrl = ref('')
const addingGroup = ref(false)
const newGroup = ref('')

onShow(async () => {
  if (!user.hasCouple) { uni.reLaunch({ url: '/pages/bind/bind' }); return }
  if (!menu.groups.length) await menu.loadAll()

  if (menu.editingId) {
    // 编辑态：从 store 取被编辑的菜灌进表单
    const d = menu.dishById(menu.editingId)
    if (d) {
      editId.value = d.id
      name.value = d.name
      desc.value = d.desc
      groupId.value = d.groupId
      spice.value = d.spice
      iconEmoji.value = d.iconEmoji || ''
      imageUrl.value = d.imageUrl || ''
    }
    menu.editingId = null
  } else {
    resetForm()
  }
})

function resetForm() {
  editId.value = null
  name.value = ''
  desc.value = ''
  groupId.value = menu.groups[0] ? menu.groups[0].id : null
  spice.value = 1
  iconEmoji.value = ''
  imageUrl.value = ''
  addingGroup.value = false
  newGroup.value = ''
}

function pickIcon(i) {
  iconEmoji.value = i
  imageUrl.value = ''
}

// 上传照片：chooseImage → uploadFile 到 /api/upload
function choosePhoto() {
  // #ifdef MP-WEIXIN
  if (USE_CLOUD_CONTAINER) {
    uni.showToast({ title: '云托管上传待接入 COS，先用图标吧 🌿', icon: 'none' })
    return
  }
  // #endif
  uni.chooseImage({
    count: 1,
    success: (res) => {
      const filePath = res.tempFilePaths[0]
      uni.showLoading({ title: '上传中…' })
      uni.uploadFile({
        url: API_BASE_URL + API_PREFIX + '/upload',
        filePath,
        name: 'file',
        header: { Authorization: 'Bearer ' + getToken() },
        success: (r) => {
          try {
            const body = JSON.parse(r.data)
            if (body.code === 0) {
              imageUrl.value = body.data.imageUrl
              iconEmoji.value = ''
            } else uni.showToast({ title: body.msg, icon: 'none' })
          } catch (e) { uni.showToast({ title: '上传失败', icon: 'none' }) }
        },
        fail: () => uni.showToast({ title: '上传失败', icon: 'none' }),
        complete: () => uni.hideLoading(),
      })
    },
  })
}

async function confirmNewGroup() {
  const n = newGroup.value.trim()
  if (!n) return uni.showToast({ title: '写个分组名', icon: 'none' })
  const exist = menu.groups.find((g) => g.name === n)
  if (exist) { groupId.value = exist.id }
  else {
    const g = await menu.addGroup(n)
    groupId.value = g.id
  }
  addingGroup.value = false
  newGroup.value = ''
}

async function submit() {
  if (!name.value.trim()) return uni.showToast({ title: '给它起个名字吧～', icon: 'none' })
  const payload = {
    name: name.value.trim(),
    groupId: groupId.value,
    spice: spice.value,
    desc: desc.value.trim(),
    iconEmoji: iconEmoji.value,
    imageUrl: imageUrl.value,
  }
  try {
    if (editId.value) {
      await menu.updateDish(editId.value, payload)
      uni.showToast({ title: '改好啦 ✏️', icon: 'none' })
    } else {
      await menu.addDish(payload)
      uni.showToast({ title: '加好啦，去点吧 🎉', icon: 'none' })
    }
    resetForm()
    setTimeout(() => uni.switchTab({ url: '/pages/menu/menu' }), 500)
  } catch (e) {
    uni.showToast({ title: e.message, icon: 'none' })
  }
}

async function del() {
  uni.showModal({
    title: '删除这道菜', content: '删了就不在菜单里了哦',
    success: async ({ confirm }) => {
      if (!confirm) return
      await menu.removeDish(editId.value)
      uni.showToast({ title: '已删除', icon: 'none' })
      resetForm()
      setTimeout(() => uni.switchTab({ url: '/pages/menu/menu' }), 500)
    },
  })
}

const spices = ['不辣', '微辣', '中辣', '重辣']
</script>

<template>
  <view class="page">
    <text class="h1">{{ editId ? '编辑菜品' : '加一道新菜' }}</text>

    <text class="label">菜名</text>
    <input class="mc-input" v-model="name" placeholder="这道菜叫什么？" />

    <text class="label">配图</text>
    <view class="photo-row">
      <view class="upload" @tap="choosePhoto">
        <view v-if="imageUrl" class="preview" :style="`background-image:url(${imageUrl})`" />
        <text v-else class="upload-txt">＋ 上传照片</text>
      </view>
      <view class="icons">
        <text v-for="i in ICONS" :key="i" class="icon" :class="{ on: iconEmoji === i }" @tap="pickIcon(i)">{{ i }}</text>
      </view>
    </view>

    <text class="label">分组</text>
    <view class="chips">
      <text v-for="g in menu.groups" :key="g.id" class="mc-chip" :class="{ on: groupId === g.id }" @tap="groupId = g.id">{{ g.name }}</text>
      <text v-if="!addingGroup" class="mc-chip ghost" @tap="addingGroup = true">＋新建分组</text>
      <view v-else class="newgroup">
        <input class="mc-input sm" v-model="newGroup" placeholder="分组名" @confirm="confirmNewGroup" />
        <text class="ok" @tap="confirmNewGroup">✓</text>
      </view>
    </view>

    <text class="label">辣度</text>
    <view class="chips">
      <text v-for="(s, i) in spices" :key="i" class="mc-chip" :class="{ on: spice === i }" @tap="spice = i">{{ s }}</text>
    </view>

    <text class="label">一句话描述</text>
    <input class="mc-input" v-model="desc" placeholder="你上周夸过好吃的那道 ✨" />

    <view class="mc-btn full" @tap="submit">{{ editId ? '保存修改 ✏️' : '加进菜单 🌿' }}</view>
    <view v-if="editId" class="mc-btn danger full mt" @tap="del">删除这道菜</view>
    <view style="height: 60rpx" />
  </view>
</template>

<style lang="scss" scoped>
.page { padding: 32rpx $page-pad; }
.h1 { display: block; font-size: 40rpx; font-weight: 700; color: $text-title; margin-bottom: 20rpx; }
.label { display: block; font-size: 28rpx; color: $text-title; font-weight: 700; margin: 30rpx 0 14rpx; }
.photo-row { display: flex; gap: 20rpx; }
.upload { width: 160rpx; height: 160rpx; border-radius: 24rpx; border: 3rpx dashed $input-border; display: flex; align-items: center; justify-content: center; flex-shrink: 0; overflow: hidden; }
.upload-txt { font-size: 22rpx; color: $text-weak; }
.preview { width: 100%; height: 100%; background-size: cover; background-position: center; }
.icons { flex: 1; display: flex; flex-wrap: wrap; gap: 12rpx; }
.icon { width: 72rpx; height: 72rpx; line-height: 72rpx; text-align: center; font-size: 40rpx; border-radius: 18rpx; background: $card-bg; border: 2rpx solid $card-border; }
.icon.on { background: $sage-soft-bg; border-color: $sage; }
.chips { display: flex; flex-wrap: wrap; gap: 16rpx; align-items: center; }
.mc-chip.ghost { color: $sage; border-style: dashed; }
.newgroup { display: flex; align-items: center; gap: 10rpx; }
.mc-input.sm { width: 200rpx; padding: 12rpx 18rpx; }
.ok { color: $sage; font-size: 36rpx; }
.mc-btn.full { width: 100%; margin-top: 44rpx; }
.mc-btn.mt { margin-top: 20rpx; }
</style>
