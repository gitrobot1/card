# Card Hub

卡牌游戏平台，当前已实现 **斗地主、斗牛、宇宙杀** 等玩法（单机 + 部分联机）。前端 Vue 3，后端 Go + Gin。

---

## 生产部署（后端）

下面以 **Linux 服务器**（常见 x86_64 VPS）为例。后端是单个静态链接友好的 Go 二进制，**推荐在开发机交叉编译后上传**，服务器上 **不必安装 Go**。

### 服务器需要什么

| 组件 | 要求 | 说明 |
|------|------|------|
| **Go** | **不需要**（若采用交叉编译） | 仅在服务器上现场编译时才需要，见下文 |
| **操作系统** | Linux x86_64 / arm64 | 与编译时的 `GOOS` / `GOARCH` 一致 |
| **MySQL** | 8.x | 首次启动自动建表（GORM AutoMigrate） |
| **Redis** | 6.x+ | 健康检查与部分功能依赖 |
| **开放端口** | 默认 **8088**（或由 Nginx 反代，不对外暴露） | 见 `config.yaml` 的 `server.port` |

> 运行时 **只需要**：`card-server` 二进制 + `config.yaml` + 可连通的 MySQL / Redis。  
> 宇宙杀武将/皮肤等静态数据已 **embed 进二进制**，无需额外拷贝 JSON。

### 开发机：打包后端

在项目根目录执行（需本机已装 **Go 1.23.0**，见 `backend/.go-version`）：

```bash
cd backend

# 拉依赖（首次或 go.mod 变更后）
go mod download

# 交叉编译 Linux x86_64 可执行文件（最常见 VPS）
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -ldflags="-s -w" -o dist/card-server ./cmd/server

# 若是 ARM 服务器（如部分云 ARM 实例）：
# CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o dist/card-server ./cmd/server
```

产物：

```text
backend/dist/card-server    # 单个可执行文件，可直接 scp 到服务器
```

上传到服务器示例：

```bash
scp backend/dist/card-server user@your-server:/opt/card/
scp backend/config/config.example.yaml user@your-server:/opt/card/config.yaml
```

### 服务器：Go 环境（仅当在服务器上编译时）

若选择在服务器上 `git pull` 后现场编译，需安装：

- **Go 1.23.0**（与 `backend/go.mod` 中 `go 1.23.0` 一致）
- 无需 GCC（使用 `CGO_ENABLED=0` 纯 Go 编译即可）

```bash
# 验证版本
go version   # 应显示 go1.23.0

cd /opt/card/backend
export GOPROXY=https://goproxy.cn,direct   # 国内可选
CGO_ENABLED=0 go build -ldflags="-s -w" -o /opt/card/card-server ./cmd/server
```

**结论**：生产环境更省事的做法是 **开发机交叉编译 → 只上传二进制**；服务器只跑 MySQL、Redis 和 `card-server`。

### 服务器：配置

```bash
mkdir -p /opt/card
# 编辑生产配置（勿提交仓库）
vi /opt/card/config.yaml
```

由 `backend/config/config.example.yaml` 复制并修改，生产环境建议至少改：

```yaml
server:
  host: "0.0.0.0"
  port: 8088
  mode: "release"          # 生产务必 release

mysql:
  host: "127.0.0.1"        # 或内网 RDS 地址
  port: 3306
  database: "card_db"
  username: "card"
  password: "<强密码>"

redis:
  host: "127.0.0.1"
  port: 6379
  password: ""             # 有密码则填写

auth:
  jwt_secret: "<随机长字符串>"   # 必须修改
  token_ttl: "720h"
```

MySQL 需预先创建空库，例如：

```sql
CREATE DATABASE card_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'card'@'%' IDENTIFIED BY '<强密码>';
GRANT ALL ON card_db.* TO 'card'@'%';
FLUSH PRIVILEGES;
```

### 服务器：运行

**前台试跑**（确认 MySQL / Redis / 配置无误）：

```bash
cd /opt/card
chmod +x card-server
./card-server -config ./config.yaml
# 或使用环境变量指定配置路径：
# CARD_CONFIG_PATH=/opt/card/config.yaml ./card-server
```

健康检查：

```bash
curl -s http://127.0.0.1:8088/health
```

**systemd 常驻**（示例）：

```ini
# /etc/systemd/system/card-server.service
[Unit]
Description=Card Hub API
After=network.target mysql.service redis.service

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/card
Environment=CARD_CONFIG_PATH=/opt/card/config.yaml
ExecStart=/opt/card/card-server
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now card-server
sudo systemctl status card-server
journalctl -u card-server -f
```

### 前端静态资源（简要）

后端 **不** 托管前端页面。生产通常用 Nginx 提供 `frontend/dist`，并把 `/api`、`/health`、`/ws` 反代到 `127.0.0.1:8088`：

```bash
cd frontend
npm ci
npm run build    # 产出 frontend/dist/
```

同域部署时 `frontend/public/config.json` 中 `apiBaseUrl` 留空即可（浏览器用当前站点 origin）。  
若 API 在独立域名，部署前改为 `"apiBaseUrl": "https://api.example.com"`。

Nginx 反代要点：

```nginx
location /api/ {
    proxy_pass http://127.0.0.1:8088;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
}
location /health {
    proxy_pass http://127.0.0.1:8088;
}
location /ws/ {
    proxy_pass http://127.0.0.1:8088;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
}
location / {
    root /var/www/card/dist;
    try_files $uri $uri/ /index.html;
}
```

---

## 本地开发

### 环境要求

| 工具 | 版本 |
|------|------|
| Node.js | `20.19.6`（`frontend/.nvmrc`） |
| Go | `1.23.0`（`backend/.go-version`） |
| MySQL | 8.x |
| Redis | 6.x+ |

### 快速启动

```bash
git clone git@github.com:gitrobot1/card.git
cd card

cp backend/config/config.example.yaml backend/config/config.yaml
# 按需改 MySQL / Redis / JWT

cd backend && ./scripts/run.sh          # http://0.0.0.0:8088
cd frontend && ./scripts/dev.sh         # http://0.0.0.0:6677
```

浏览器访问 [http://localhost:6677](http://localhost:6677)。  
开发时 Vite 将 `/api`、`/health`、`/ws` 代理到 `127.0.0.1:8088`。

同一局域网其他设备用宿主机 IP，例如 `http://192.168.x.x:6677`。

### 后端测试

```bash
cd backend
./scripts/test.sh smoke -v              # 冒烟
./scripts/test.sh yzs -v                # 宇宙杀 cardtest
```

宇宙杀开发规范见 [`backend/internal/game/yuzhousha/dev-guide.md`](backend/internal/game/yuzhousha/dev-guide.md)。

---

## 项目结构

```text
card/
├── backend/
│   ├── cmd/server/           # 入口 main.go
│   ├── config/
│   │   ├── config.example.yaml
│   │   └── config.yaml       # 本地/生产配置（不入库）
│   ├── internal/
│   │   ├── game/             # 各游戏引擎
│   │   ├── handler/          # HTTP / WebSocket
│   │   ├── service/
│   │   └── router/
│   └── scripts/
│       ├── run.sh            # 本地 go run
│       └── test.sh           # cardtest 套件
├── frontend/
│   ├── public/config.json    # 生产 API 地址（可选）
│   └── src/
└── README.md
```

---

## 运维说明

- 联机房间 / 对局状态当前多为 **内存存储**，**重启后端会丢失进行中的房间**（持久化在规划中）。
- 勿将 `backend/config/config.yaml`、`.env`、`sim_logs/` 提交到 Git。
- 生产务必修改 `auth.jwt_secret`，并将 `server.mode` 设为 `release`。

---

## License

Private / 暂未指定开源协议
