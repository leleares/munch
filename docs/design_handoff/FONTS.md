# 字体接入指南 · 霞鹜文楷 (LXGW WenKai TC)

设计稿全局字体是 **LXGW WenKai TC（霞鹜文楷 · 繁体字库，同时含大量简体字形）**，标题/问候语/弹窗菜名用 `700`，正文用 `400`。回退顺序：`"LXGW WenKai TC", "Noto Sans SC", system-ui, sans-serif`。

> 许可证：**SIL Open Font License 1.1**，免费商用，可随包分发（附带 OFL.txt 即可）。项目：github.com/lxgw/LxgwWenKai

---

## A. Web / H5 / React / Taro-H5 / Vue

最省事——用现成 webfont npm 包（已切好 woff2、自带 @font-face）：

```bash
npm i lxgw-wenkai-tc-webfont
```
```js
import 'lxgw-wenkai-tc-webfont/style.css';   // 入口引入一次
```
```css
body { font-family: "LXGW WenKai TC", "Noto Sans SC", system-ui, sans-serif; }
```

或 CDN（免安装）：
```html
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/lxgw-wenkai-tc-webfont/style.css">
```

Google Fonts 也有（但国内访问不稳，正式环境别依赖）：
```html
<link href="https://fonts.googleapis.com/css2?family=LXGW+WenKai+TC:wght@400;700&display=swap" rel="stylesheet">
```

---

## B. 微信小程序（重点，坑最多）

**小程序不能用 CSS `@font-face` 加载网络字体**，必须用 `wx.loadFontFace`，且：
- 只支持 **woff / woff2 / ttf**，来源必须是 **https**，域名要加进小程序后台「downloadFile 合法域名」；
- 全局生效要 `global:true`，且建议在 `app.js` 的 `onLaunch` 里调用；页面渲染前字体可能还没到，先用回退字体，加载完成后会自动替换。

```js
// app.js
onLaunch() {
  wx.loadFontFace({
    global: true,
    family: 'LXGW WenKai TC',
    source: 'url("https://你的CDN域名/LXGWWenKaiTC-Regular.woff2")',
    success: () => console.log('font ready'),
  });
  // 需要粗体再加载一个 700 字重文件，family 用另一个名字，如 'LXGW WenKai TC Bold'
}
```
```css
/* wxss */
.page { font-family: 'LXGW WenKai TC', -apple-system, sans-serif; }
.title { font-family: 'LXGW WenKai TC Bold', 'LXGW WenKai TC', sans-serif; }
```

### 强烈建议：字体子集化
完整中文字库单个字重 **10MB+**，直接加载会很慢、影响首屏。做法：
1. 收集项目里所有会出现的汉字（菜名多为用户输入，可给个常用 3500 字 + 界面固定文案）；
2. 用 `fonttools` / `font-spider` / `cn-font-split` 切出子集 woff2（通常能压到几百 KB）；
3. 传到你的 CDN / 对象存储，再用上面的 `wx.loadFontFace` 引。

> 若不想折腾字体加载，也可退而用小程序默认字体 + 我们的配色/圆角/间距，视觉损失有限；但文楷的温馨手写感是这套设计的重要气质，能上尽量上。

---

## C. iOS / Android 原生
把 woff2/ttf 放进工程资源，iOS 在 Info.plist 注册 `UIAppFonts`，Android 放 `res/font/` 或 assets 后按名调用。字体名同样是 `LXGW WenKai TC`。
