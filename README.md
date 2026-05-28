# VisionVault

VisionVault 是一个面向企业文件采集、集中存储、检索与运维管理的多模块系统。当前仓库包含服务端 `SkyBase`、前端管理台 `SkyView`、终端采集 Agent `SkyDrop`，以及 Agent 自动更新程序 `SkyDropAutoUpdate`。

## 系统组成

### 1. SkyBase

`SkyBase` 是系统控制平面，负责：

- 用户、部门、角色、菜单、系统配置等后台能力
- Agent 注册、心跳、任务提交、版本分发
- 存储管理、文件管理、同步日志、扫描报告
- 告警、任务进度、License 等业务能力

技术栈：

- Go 1.25
- MySQL
- Redis

默认监听地址：

- `http://localhost:8080`

### 2. SkyView

`SkyView` 是基于 Vue 3 + Vite 的管理前端，用于：

- 监控总览
- Agent、分组、版本、日志管理
- 文件、存储、权限、任务进度管理
- 用户、部门、角色、登录日志、系统设置管理

技术栈：

- Vue 3
- TypeScript
- Vite
- Arco Design Vue

默认开发地址：

- `http://localhost:5173`

### 3. SkyDrop

`SkyDrop` 是部署在终端主机上的采集 Agent，主要负责：

- 向 `SkyBase` 定期发送心跳
- 接收并执行后台下发的采集/扫描任务
- 提交同步结果与扫描结果
- 提供本地目录访问接口给宿主程序使用

### 4. SkyDropAutoUpdate

`SkyDropAutoUpdate` 是 Agent 自动更新守护程序，主要负责：

- 轮询 `SkyBase` 查询新版本
- 下载升级包
- 停止旧 Agent
- 替换可执行文件并拉起新版本

## 仓库结构

```text
VisionVault/
├── SkyBase/             # Go 服务端
├── SkyView/             # Vue 前端管理台
├── SkyDrop/             # Go 采集 Agent
├── SkyDropAutoUpdate/   # Go 自动更新程序
├── docs/                # 需求和设计资料
└── LICENSE
```

## 模块协作关系

```text
SkyView  <---- HTTP/JSON ---->  SkyBase  <---- HTTP/JSON ---->  SkyDrop
                                           ^
                                           |
                                   SkyDropAutoUpdate
```

## 快速启动

### 环境要求

- Go
  - `SkyBase` 需要 Go 1.25+
  - `SkyDrop` / `SkyDropAutoUpdate` 当前 `go.mod` 为 Go 1.19
- Node.js 18+
- MySQL 8.x
- Redis 6.x 或更高版本

### 1. 初始化数据库

执行数据库脚本：

```bash
mysql -uroot -p < SkyBase/db/001_init.sql
```

默认会创建数据库：

- `skyvv`

### 2. 启动 SkyBase

方式一：使用脚本启动

```bash
cd SkyBase
./scripts/start.sh
```

方式二：手动启动

```bash
cd SkyBase
source ./scripts/env.sh
go run ./cmd/server
```

启动后可检查：

- 首页: `http://localhost:8080/`
- 健康检查: `http://localhost:8080/healthz`

### 3. 启动 SkyView

```bash
cd SkyView
npm install
npm run dev
```

默认前端会请求：

- `VITE_API_BASE=http://localhost:8080`

如果后端地址不是本机 `8080`，可以这样启动：

```bash
cd SkyView
VITE_API_BASE=http://your-host:8080 npm run dev
```

### 4. 运行 SkyDrop

```bash
cd SkyDrop
go run .
```

常用参数：

- `-console`：SkyBase 地址，默认示例值为内网地址
- `-workspace`：工作目录
- `-ip`：网卡 IP 选择条件
- `-help`：显示帮助
- `-version`：显示版本

示例：

```bash
cd SkyDrop
go run . -console http://localhost:8080 -workspace ./ -ip 10.
```

### 5. 运行 SkyDropAutoUpdate

```bash
cd SkyDropAutoUpdate
go run . -console http://localhost:8080 -version 1.9.2
```

## SkyBase 环境变量

`SkyBase/scripts/env.sh` 已提供默认环境变量。常用项如下：

| 变量名 | 默认值 | 说明 |
| --- | --- | --- |
| `SKYBASE_HTTP_ADDR` | `:8080` | 服务监听地址 |
| `SKYBASE_ADMIN_USERNAME` | `admin` | 初始管理员账号 |
| `SKYBASE_ADMIN_PASSWORD` | `admin123` | 初始管理员密码 |
| `SKYBASE_DB_HOST` | `172.16.61.36` | MySQL 主机 |
| `SKYBASE_DB_PORT` | `6612` | MySQL 端口 |
| `SKYBASE_DB_USER` | `root` | MySQL 用户 |
| `SKYBASE_DB_PASSWORD` | `emacle2013` | MySQL 密码 |
| `SKYBASE_DB_NAME` | `skyvv` | MySQL 数据库名 |
| `SKYBASE_REDIS_ADDR` | `172.16.61.36:6613` | Redis 地址 |
| `SKYBASE_REDIS_PASSWORD` | 空 | Redis 密码 |
| `SKYBASE_REDIS_DB` | `0` | Redis DB |

建议在本地、测试、生产环境中分别覆盖这些默认值，不要直接沿用脚本中的内网地址和明文密码。

## 主要接口

### SkyBase 管理端接口

部分核心接口前缀：

- `/api/v1/auth/*`
- `/api/v1/agents*`
- `/api/v1/groups*`
- `/api/v1/storage*`
- `/api/v1/files*`
- `/api/v1/sync-logs`
- `/api/v1/versions*`
- `/api/v1/roles*`
- `/api/v1/departments*`
- `/api/v1/users*`
- `/api/v1/alert-*`
- `/api/v1/system-configs*`
- `/api/v1/licenses*`

### SkyDrop / AutoUpdate 开放接口

Agent 与自动更新程序使用的接口前缀：

- `/sky/agent/heartbeat`
- `/sky/agent/commit`
- `/sky/agent/scan/commit`
- `/sky/agent/version`
- `/sky/agent/download`

## 当前前端模块

从 `SkyView` 当前路由和菜单可见，系统已覆盖以下页面方向：

- Overview
- Agents / Groups / Sync Logs / Monitor / Versions
- Files / File Logs / File Permissions / Task Progress
- Alert Groups / Alert Logs / Message Channels / Alert Policies
- Backup Devices / Backup Tasks / Backup Logs
- Storage / Users / Departments / Roles / Login Logs / Settings / License

## 文档资料

- 需求文档：[`docs/需求文档.md`](docs/需求文档.md)
- 前端设计提案：[`SkyView/SKYVIEW_DESIGN_PROPOSAL.md`](SkyView/SKYVIEW_DESIGN_PROPOSAL.md)
- Agent 安装说明：[`SkyDrop/Readme.txt`](SkyDrop/Readme.txt)

## 开发说明

- 仓库当前为多模块结构，没有统一的 monorepo 启动脚本
- `SkyBase`、`SkyView`、`SkyDrop`、`SkyDropAutoUpdate` 需要分别启动或构建
- `SkyDrop` 与 `SkyDropAutoUpdate` 更偏向 Windows 场景，仓库中包含 `.bat`、签名工具和安装相关文件

## 建议补充

如果后续继续完善仓库，建议补齐以下内容：

- 根目录 `.env.example`
- 一键本地开发启动脚本
- Docker / docker-compose 部署方案
- 模块级 README
- 接口文档与部署文档
