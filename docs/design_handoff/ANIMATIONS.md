# 补充交接文档 · 两个动画的精确规格

> 针对两处实现与设计稿不一致：①「今天吃什么」随机转盘动画 + 「换一个」交互；②点菜品「＋」的抛物线飞入购物车动画。以下给出**逐帧、可直接照抄的**规格，源实现见 `小食记.dc.html`（`roll()` 与 `flyToCart()`）。

---

## 一、「今天吃什么」随机转盘

### 交互流程
1. 首页「今天吃什么」卡片被点 → 打开居中弹窗（遮罩 `rgba(50,54,40,.4)` + `blur(3px)`；弹窗容器入场 `pop .3s`）。
2. 弹窗有两个阶段，用一个 `rolling` 布尔切换：
   - **rolling = true（抽奖中）**：显示一个旋转的 🎲 + 背后模糊的候选菜 emoji + 文案「让我想想…」。
   - **rolling = false（揭晓）**：显示最终菜品（缩略图 / 菜名 / 描述）+ 两个按钮「换一个 🎲」「就它啦 🌿」。
3. 「就它啦」= 把当前菜加入购物车并关闭；「换一个」= **重新跑一次抽奖动画**（不是瞬间换一个静态结果——这是最容易做错的点）。

### 抽奖动画怎么做的（核心）
不是 CSS 关键帧滚动，而是**用定时器快速切换"当前选中菜"的 id**，制造老虎机式跳动，然后停下：

```js
roll() {
  const pool = this.state.dishes;               // 全部菜
  const pick = () => pool[Math.floor(Math.random() * pool.length)].id;
  clearInterval(this._rl);                       // 防止重复点叠加多个定时器
  this.setState({ showRandom: true, rolling: true, randomId: pick() });
  let n = 0;
  this._rl = setInterval(() => {
    n++;
    this.setState({ randomId: pick() });         // 每 85ms 换一个候选 → 视觉上在"跳动"
    if (n >= 13) {                               // 跳 13 次后停 (~1.1s)
      clearInterval(this._rl);
      this.setState({ rolling: false });         // 停下 → 触发揭晓
    }
  }, 85);
}
```
- **节奏**：`85ms × 13 次 ≈ 1.1 秒`。想要"先快后慢"的减速手感，可把固定 85ms 改成递增间隔（如 60→180ms）用递归 `setTimeout` 实现；当前设计是匀速跳动，够用。
- 「换一个」按钮直接**再调 `roll()`**，完整重播抽奖，不要只 `randomId = pick()`。
- 关闭弹窗 / 接受结果时都要 `clearInterval(this._rl)` 并把 `rolling` 复位，避免定时器泄漏。
- 抽奖进行中（`rolling===true`）时「就它啦」应无效（防止抽到一半确认）。

### 两阶段的视觉（照抄）
**rolling 阶段：**
```html
<div style="width:110px;height:110px;border-radius:18px;margin:16px auto;background:#eef0e4;
            display:flex;align-items:center;justify-content:center;overflow:hidden;position:relative">
  <!-- 背后当前候选菜 emoji，模糊虚化 -->
  <span style="position:absolute;font-size:44px;opacity:.3;filter:blur(5px)">{候选菜 emoji}</span>
  <!-- 前景骰子旋转 -->
  <span style="font-size:46px;animation:spin .6s linear infinite">🎲</span>
</div>
<div style="font-family:'LXGW WenKai TC',serif;font-weight:700;font-size:22px;color:#8a9a72;
            animation:pulse 1s ease-in-out infinite">让我想想…</div>
<div style="font-size:12px;color:#b0ad98;margin-top:5px">正在为你挑一道 ✨</div>
```
**揭晓阶段：** 缩略图和菜名都套 `reveal .45s ease-out` 入场；下面两个按钮（chip「换一个 🎲」+ primary「就它啦 🌿」）。

### 用到的 keyframes
```css
@keyframes spin  { to { transform: rotate(360deg); } }
@keyframes pulse { 0%,100%{ opacity:.5; transform:scale(.97);} 50%{ opacity:1; transform:scale(1.03);} }
@keyframes reveal{ 0%{ transform:scale(.55) rotate(-7deg); opacity:0;}
                   60%{ transform:scale(1.08) rotate(3deg);}
                   100%{ transform:scale(1) rotate(0); opacity:1;} }
@keyframes pop   { 0%{ transform:scale(.7); opacity:0;} 100%{ transform:scale(1); opacity:1;} }
```

### 小程序落地提示
- `setInterval` + `setData`（把 `randomId`/`rolling` 塞进 data）逻辑完全一致。
- `spin/pulse/reveal/pop` 用 wxss `@keyframes` 即可；emoji 用 `text` 组件。
- 每次 `setData` 只改 `randomId` 一个字段，性能没问题。

---

## 二、点菜「＋」→ 抛物线飞入购物车

### 交互
在菜单列表点某道菜右侧的 ＋（未加入时的空心 ＋，或已加入时步进器的 ＋）：
1. 立即 `cart[id]++`（数据先加，别等动画）。
2. 同时**克隆一个小方块**（菜的图标/图片），从被点按钮的位置，沿**左上四分之一圆弧**飞到底部购物车图标；
3. 到达后小方块消失，购物车图标 `scale 1→1.4→1` 弹一下作为"接住"反馈。

> 弧线形状要点（之前反复确认过）：是**左上四分之一圆**——元素先大致横向移动、再向下拐进购物车，弧线鼓在起点→终点连线的**上方**；不是向右下方鼓出，也不是先往上抛。

### 动画怎么写的（核心）
用 Web Animations API，按四分之一圆参数方程**采样 11 个关键帧**：

```js
flyToCart(e, dish) {
  const el = e.currentTarget;
  const src = el.getBoundingClientRect();
  // 购物车图标上挂了 id="cart-fly-target"
  const target = document.querySelector('#cart-fly-target');   // 小程序里改成节点查询，见下
  if (!target) return;
  const t = target.getBoundingClientRect();

  const sx = src.left + src.width / 2,  sy = src.top + src.height / 2;   // 起点中心
  const ex = t.left  + t.width  / 2,    ey = t.top  + t.height  / 2;     // 终点中心
  const dx = ex - sx, dy = ey - sy;

  // 克隆的飞行元素：固定定位在起点，内容是菜的图片或 emoji
  const fly = document.createElement('div');
  fly.style.cssText = `position:fixed;z-index:99999;pointer-events:none;
    left:${sx}px;top:${sy}px;width:36px;height:36px;border-radius:11px;
    display:flex;align-items:center;justify-content:center;font-size:24px;
    box-shadow:0 8px 18px rgba(80,90,60,.32);background:#eef0e4;`;
  if (dish.img) fly.style.background = `url(${dish.img}) center/cover`;
  else fly.textContent = dish.icon || dish.emoji || '🍽';
  document.body.appendChild(fly);

  // 关键：沿四分之一圆采样。θ 从 0→90°
  //   x = dx * sin(θ)      横向分量先快后慢
  //   y = dy * (1 - cos θ) 纵向分量先慢后快  → 合成"左上四分之一圆弧"
  const frames = [];
  for (let i = 0; i <= 10; i++) {
    const p = i / 10, th = p * Math.PI / 2;
    const x = dx * Math.sin(th);
    const y = dy * (1 - Math.cos(th));
    frames.push({
      transform: `translate(calc(-50% + ${x}px), calc(-50% + ${y}px)) scale(${1 - p * 0.78})`,
      opacity: i === 10 ? 0.25 : 1,     // 落地瞬间略微淡出
    });
  }

  fly.animate(frames, { duration: 700, easing: 'ease-in' }).onfinish = () => {
    fly.remove();
    // 购物车"接住"回弹
    target.animate(
      [{ transform:'scale(1)' }, { transform:'scale(1.4)' }, { transform:'scale(1)' }],
      { duration: 320, easing: 'ease-out' }
    );
  };
}
```

### 参数一览（照抄这些数值）
| 项 | 值 |
|---|---|
| 飞行块尺寸 | 36×36，圆角 11，`box-shadow:0 8px 18px rgba(80,90,60,.32)`，底色 `#eef0e4` |
| 内容 | 有图用图 `url(...) center/cover`，否则用菜的 emoji/图标 |
| 关键帧数 | 11（i=0..10） |
| 弧线公式 | `x = dx·sin θ`，`y = dy·(1−cos θ)`，θ:0→π/2 |
| 缩放 | `scale(1 − p·0.78)`：从 1 缩到约 0.22 |
| 时长 / 缓动 | 700ms / `ease-in` |
| 落地淡出 | 最后一帧 `opacity .25` |
| 购物车回弹 | `scale 1→1.4→1`，320ms，`ease-out` |
| 数据变更 | 点击即刻 `cart[id]++`，与动画并行（动画纯视觉，不阻塞加购） |

### 为什么是这两条公式
`sin θ` 让**横向位移一开始快**（θ 小时 sin 增长快），`1−cos θ` 让**纵向位移一开始慢、后面快**（像自由下落）。两者合成的轨迹正是一段**凸向左上**的四分之一圆弧——先横向滑出、末段俯冲进购物车。若发现弧线方向反了（鼓向右下），就是把 x/y 的两个公式对调了，交换回来即可。

### 小程序落地提示
- 小程序没有 `document.createElement` / `getBoundingClientRect`：
  - 用 `wx.createSelectorQuery().select('#cart-fly-target').boundingClientRect()` 拿起点/终点坐标；
  - 飞行元素做成页面里一个**预置的 `position:fixed` 节点**（或用 `movable-view`），通过 `setData` 改它的 `left/top/transform` + `wx.createAnimation` 或直接 CSS transition 驱动；
  - 由于逐帧 11 关键帧在小程序不方便，可退一步用 **CSS 三次贝塞尔近似**：给该节点一个从起点到终点的 transition，配合一个"控制点在左上"的路径感——但最贴近的是用 `wx.createAnimation` 分两段（先横移到中途点、再下落到终点）模拟弧线。
  - 购物车回弹用 wxss `@keyframes` `cart-bump`（`scale 1→1.4→1`）加类名即可。
- 关键是**保住手感三要素**：左上弧线、末段缩小+淡出、落点回弹。

---

## 三、验收 checklist
- [ ] 「换一个」会**重播**抽奖跳动动画，而不是瞬间静态换一道。
- [ ] 抽奖有 rolling / 揭晓两阶段，rolling 时骰子转、文字脉冲，揭晓有 `reveal` 弹入。
- [ ] 抽奖中「就它啦」不可确认。
- [ ] ＋ 点击后购物车数字立刻 +1（不等动画）。
- [ ] 飞行块沿**左上四分之一圆弧**运动，末段缩小并淡出。
- [ ] 落点购物车图标有 `scale 1→1.4→1` 回弹。
- [ ] 快速连点 ＋ 不会卡顿 / 定时器泄漏（转盘每次先 `clearInterval`）。
