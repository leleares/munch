# 小食记 · 部署操作手册

目标形态：**微信云托管跑 Go 后端 + 小程序前端通过 `callContainer` 调用（免备案）**。

> 全文按「照着做」编排。每一步都标了**做完怎么验证**，出问题时能立刻定位在哪一环。

---

## 零、部署链路全景

```
微信开发者工具                微信云托管                     
┌──────────────┐          ┌────────────────────────────┐
│ 小程序前端    │          │  服务 munch-server          │
│ uni-app 编译  │          │  ├ Go + Gin 容器（监听 80） │
│ 产物         │          │  └ 环境变量 JWT_SECRET 等   │
└──────┬───────┘          │            ↕                │
       │ wx.cloud         │  内置 MySQL 实例            │
       │ .callContainer   │  （平台注入 MYSQL_* 变量）   │
       │ 平台自动注入      │                            │
       │ X-WX-OPENID  ────┼──→ 中间件据此识别用户        │
       └──────────────────┘                            
                          └────────────────────────────┘
```

关键点：**走 `callContainer` 不需要备案域名**，平台在请求头注入 `X-WX-OPENID`，后端中间件直接用它认人，无需 code2session。

---

## 一、部署前检查清单

上线前逐条确认，尤其是前两条安全项。

- [ ] **`ALLOW_DEV_LOGIN` 绝对不要配到线上**。它是本地免微信环境的后门，开启后任何人打 `/api/login` 传个 openid 就能伪造成你或女朋友的身份。线上不配这个变量即为关闭（默认 false）。
- [ ] **`JWT_SECRET` 换成随机长字符串**。别用默认值。生成：`openssl rand -base64 32`
- [ ] `miniprogram/src/manifest.json` 的 `mp-weixin.appid` 已填 `wxd7e8580374c09508` ✅（已完成）
- [ ] 小程序服务类目已选**工具 / 生活服务**类，**不要选餐饮服务**（会要食品经营许可证）
- [ ] 本地 `docker compose up -d mysql` + `go run ./cmd/api` 能跑通全流程

---

## 二、开通云托管 & 建环境

1. 微信公众平台（小程序后台）→ **开发 → 云开发/云托管** → 开通**微信云托管**。
2. 新建环境，**记下「环境 ID」**（形如 `prod-xxxxxx`）→ 前端 `CLOUD_ENV` 要用。
3. 环境内创建 **MySQL 实例**（云托管内置的即可，选最小规格）。

### ⚠️ 这一步最容易漏：要手动建库

平台只给你一个 MySQL **实例**，不会自动建 `munch` 这个**数据库**。代码里的 `AutoMigrate` 只建**表**、不建**库**。

在云托管 MySQL 控制台执行一次：

```sql
CREATE DATABASE IF NOT EXISTS munch
  DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;
```

**验证**：控制台 `SHOW DATABASES;` 能看到 `munch`。

---

## 三、部署后端服务

1. 云托管 → **新建服务**，服务名填 **`munch-server`**（要和前端 `CLOUD_SERVICE` 一致）。
2. 部署方式选「**代码包上传**」或关联 Git 仓库。
   - **Dockerfile 路径**：`server/Dockerfile`
   - **构建目录 / 上下文**：`server`（重要，Dockerfile 里 `COPY assets` 是相对 server 目录的）
3. **监听端口填 `80`**（Dockerfile 里 `ENV PORT=80`，代码默认也是 80）。

### 环境变量

| 变量 | 值 | 说明 |
|---|---|---|
| `JWT_SECRET` | 随机长串 | **必填**，别用默认值 |
| `DB_NAME` | `munch` | 上一步建的库名 |
| `STORAGE_DRIVER` | `local` | 图片存储，见第六节的限制说明 |
| `STATIC_DIR` | `/app/data/uploads` | 容器内上传目录 |
| `PUBLIC_BASE_URL` | 服务公网域名 | 拼图片可访问地址用，没开公网可先留空 |
| ~~`ALLOW_DEV_LOGIN`~~ | **不要配** | 配了就是安全漏洞 |

> `MYSQL_ADDRESS` / `MYSQL_USERNAME` / `MYSQL_PASSWORD` 由平台自动注入，**不用你填**，`config.go` 已兼容读取。

4. 点击部署，等待构建完成。

**验证**：
- 服务日志出现 `[db] connected & migrated` → 数据库连通且建表成功
- 日志出现 `[munch] listening on :80`
- 若日志报连不上数据库 → 多半是没建 `munch` 库，回第二步

---

## 四、开启公网访问（字体需要）

小程序加载自定义字体走 `wx.loadFontFace`，**必须是公网 https 地址**，且域名要在白名单里。

1. 云托管服务 → **开启公网访问**，拿到默认域名（形如 `munch-server-xxx.ap-shanghai.run.tcloudbase.com`）。
2. 小程序后台 → **开发管理 → 开发设置 → 服务器域名** → 把该域名加入 **`downloadFile` 合法域名**。

**验证**：浏览器直接打开
`https://<你的域名>/assets/fonts/lxgw-wenkai-tc-subset.woff2`
能下载到约 971KB 的文件即成功。

> 走 `callContainer` 的**接口调用不需要**配 request 合法域名；这里配域名**只为字体（和将来的图片）能被 downloadFile/image 加载**。

---

## 五、前端配置与上传

### 1. 改 `miniprogram/src/config.js`

```js
export const USE_CLOUD_CONTAINER = true;              // 切到云托管通道
export const CLOUD_ENV = "prod-xxxxxx";               // 第二步的环境 ID
export const CLOUD_SERVICE = "munch-server";          // 第三步的服务名

// 字体换成公网地址（第四步拿到的域名）
export const FONT_URL = "https://<你的域名>/assets/fonts/lxgw-wenkai-tc-subset.woff2";
```

> `API_BASE_URL` 在云托管模式下不再用于接口调用，留着给本地开发切回去用。

### 2. 编译并上传

```bash
cd miniprogram
npm run build:mp-weixin      # 产物在 dist/build/mp-weixin
```

微信开发者工具 → 导入 `dist/build/mp-weixin` → **上传** → 小程序后台设为**体验版**。

**验证**（开发者工具 Console 搜 `🐶`）：
- `🐶 请求 GET /api/profile [callContainer]` ← 通道是 callContainer 而非 http，说明切换成功
- `🐶 字体已加载 LXGW WenKai TC`

---

## 六、首次使用：绑定情侣空间

两人各自打开体验版：

1. **你**先打开 → 绑定页选「我来创建」→ 角色选**大厨** → 创建 → 记下 **6 位邀请码**
2. **女朋友**打开 → 选「我有邀请码」→ 输入邀请码 → 角色选**点菜的** → 加入

之后两人看到的是同一份菜单和订单（数据按 `coupleId` 隔离，别人看不到）。

**验证**：她下一单，你在大厨端能看到（轮询 3 秒内刷新出来）。

---

## 七、已知限制 · 上线前需要你决策

### 🔴 图片上传：COS 尚未接入

当前 `STORAGE_DRIVER=cos` 在代码里**还是回退到本地磁盘**（`storage.go` 里留了 TODO 驱动位）。这意味着：

- 上传的菜品照片存在**容器本地磁盘**，**重新部署会丢失**。

两个选择：

| 方案 | 说明 |
|---|---|
| **A. 先只用 emoji 图标**（推荐先这样） | 加新菜时选 emoji，完全不受影响，零额外配置。照片功能暂不用 |
| **B. 接入腾讯云 COS** | 需要你开通 COS、建桶、给我 `COS_SECRET_ID/KEY/BUCKET_URL`，我把驱动实现补上（约 60 行 + 一个 SDK 依赖） |

### 🟡 其余待补（不影响上线）

- 下单加载的**煮锅动画**（现为系统 loading）
- **自定义 Toast**（现为系统样式）
- 长按 **480ms + 震动**（现用原生 longpress，约 350ms）
- 记录页 **Tab fade** 过渡
- 标题**字重 700**（只加载了 Regular，现为伪粗；需再传一个 Medium 字重文件）

---

## 八、日常运维

| 事项 | 做法 |
|---|---|
| 看日志 | 云托管控制台 → 服务 → 日志，搜 `[db]` / `[munch]` / `panic` |
| 发新版本 | 改代码 → 重新部署服务（后端）/ 重新编译上传（前端） |
| 回滚 | 云托管服务 → 版本列表 → 切回上一个版本 |
| 数据备份 | MySQL 控制台开启自动备份；重要数据可 `mysqldump` 导出 |
| 改配置 | 改环境变量后需**重启/重新部署**服务才生效 |

---

## 九、排错速查

| 现象 | 大概率原因 |
|---|---|
| 日志报数据库连接失败 | `munch` 库没建（第二步） |
| 接口全 401「未登录」 | 前端 `USE_CLOUD_CONTAINER` 还是 `false`，走了 http 通道没带 openid |
| 接口报「服务不存在」 | `CLOUD_SERVICE` 和云托管服务名不一致 |
| 字体没生效、控制台报加载失败 | 域名没加进 `downloadFile` 合法域名，或 `FONT_URL` 还指着 `127.0.0.1` |
| 提示「还没有情侣空间」 | 正常，去绑定页创建或加入 |
| 上传的照片过一阵没了 | 已知限制，见第七节（容器磁盘非持久） |
