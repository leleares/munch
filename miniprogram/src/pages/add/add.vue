<script setup>
import { ref } from "vue";
import { onShow } from "@dcloudio/uni-app";
import { useMenuStore } from "../../stores/menu";
import { useUserStore } from "../../stores/user";
import { API_BASE_URL, API_PREFIX, USE_CLOUD_CONTAINER } from "../../config";
import { getToken, request } from "../../api/request";

const menu = useMenuStore();
const user = useUserStore();

const ICONS = [
  "🍚",
  "🍜",
  "🍲",
  "🥟",
  "🥬",
  "🦐",
  "🐟",
  "🍅",
  "🌶",
  "🍆",
  "🥩",
  "🐔",
  "🥚",
  "🫛",
];

const editId = ref(null);
const name = ref("");
const desc = ref("");
const recipe = ref("");
const remark = ref("");
const groupId = ref(null);
const spice = ref(1);
const iconEmoji = ref("");
const imageUrl = ref("");
const addingGroup = ref(false);
const newGroup = ref("");
// 长按分组进入删除模式：所有分组 tag 右侧出现 × ，点 × 二次确认后删除
const groupEditMode = ref(false);

onShow(async () => {
  if (!user.hasCouple) {
    uni.reLaunch({ url: "/pages/bind/bind" });
    return;
  }
  groupEditMode.value = false; // 每次进页面都退出删除模式，避免误删
  if (!menu.groups.length) await menu.loadAll();

  if (menu.editingId) {
    // 编辑态：从 store 取被编辑的菜灌进表单
    const d = menu.dishById(menu.editingId);
    if (d) {
      editId.value = d.id;
      name.value = d.name;
      desc.value = d.desc;
      recipe.value = d.recipe || "";
      remark.value = d.remark || "";
      groupId.value = d.groupId;
      spice.value = d.spice;
      iconEmoji.value = d.iconEmoji || "";
      imageUrl.value = d.imageUrl || "";
    }
    menu.editingId = null;
  } else {
    resetForm();
  }
});

function resetForm() {
  editId.value = null;
  name.value = "";
  desc.value = "";
  recipe.value = "";
  remark.value = "";
  groupId.value = menu.groups[0] ? menu.groups[0].id : null;
  spice.value = 1;
  iconEmoji.value = "";
  imageUrl.value = "";
  addingGroup.value = false;
  newGroup.value = "";
  groupEditMode.value = false;
}

function pickIcon(i) {
  iconEmoji.value = i;
  imageUrl.value = "";
}

// 上传照片：
//  - 云托管：chooseImage(压缩) → 读成 base64 → callContainer 打 /upload-base64（自动带 X-WX-OPENID）
//  - H5/本地：chooseImage → uploadFile multipart 打 /upload
function choosePhoto() {
  // #ifdef MP-WEIXIN
  if (USE_CLOUD_CONTAINER) {
    uni.chooseImage({
      count: 1,
      sizeType: ["compressed"],
      success: (res) => {
        const filePath = res.tempFilePaths[0];
        const ext = "." + (filePath.split(".").pop() || "jpg").toLowerCase();
        uni.showLoading({ title: "上传中…" });
        uni.getFileSystemManager().readFile({
          filePath,
          encoding: "base64",
          success: async (fr) => {
            try {
              const data = await request("/upload-base64", {
                method: "POST",
                data: { ext, data: fr.data },
              });
              imageUrl.value = data.imageUrl;
              iconEmoji.value = "";
            } catch (e) {
              uni.showToast({ title: e.message || "上传失败", icon: "none" });
            } finally {
              uni.hideLoading();
            }
          },
          fail: () => {
            uni.hideLoading();
            uni.showToast({ title: "读取图片失败", icon: "none" });
          },
        });
      },
    });
    return;
  }
  // #endif
  uni.chooseImage({
    count: 1,
    success: (res) => {
      const filePath = res.tempFilePaths[0];
      uni.showLoading({ title: "上传中…" });
      uni.uploadFile({
        url: API_BASE_URL + API_PREFIX + "/upload",
        filePath,
        name: "file",
        header: { Authorization: "Bearer " + getToken() },
        success: (r) => {
          try {
            const body = JSON.parse(r.data);
            if (body.code === 0) {
              imageUrl.value = body.data.imageUrl;
              iconEmoji.value = "";
            } else uni.showToast({ title: body.msg, icon: "none" });
          } catch (e) {
            uni.showToast({ title: "上传失败", icon: "none" });
          }
        },
        fail: () => uni.showToast({ title: "上传失败", icon: "none" }),
        complete: () => uni.hideLoading(),
      });
    },
  });
}

function enterGroupEdit() {
  groupEditMode.value = true;
  addingGroup.value = false;
  uni.vibrateShort && uni.vibrateShort({ fail: () => {} });
}

function confirmDelGroup(g) {
  uni.showModal({
    title: "删除分组",
    content: `确定删除「${g.name}」？组里的菜会移到其它分组。`,
    confirmText: "删除",
    confirmColor: "#c47a6e",
    success: async ({ confirm }) => {
      if (!confirm) return;
      try {
        const res = await menu.removeGroup(g.id);
        // 删掉的正是表单当前选中的分组，就跟到后端指定的 fallback 上
        if (groupId.value === g.id) {
          groupId.value =
            (res && res.fallbackGroupId) ||
            (menu.groups[0] && menu.groups[0].id) ||
            null;
        }
        uni.showToast({ title: "分组已删除", icon: "none" });
      } catch (e) {
        // 后端会拦「至少保留一个分组」
        uni.showToast({ title: e.message, icon: "none" });
      }
    },
  });
}

async function confirmNewGroup() {
  const n = newGroup.value.trim();
  if (!n) return uni.showToast({ title: "写个分组名", icon: "none" });
  const exist = menu.groups.find((g) => g.name === n);
  if (exist) {
    groupId.value = exist.id;
  } else {
    const g = await menu.addGroup(n);
    groupId.value = g.id;
  }
  addingGroup.value = false;
  newGroup.value = "";
}

async function submit() {
  if (!name.value.trim())
    return uni.showToast({ title: "给它起个名字吧～", icon: "none" });
  const payload = {
    name: name.value.trim(),
    groupId: groupId.value,
    spice: spice.value,
    desc: desc.value.trim(),
    recipe: recipe.value.trim(),
    remark: remark.value.trim(),
    iconEmoji: iconEmoji.value,
    imageUrl: imageUrl.value,
  };
  try {
    if (editId.value) {
      await menu.updateDish(editId.value, payload);
      uni.showToast({ title: "改好啦 ✏️", icon: "none" });
    } else {
      await menu.addDish(payload);
      uni.showToast({ title: "加好啦，去点吧 🎉", icon: "none" });
    }
    resetForm();
    setTimeout(() => uni.switchTab({ url: "/pages/menu/menu" }), 500);
  } catch (e) {
    uni.showToast({ title: e.message, icon: "none" });
  }
}

async function del() {
  uni.showModal({
    title: "删除这道菜",
    content: "删了就不在菜单里了哦",
    success: async ({ confirm }) => {
      if (!confirm) return;
      await menu.removeDish(editId.value);
      uni.showToast({ title: "已删除", icon: "none" });
      resetForm();
      setTimeout(() => uni.switchTab({ url: "/pages/menu/menu" }), 500);
    },
  });
}

const spices = ["不辣", "微辣", "中辣", "重辣"];
</script>

<template>
  <view class="page">
    <text class="h1">{{ editId ? "编辑菜品" : "加一道新菜" }}</text>

    <text class="label">菜名</text>
    <input
      class="mc-input"
      :value="name"
      @input="name = $event.detail.value"
      placeholder="这道菜叫什么？"
    />

    <text class="label">配图</text>
    <view class="photo-row">
      <view
        class="upload"
        :class="{ filled: imageUrl || iconEmoji }"
        @tap="choosePhoto"
      >
        <view
          v-if="imageUrl"
          class="preview"
          :style="`background-image:url(${imageUrl})`"
        />
        <text v-else-if="iconEmoji" class="preview-emoji">{{ iconEmoji }}</text>
        <text v-else class="upload-txt">＋ 上传照片</text>
      </view>
      <view class="icons">
        <text
          v-for="i in ICONS"
          :key="i"
          class="icon"
          :class="{ on: iconEmoji === i }"
          @tap="pickIcon(i)"
          >{{ i }}</text
        >
      </view>
    </view>

    <text class="label">分组</text>
    <view class="chips">
      <view
        v-for="g in menu.groups"
        :key="g.id"
        class="mc-chip grp"
        :class="{ on: groupId === g.id }"
        @tap="groupId = g.id"
        @longpress="enterGroupEdit"
      >
        <text>{{ g.name }}</text>
        <text
          v-if="groupEditMode"
          class="grp-del"
          @tap.stop="confirmDelGroup(g)"
          >×</text
        >
      </view>

      <!-- 删除模式下只留「完成」；平时才显示「＋新建分组」 -->
      <text
        v-if="groupEditMode"
        class="mc-chip done"
        @tap="groupEditMode = false"
        >完成</text
      >
      <template v-else>
        <text
          v-if="!addingGroup"
          class="mc-chip ghost"
          @tap="addingGroup = true"
          >＋新建分组</text
        >
        <view v-else class="newgroup">
          <input
            class="mc-input sm"
            :value="newGroup"
            @input="newGroup = $event.detail.value"
            placeholder="分组名"
            @confirm="confirmNewGroup"
          />
          <text class="ok" @tap="confirmNewGroup">✓</text>
        </view>
      </template>
    </view>
    <text v-if="groupEditMode" class="grp-tip"
      >点 × 删除分组，组里的菜会移到其它分组</text
    >

    <text class="label">辣度</text>
    <view class="chips">
      <text
        v-for="(s, i) in spices"
        :key="i"
        class="mc-chip"
        :class="{ on: spice === i }"
        @tap="spice = i"
        >{{ s }}</text
      >
    </view>

    <text class="label">一句话描述</text>
    <input
      class="mc-input"
      :value="desc"
      @input="desc = $event.detail.value"
      placeholder="你上周夸过好吃的那道 ✨"
    />

    <text class="label">菜谱 / 做法</text>
    <textarea
      class="mc-input area recipe"
      :value="recipe"
      @input="recipe = $event.detail.value"
      placeholder="随手记下做法步骤，做菜时照着来 🍳"
      :maxlength="-1"
    />

    <text class="label">备注</text>
    <textarea
      class="mc-input area"
      :value="remark"
      @input="remark = $event.detail.value"
      placeholder="这道菜的小提醒，比如「她爱吃肥一点的」"
      :maxlength="500"
    />

    <view class="mc-btn full" @tap="submit">{{
      editId ? "保存修改 ✏️" : "加进菜单 🌿"
    }}</view>
    <view v-if="editId" class="mc-btn danger full mt" @tap="del"
      >删除这道菜</view
    >
    <view style="height: 60rpx" />
  </view>
</template>

<style lang="scss" scoped>
.page {
  padding: 32rpx $page-pad;
}
.h1 {
  display: block;
  font-size: 40rpx;
  font-weight: 700;
  color: $text-title;
  margin-bottom: 20rpx;
}
.label {
  display: block;
  font-size: 28rpx;
  color: $text-title;
  font-weight: 700;
  margin: 30rpx 0 14rpx;
}
.photo-row {
  display: flex;
  gap: 20rpx;
}
.upload {
  width: 160rpx;
  height: 160rpx;
  border-radius: 24rpx;
  border: 3rpx dashed $input-border;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  overflow: hidden;
}
/* 选了图标/照片后：实线描边 + 浅主色底，视觉上表示「已选」 */
.upload.filled {
  border-style: solid;
  border-color: $sage-soft-bd;
  background: $sage-soft-bg;
}
.upload-txt {
  font-size: 22rpx;
  color: $text-weak;
}
.preview-emoji {
  font-size: 84rpx;
  line-height: 1;
}
.preview {
  width: 100%;
  height: 100%;
  background-size: cover;
  background-position: center;
}
.icons {
  flex: 1;
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
}
.icon {
  width: 72rpx;
  height: 72rpx;
  line-height: 72rpx;
  text-align: center;
  font-size: 40rpx;
  border-radius: 18rpx;
  background: $card-bg;
  border: 2rpx solid $card-border;
}
.icon.on {
  background: $sage-soft-bg;
  border-color: $sage;
}
.chips {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
  align-items: center;
}
.mc-chip.ghost {
  color: $sage;
  border-style: dashed;
}
/* 分组 tag：删除模式下右侧带 × */
.mc-chip.grp {
  gap: 12rpx;
}
.grp-del {
  color: $danger-text;
  font-size: 32rpx;
  line-height: 1;
  padding: 0 2rpx 4rpx;
}
.mc-chip.on .grp-del {
  color: #fff;
}
.mc-chip.done {
  color: $sage;
  background: $sage-soft-bg;
  border-color: $sage-soft-bd;
}
.grp-tip {
  display: block;
  font-size: 22rpx;
  color: $text-weak;
  margin-top: 14rpx;
}
.newgroup {
  display: flex;
  align-items: center;
  gap: 10rpx;
}
.mc-input.sm {
  width: 200rpx;
  padding: 12rpx 18rpx;
}
/* 多行输入：菜谱/备注 */
.mc-input.area {
  height: 160rpx;
  padding: 20rpx 26rpx;
  line-height: 1.5;
}
.mc-input.area.recipe {
  height: 240rpx;
}
.ok {
  color: $sage;
  font-size: 36rpx;
}
.mc-btn.full {
  width: 100%;
  margin-top: 44rpx;
}
.mc-btn.mt {
  margin-top: 20rpx;
}
</style>
