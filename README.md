# BSL — Bilibili StreetLight（B站路灯）

B站直播弹幕监控系统，实时检测 `#路灯` 指令弹幕，通过 QQ 机器人（OneBot V11）推送通知，提供 Web 管理面板。

## 功能概览

- **直播弹幕采集** — 连接 B站直播间 WebSocket，解析二进制协议（含 Brotli 解压），实时收集弹幕、礼物、SuperChat、舰长
- **路灯检测** — 监听 `#路灯 内容` 弹幕，匹配后标记为「路灯事件」
- **QQ 通知** — 通过 OneBot V11 反向 WebSocket 将路灯事件推送到指定 QQ 群
- **Web 管理面板** — 仪表盘、房间管理、弹幕/礼物/SC 历史、黑名单、统计数据
- **黑名单过滤** — 屏蔽指定用户的弹幕，路灯事件也会被过滤

## 架构

```
┌─────────────┐    WebSocket     ┌──────────────┐
│  B站直播间   │ ◄────────────── │   Collector   │
│  (CDN弹幕)  │    Binary/JSON   │   (采集引擎)   │
└─────────────┘                  └──────┬───────┘
                                        │ Callbacks
                               ┌────────┴────────┐
                               │                 │
                          ┌────▼────┐      ┌─────▼─────┐
                          │PostgreSQL│     │  Notifier  │
                          └────┬────┘      └─────┬─────┘
                               │                 │
                    ┌──────────┴──────┐   ┌──────▼──────┐
                    │   API Server    │   │  QQ Bot      │
                    │  (Gin :8080)    │   │ (OneBot V11) │
                    └────────┬───────┘   └──────┬──────┘
                             │                  │
                    ┌────────▼───────┐   ┌──────▼──────┐
                    │  Web Dashboard │   │   QQ群通知    │
                    │  (React + AntD)│   └─────────────┘
                    └────────────────┘
```

## 技术栈

| 层 | 技术 |
|---|------|
| 后端 | Go 1.25+, Gin, pgx/v5, gorilla/websocket |
| 数据库 | PostgreSQL 14+ |
| 前端 | React 18, Vite, Ant Design 5, ECharts |
| 协议 | B站直播 WebSocket 二进制协议, OneBot V11 |
| 部署 | Docker (多阶段构建), systemd |

## 前置依赖

- **Go** ≥ 1.25
- **Node.js** ≥ 22（仅前端构建）
- **PostgreSQL** ≥ 14
- **OneBot V11 实现**（如 [go-cqhttp](https://github.com/Mrs4s/go-cqhttp)、[Lagrange](https://github.com/LagrangeDev/Lagrange.Core)）

## 快速开始

### 1. 克隆仓库

```bash
git clone https://github.com/RAINighty/bsl.git
cd bsl
```

### 2. 配置

编辑 `config.yaml`：

```yaml
server:
  port: 8080           # API 和 Web 面板端口
  host: "0.0.0.0"

database:
  url: "postgres://bsl:password@localhost:5432/bsl?sslmode=disable"
  max_connections: 20

bilibili:
  cookie: ""           # B站 Cookie，留空则匿名连接

onebot:
  ws_path: "/onebot"   # OneBot 反向 WebSocket 路径

collector:
  heartbeat_interval: 30       # 心跳间隔（秒）
  reconnect_backoff_max: 60    # 重连退避上限（秒）
  stats_flush_interval: 60     # 统计刷新间隔（秒）
```

### 3. 初始化数据库

```bash
# 创建 PostgreSQL 数据库
createdb bsl
```

程序启动时会自动执行数据库迁移（建表），无需手动操作。

### 4. 构建前端

```bash
cd web
npm install
npm run build
cd ..
```

### 5. 启动

```bash
go run .
```

程序将启动 API 服务器（默认 `0.0.0.0:8080`），并等待 OneBot 连接。

### 6. 配置 OneBot

在你的 OneBot 客户端中设置反向 WebSocket 地址：

```
ws://127.0.0.1:8080/onebot
```

连接成功后，在 QQ 群发送 `help` 查看可用指令。

### 7. 添加直播间

打开浏览器访问 `http://localhost:8080`，在 **房间管理** 页面添加要监控的 B站直播间房间号。

## QQ 群指令

在 QQ 群中 @机器人 或发消息，支持以下指令：

| 指令 | 说明 |
|------|------|
| `help` / `帮助` | 显示帮助 |
| `房间 123456` | 订阅/取消订阅指定房间的路灯通知（开关式） |
| `订阅列表` | 查看本群已订阅的房间 |
| `黑名单 list` | 查看全局黑名单 |
| `黑名单 add 12345 刷屏` | 添加用户到黑名单 |
| `黑名单 remove 12345` | 从黑名单移除用户 |

## Web 管理面板

启动后访问 `http://localhost:8080`：

| 页面 | 功能 |
|------|------|
| **仪表盘** | 监听房间数、直播中、今日路灯数、在线房间列表 |
| **房间管理** | 添加/删除房间，开关监听和统计 |
| **房间详情** | 弹幕历史、路灯记录、礼物/SC/舰长记录、弹幕统计图表 |
| **黑名单** | 管理被屏蔽的用户 |

## API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/dashboard` | 仪表盘汇总数据 |
| GET | `/api/rooms` | 所有房间列表 |
| POST | `/api/rooms` | 添加房间 |
| DELETE | `/api/rooms/:id` | 删除房间 |
| PUT | `/api/rooms/:id/listening` | 开关监听 |
| PUT | `/api/rooms/:id/stats` | 开关统计 |
| GET | `/api/rooms/:id` | 房间详情 |
| GET | `/api/rooms/:id/danmaku` | 弹幕列表（分页） |
| GET | `/api/rooms/:id/streetlights` | 路灯事件列表（分页） |
| GET | `/api/rooms/:id/gifts` | 礼物记录（分页） |
| GET | `/api/rooms/:id/sc` | SC 记录（分页） |
| GET | `/api/rooms/:id/guards` | 舰长记录（分页） |
| GET | `/api/rooms/:id/stats` | 弹幕统计（按分钟） |
| GET | `/api/blacklist` | 黑名单列表 |
| POST | `/api/blacklist` | 添加黑名单 `{"uid":123,"reason":"刷屏"}` |
| DELETE | `/api/blacklist/:uid` | 移除黑名单 |

分页参数：`?limit=50&offset=0`，limit 最大 200。

## Docker 部署

```bash
# 构建镜像
docker build -t bsl .

# 运行（需要 PostgreSQL 已启动）
docker run -d \
  --name bsl \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  -e DATABASE_URL="postgres://bsl:password@host.docker.internal:5432/bsl" \
  bsl
```

## systemd 部署

```bash
# 复制二进制和配置
sudo mkdir -p /opt/bsl
sudo cp bsl config.yaml /opt/bsl/
sudo cp -r web/dist /opt/bsl/web/dist

# 安装服务
sudo cp bsl.service /etc/systemd/system/
sudo useradd -r -s /bin/false bsl
sudo systemctl daemon-reload
sudo systemctl enable --now bsl
```

## 项目结构

```
bsl/
├── main.go                  # 入口：组装各模块并启动
├── config.yaml              # 配置文件
├── Dockerfile               # 多阶段构建（Go + Node → Alpine）
├── bsl.service              # systemd 单元文件
├── go.mod / go.sum
├── internal/
│   ├── api/routes.go        # REST API（Gin）
│   ├── bot/
│   │   ├── server.go        # OneBot V11 WebSocket 服务端
│   │   ├── commands.go      # QQ 群指令处理
│   │   └── notifier.go      # 路灯事件 → QQ 通知
│   ├── collector/
│   │   ├── manager.go       # 多房间管理，添加/移除/重连
│   │   ├── client.go        # 单房间 WebSocket 客户端
│   │   ├── packet.go        # B站二进制协议封包/解包
│   │   ├── parser.go        # 弹幕消息解析（含 Brotli 解压）
│   │   ├── streetlight.go   # #路灯 检测
│   │   └── stats.go         # 弹幕统计
│   ├── config/config.go     # YAML 配置加载
│   ├── db/
│   │   ├── db.go            # 数据库连接池
│   │   ├── migrations.go    # 自动建表
│   │   └── queries.go       # 数据访问层
│   └── model/models.go      # 数据结构定义
├── web/
│   ├── src/
│   │   ├── main.jsx         # React 入口
│   │   ├── App.jsx          # 路由和布局
│   │   ├── api/index.js     # API 请求封装
│   │   └── pages/
│   │       ├── Dashboard.jsx  # 仪表盘
│   │       ├── Rooms.jsx      # 房间管理
│   │       ├── RoomDetail.jsx # 房间详情
│   │       ├── Blacklist.jsx  # 黑名单管理
│   │       └── Settings.jsx   # 设置页（占位）
│   └── vite.config.js
└── test/
    ├── packet_test.go       # 协议封包/解包测试
    ├── streetlight_test.go  # 路灯检测测试
    └── config_test.go       # 配置加载测试
```

## 开发

```bash
# 运行测试
go test ./...

# 前端开发模式（热更新）
cd web && npm run dev

# 后端编译
go build -o bsl .
```

## License

MIT
