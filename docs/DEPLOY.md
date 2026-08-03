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
| `STORAGE_DRIVER` | `cos` | 图片存 COS（跨端可用）；填 `local` 则落容器磁盘（重部署会丢） |
| `COS_SECRET_ID` | 子用户 SecretId | `cos` 时必填，见第七节 |
| `COS_SECRET_KEY` | 子用户 SecretKey | `cos` 时必填，见第七节 |
| `COS_BUCKET_URL` | 桶访问域名 | `cos` 时必填，如 `https://ares1-1330007488.cos.ap-beijing.myqcloud.com` |
| `STATIC_DIR` | `/app/data/uploads` | 仅 `local` 驱动用的容器内目录 |
| `PUBLIC_BASE_URL` | 服务公网域名 | 仅 `local` 驱动拼图片地址用；用 COS 可留空 |
| ~~`ALLOW_DEV_LOGIN`~~ | **不要配** | 配了就是安全漏洞 |

> `MYSQL_ADDRESS` / `MYSQL_USERNAME` / `MYSQL_PASSWORD` 由平台自动注入，**不用你填**，`config.go` 已兼容读取。
> COS 变量填不全或初始化失败时，后端会**自动退回本地磁盘**并在日志打印，不会导致服务起不来。

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

> 走 `callContainer` 的**接口调用不需要**配 request 合法域名；这里配**服务域名**只为字体能被 `loadFontFace` 加载。图片不走这个域名——它直传到 COS 桶，桶域名在第七节单独配。

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
> 图片直传用的 `COS_BUCKET` / `COS_REGION` 已内置在 `config.js`，换桶时才需改（见第七节 5）。

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

## 七、图片存储：COS 直传（STS 临时密钥）

图片走独立 COS 桶（`ares1-1330007488` / `ap-beijing`），web、native、小程序三端都能直接用返回的 URL。

**小程序为什么不走后端中继**：`callContainer` 是 JSON 通道，图片得转 base64，而它对单次请求包体有限制（约几 MB，报错码 `-606001`），手机相册动辄 5~10MB 的图必然超限。所以小程序采用**前端直传**：图片二进制直接 PUT 到 COS，不过后端、不过 callContainer，无大小限制。

上传链路：

```
小程序：chooseImage →（轻压，仅省流量）→ cos-wx-sdk-v5 直传 COS
            ↑ 拿临时密钥
        GET /api/cos-credential  ← 走 callContainer（body 空，无包体问题，自动带 X-WX-OPENID 认人）
            ↓ 后端用永久密钥 + policy 换 15 分钟临时密钥（只能 PutObject 到 munch/*）

H5/本地：chooseImage → uploadFile(multipart) → POST /api/upload → 后端转存 COS
```

> 核心安全点：**永久密钥只在后端**；前端拿到的是 15 分钟过期、只能上传本桶 `munch/` 前缀的临时密钥，反编译也无危害。这是图床/网盘的主流做法。

### 1. 建子用户密钥（别用主账号密钥）

SecretId/SecretKey 是**腾讯云账号级全局密钥**，泄露风险大。推荐建子用户：

1. 腾讯云控制台 → **访问管理 CAM → 用户 → 新建用户**（子用户）
2. 授权两项权限：
   - `QcloudCOSDataFullControl`（或用策略生成器精确到 `ares1-1330007488` 桶的读写）
   - **`sts:GetFederationToken`**（签发临时密钥必需，漏了这个直传会 403）——可加 `QcloudSTSFullAccess`，或在策略里单独放行该 action
3. 拿到该子用户的 **SecretId / SecretKey**

### 2. 桶设为公有读

COS 控制台 → 桶 `ares1-1330007488` → **权限管理 → 公有读私有写**（或配 CDN 域名）。否则返回的图片 URL 打不开。

### 3. 云托管加环境变量

munch-server 服务 → 新版本 → 环境变量：

| 变量 | 值 |
|---|---|
| `STORAGE_DRIVER` | `cos` |
| `COS_SECRET_ID` | 子用户 SecretId |
| `COS_SECRET_KEY` | 子用户 SecretKey |
| `COS_BUCKET_URL` | `https://ares1-1330007488.cos.ap-beijing.myqcloud.com` |

> `COS_REGION`（ap-beijing）和 `COS_APPID`（1330007488）会从 `COS_BUCKET_URL` 自动解析，不用单独配。

### 4. 小程序后台加合法域名

开发管理 → 开发设置 → 服务器域名，给 `ares1-1330007488.cos.ap-beijing.myqcloud.com` **同时加两处**：

- **`uploadFile` 合法域名** ← 直传上传用
- **`downloadFile` 合法域名** ← `<image>` 显示图用

漏了 uploadFile 会传不上去，漏了 downloadFile 会显示不出。

### 5. 前端配置

`miniprogram/src/config.js` 里已内置（非机密，可放前端）：

```js
export const COS_BUCKET = "ares1-1330007488";
export const COS_REGION = "ap-beijing";
```

换桶时改这两个 + 后端 `COS_BUCKET_URL` 即可。

**验证**：加新菜 → 拍照或选相册大图上传 → 开发者工具 Console 出现 `🐶 COS 上传成功 https://...`；该 URL 浏览器能直接打开；菜品卡片能显示图。

> COS 在**北京**、云托管在**上海**，签发密钥走公网（流量极小，无感知）；图片直传是小程序↔COS 直连，也不经过云托管。

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
| 上传照片成功但图显示不出 | COS 桶没设公有读，或域名没加进 `downloadFile` 合法域名（第七节 2/4） |
| 上传报「上传失败」/ 域名不合法 | 域名没加进 `uploadFile` 合法域名（第七节 4） |
| 上传报 403 / 获取临时密钥失败 | 子用户少了 `sts:GetFederationToken` 权限，或 policy 资源写错（第七节 1） |
