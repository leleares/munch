# 小食记 · munch 🌿

情侣专属点菜小程序：**你点菜，我下厨**。
一方（点菜端）浏览菜单、挑菜、备注口味、下单；另一方（大厨端）接单、备菜、上菜。

> 设计走温馨克制的抹茶绿原木风。产品规格见 [`docs/design_handoff/README.md`](docs/design_handoff/README.md)，高保真原型见 `docs/design_handoff/prototype.html`（浏览器直接打开可点按）。

---

## 技术栈

| 层 | 选型 |
|---|---|
| 小程序前端 | uni-app（Vue3 + Pinia + Vite），设计 token 落成 SCSS 变量 |
| 后端 | Go + Gin + GORM + MySQL，分层架构 |
| 部署 | 微信云托管（Go 容器 + 自带 MySQL + COS 存图，`callContainer` 免备案） |
| 登录 | 微信云托管注入 `X-WX-OPENID`（免 code2session）；本地/H5 走自签 JWT 兜底 |
| 实时 | 前端轮询 `/orders`（记录页 4s、大厨端 3s），不上订阅消息 |

## 目录结构

```
munch/
├── server/          # Go 后端（Gin + GORM + MySQL）
│   ├── cmd/api/          入口
│   ├── internal/
│   │   ├── config/       环境变量装配（兼容微信云托管的 MYSQL_* 注入）
│   │   ├── database/     GORM 连接 + AutoMigrate
│   │   ├── model/        数据模型
│   │   ├── handler/      业务处理函数
│   │   ├── service/      code2session / JWT
│   │   ├── middleware/   鉴权（X-WX-OPENID 或 JWT）
│   │   ├── storage/      图片存储（local | cos）
│   │   └── router/       路由
│   ├── Dockerfile       微信云托管构建用
│   └── .env.example
├── miniprogram/     # uni-app 前端（Vue3）
│   └── src/
│       ├── pages/       menu / detail / add / orders / chef / done / bind
│       ├── stores/      pinia：user / menu / cart / order
│       ├── api/         请求封装（callContainer / uni.request 双通道）
│       ├── styles/      通用样式片段
│       ├── uni.scss     设计 token（全局注入）
│       └── config.js    运行时配置（部署后填云环境）
├── deploy/
│   └── docker-compose.yml   本地一键起 MySQL + 后端
└── docs/design_handoff/     设计交接物料
```

---

## 本地开发

### 后端

```bash
# 1) 起本地 MySQL（用 compose 里的 mysql 服务）
cd deploy && docker compose up -d mysql

# 2) 配置环境变量
cd ../server && cp .env.example .env   # 按需改端口/密码

# 3) 跑服务（默认读 .env，监听 18090；本机 8080 被 nginx 占用）
go run ./cmd/api
```

健康检查：`curl http://127.0.0.1:18090/health` → `{"status":"ok"}`

> 本地无微信环境也能联调：登录接口支持直接传 `{"openid":"..."}` 造用户（见 `POST /api/login`）。

### 前端

```bash
cd miniprogram
npm install
npm run dev:mp-weixin   # 产物在 dist/dev/mp-weixin，用微信开发者工具导入
# 或 npm run dev:h5      # 浏览器预览（走 config.js 里的 API_BASE_URL）
```

- H5 联调：把 `src/config.js` 的 `USE_CLOUD_CONTAINER` 设为 `false`，`API_BASE_URL` 指向本地后端。
- 微信开发者工具：开发期可勾「不校验合法域名」，用 `wx.request` 直连本地；正式走云托管。

---

## 部署（微信云托管）

**完整分步操作手册见 [`docs/DEPLOY.md`](docs/DEPLOY.md)**，含验证点与排错速查。概要：

1. 开通云托管 → 建环境（记环境 ID）→ 建 MySQL 实例 → **手动 `CREATE DATABASE munch`**（平台不会自动建库）。
2. 新建服务 `munch-server`，构建上下文 `server`、Dockerfile `server/Dockerfile`、监听 `80`。
3. 配环境变量 `JWT_SECRET`、`DB_NAME=munch`；**`ALLOW_DEV_LOGIN` 绝不能配到线上**。
4. 开启公网访问 → 域名加进小程序后台 `downloadFile` 合法域名（字体要用）。
5. 前端 `config.js` 设 `USE_CLOUD_CONTAINER=true` + `CLOUD_ENV` + `CLOUD_SERVICE` + 公网 `FONT_URL`，编译上传体验版。

> `callContainer` 调用免备案；平台自动注入 `X-WX-OPENID`，后端据此识别用户，无需自己做 code2session。

---

## 接口一览（前缀 `/api`，统一返回 `{code,msg,data}`，code=0 为成功）

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/login` | 登录（code2session 或本地 openid），签发 JWT |
| GET | `/profile` | 当前用户 |
| POST | `/couple` · `/couple/join` · GET `/couple` | 情侣空间：建 / 用邀请码加入 / 查 |
| GET/POST/PATCH/DELETE | `/dishes` `/dishes/:id` | 菜品（软删，支持 `?groupId=&fav=1`） |
| GET/POST/PATCH/DELETE | `/groups` `/groups/:id` | 分组（删除时菜品迁到 fallback） |
| GET/POST | `/orders` | 下单（items 存快照）/ 列表 |
| PATCH | `/orders/:id/status` | 大厨推进：待接单→备菜中→已上菜（服务端强制状态机） |
| GET/POST/PATCH/DELETE | `/shop-items` `/shop-items/:id` | 买菜清单 |
| POST | `/upload` | 图片上传（local 落磁盘 / cos 上云） |

数据全部按 `coupleId` 隔离——两个人绑到同一情侣空间才能看到彼此的菜与单。

---

## 已完成 / 待办

**已完成**
- ✅ 后端全部接口 + 鉴权 + 数据隔离，Docker 化，本地全链路冒烟通过
- ✅ 前端全部页面（点菜/详情/加编辑/记录三段/大厨端/下单成功/绑定）
- ✅ 设计 token、购物车抽屉、状态机推进、轮询实时、买菜清单、常点
- ✅ 图标资产接入（TabBar 两态 / 双端 FAB / chevron）
- ✅ 霞鹜文楷子集化（3894 字 / 971KB）+ `loadFontFace` 接入
- ✅ 核心动效：转盘老虎机抽奖、加菜飞入购物车弧线（并发）、抽屉 slideUp、弹窗 pop、购物车回弹
- ✅ 安全：`/login` 裸 openid 后门收口到 `ALLOW_DEV_LOGIN`，默认关闭

**待补**
- ⏳ 图片上传接 COS（`server/internal/storage` 留了驱动位；未接前照片存容器磁盘，重部署会丢）
- ⏳ 煮锅加载动画、自定义 Toast、长按 480ms+震动、记录页 Tab fade
- ⏳ 标题字重 700（当前只加载 Regular，为伪粗；需再传 Medium 字重）
- ⏳ 可选：菜名 AI 配图（见交接文档第九节，key 放后端）
