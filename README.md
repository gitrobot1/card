# Card Hub

一个卡牌游戏平台，当前已实现 **斗地主（Dou Dizhu）** 的完整单机与联机玩法。后续预留炸金花、斗牛、宇宙杀、UNO 等游戏入口。

## 功能概览

### 斗地主

- **单机模式**：1 人对 2 电脑，支持叫地主、出牌、不出、提示
- **联机模式**：3 人房间，加入房间 / 准备 / 自动开局 / 下一局
- **游戏机制**：35 秒回合计时、超时自动处理、AI 出牌、牌型校验
- **界面与动画**：QQ 风格叠牌手牌、选中牌两侧让位、发牌/出牌/底牌飞入动画
- **身份与状态**：地主/农民标识、出牌计时圈、不出「不要」提示、结算与准备

## 技术栈

| 层级 | 技术 |
|------|------|
| 前端 | Vue 3、TypeScript、Vite、Vue Router、GSAP |
| 后端 | Go 1.23、Gin、GORM、JWT |
| 数据 | MySQL、Redis |

## 环境要求

- **Node.js** `20.19.6`（见 `frontend/.nvmrc`）
- **Go** `1.23.0`（见 `backend/.go-version`）
- **MySQL** 8.x
- **Redis** 6.x+

## 快速开始

### 1. 克隆项目

```bash
git clone git@github.com:gitrobot1/card.git
cd card
```

### 2. 配置后端

```bash
cp backend/config/config.example.yaml backend/config/config.yaml
```

按需修改 `backend/config/config.yaml` 中的 MySQL、Redis、JWT 等配置。

> 默认后端端口为 **8088**，需与前端 Vite 代理一致。

### 3. 准备数据库

创建 MySQL 数据库（默认名 `card_db`），启动 MySQL 与 Redis 后，后端首次运行会自动迁移表结构。

### 4. 启动后端

```bash
cd backend
./scripts/run.sh
```

服务默认监听：`http://0.0.0.0:8088`

### 5. 启动前端

```bash
cd frontend
./scripts/dev.sh
```

开发服务器默认监听：`http://0.0.0.0:6677`

浏览器访问 [http://localhost:6677](http://localhost:6677) 即可。

### 局域网访问

前后端均已绑定 `0.0.0.0`。同一局域网内其他设备请使用宿主机 IP 访问，例如 `http://192.168.x.x:6677`，不要使用 `localhost`。

## 项目结构

```text
card/
├── backend/                 # Go 后端
│   ├── cmd/server/          # 入口
│   ├── config/              # 配置文件（config.yaml 不入库）
│   ├── internal/
│   │   ├── game/doudizhu/   # 斗地主引擎、AI、牌型
│   │   ├── handler/         # HTTP 接口
│   │   ├── service/         # 业务逻辑、房间服务
│   │   └── ...
│   └── scripts/run.sh
├── frontend/                # Vue 前端
│   ├── src/
│   │   ├── views/           # 页面（模式选择、房间、对局）
│   │   ├── components/      # 斗地主 UI 组件
│   │   └── composables/     # 动画、计时等
│   └── scripts/dev.sh
└── README.md
```

## 主要路由

| 路径 | 说明 |
|------|------|
| `/` | 首页 / 登录 |
| `/games/doudizhu` | 斗地主模式选择（单机 / 联机） |
| `/games/doudizhu/solo` | 单机对局 |
| `/games/doudizhu/online` | 联机房间 |
| `/games/doudizhu/play/:gameId` | 对局页面 |

## 开发说明

- 前端通过 Vite 将 `/api`、`/health` 代理到 `http://127.0.0.1:8088`
- 联机房间与对局状态当前为 **内存存储**，后端重启后房间数据会丢失
- 联机模式使用轮询同步状态（约 1.5s），尚未接入 WebSocket
- 本地敏感配置请勿提交：`backend/config/config.yaml`、`.env` 等
- **宇宙杀**开发规范与 2v2 测试：见 [`backend/internal/game/yuzhousha/dev-guide.md`](backend/internal/game/yuzhousha/dev-guide.md)

## 后续规划

- [ ] WebSocket 实时同步
- [ ] 房间持久化（Redis）
- [ ] 炸金花、斗牛、宇宙杀、UNO

## License

Private / 暂未指定开源协议
