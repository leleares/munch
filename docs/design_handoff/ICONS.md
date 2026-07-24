# 图标资产索引

设计稿里的图标全部是**内联线性 SVG**（`stroke=currentColor`，`stroke-linecap/linejoin=round`）。Claude Code 之前「没有图标」，是因为这些 SVG 只存在于原型里、没被当作资产交付。这里补齐两种格式，任选其一落地。

## 目录
- `assets/icons/*.svg` — 矢量源，`stroke="currentColor"`，颜色由外层 `color` 控制，可无限缩放。**Web / React / Taro / Vue 用这个**。
- `assets/icons-png/*.png` — 144×144 已上色 PNG。**微信小程序 `<image>` 用这个**（小程序对内联 SVG 支持差）。

## 清单
| 图标 | 用途 | 原型描边宽 | PNG 颜色变体 |
|---|---|---|---|
| `tab-menu` | 底 Tab「点菜」：饭碗+热气 | 1.7 | active `#8a9a72` / inactive `#b0ad98` |
| `tab-add` | 底 Tab「加新菜」：嫩芽+加号 | 1.7 | active / inactive 同上 |
| `tab-records` | 底 Tab「记录」：小票 | 1.7 | active / inactive 同上 |
| `fab-pot` | 悬浮按钮·去大厨端：锅 | 1.9 | white（叠在 sage 按钮上） |
| `fab-bowl` | 悬浮按钮·回点菜端：饭碗 | 1.9 | white |
| `chevron` | 「今天吃什么」右侧箭头（静态、粗圆头） | 3.2 | `#a8b090` |
| `loading-pot` | 下单加载：煮锅+三缕热气 | 2.2 | 见下（动画，用 SVG） |

## 用色规则
- Tab：选中 `#8a9a72`，未选中 `#b0ad98`。
- FAB 图标：白色，按钮底 `#8a9a72`。
- 详情/卡片里的强调 line：`#8a9a72`。

## 菜品缩略图
原型里的菜品图是 **emoji**（🫛🥩🥚🍜🥣🥬…）或**条纹占位**，不是图标资产。真实产品用用户上传照片 / 选定 emoji；未设图时用占位（浅底 `#eef0e4` + emoji，或条纹 `repeating-linear-gradient(45deg,#e8e6d8,#e8e6d8 6px,#dedbc9 6px,#dedbc9 12px)`）。加菜表单的图标选择器用系统 emoji 即可，无需图标文件。

## 加载动画（loading-pot.svg）
`loading-pot.svg` 内置 CSS 动画：三缕热气 `steamRise`（错开 0/.5/1s）循环上升，锅身 `bob` 轻微起伏。小程序里无法直接跑内联 SVG 动画，请在小程序端用 `<image>`（静态锅）+ wxss `@keyframes` 重建热气/起伏，或用逐帧 GIF。Web 端可直接 `<img src="loading-pot.svg">` 或内联使用。

## App 图标
- `icon_1024.png` — 方形，微信后台 / 应用商店上传用。
- `icon_1024_rounded.png` — 圆角预览版。
设计：奶油底 `#f1eee4` + 抹茶碗 `#8a9a72` + 爱心热气。
