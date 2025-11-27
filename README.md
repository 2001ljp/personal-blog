# bell-best · Go 博客服务端

`bell-best` 是一个使用 Go + Gin 构建的博客社区后端，内置用户认证、帖子/社区管理、投票、Swagger 文档与链路日志等能力，可作为个人博客或论坛型项目的脚手架。

## 功能特性

- **多模块初始化**：启动时依次加载配置、Zap 日志、MySQL、Redis、雪花算法、Gin 路由，并支持优雅关机（见 `main.go`）。
- **JWT 权限控制**：登录后发放 Token，受保护的路由通过 Gin 中间件校验身份（见 `router/routes.go`）。
- **帖子/社区能力**：提供发帖、详情查询、分页列表、按时间/热度排序、社区聚合与投票接口，控制器→业务逻辑→DAO 分层清晰（见 `controller`/`logic`/`dao`）。
- **统一配置中心**：使用 Viper 热加载 `config.yaml`，集中管理服务、日志、MySQL、Redis 等配置项（见 `setting/settings.go`）。
- **内置 Swagger**：集成 swaggo，可通过 `/swagger/index.html` 查看接口说明，与 README 的项目级文档互补。

## 技术栈

- **Language**：Go 1.24+
- **Framework**：Gin、swaggo
- **Storage**：MySQL（sqlx）、Redis（go-redis/v8）
- **Infra**：Viper 配置、Zap + Lumberjack 日志、Snowflake ID、JWT（golang-jwt）

## 目录结构

```
├── controller      # HTTP 层，请求校验 & 响应封装
├── dao             # 数据访问层（mysql/redis）
├── docs            # Swagger 生成产物
├── logic           # 业务逻辑
├── logger          # Zap 配置与 Gin 中间件
├── middlewares     # JWT 等通用中间件
├── models          # 数据模型 & 请求参数
├── pkg             # 公共库（snowflake/jwt）
├── router          # 路由注册
├── setting         # 配置加载
├── STARTUP.md      # 启动指引
└── main.go
```

## 快速开始

1. **准备环境**
   - Go 1.22+（推荐 1.24）
   - MySQL、Redis 已启动且具备访问凭证

2. **克隆与依赖**
   ```bash
   git clone https://github.com/2001ljp/personal-blog.git
   cd personal-blog
   go mod tidy
   ```

3. **配置 `config.yaml`**
   - 可参考 `STARTUP.md` 中的示例，至少需要配置 `log`、`mysql`、`redis` 与 `port` 等字段。
   - 运行目录需能读取该文件（默认当前工作目录）。

4. **运行服务**
   ```bash
   go run main.go
   ```
   默认监听 `:8081`，可通过 `config.yaml` 中的 `port` 修改。

5. **验证接口**
   - `GET http://127.0.0.1:8081/ping`（需要携带 JWT 才能通过）
   - `POST /api/v1/signup`、`POST /api/v1/login`
   - `GET http://127.0.0.1:8081/swagger/index.html` 查看接口说明

## 配置说明

`setting/settings.go` 定义了所有可配置项，核心字段如下：

| 字段 | 说明 |
| ---- | ---- |
| `name`, `mode`, `version` | 服务元信息 |
| `start_time`, `machine_id` | Snowflake ID 配置 |
| `port` | HTTP 监听端口 |
| `log` | Zap 日志级别、文件、滚动策略 |
| `mysql` | MySQL 连接、连接池配置 |
| `redis` | Redis 主机、密码、库号、连接池 |

修改配置文件后，Viper 会自动监听并热更新，无需重启。

## Swagger 文档

- 本仓库已包含 `docs/` 目录，可直接访问 `http://localhost:8081/swagger/index.html`。
- 如需重新生成，请先安装 `swag` 工具，再执行：
  ```bash
  swag init -g main.go -o docs
  ```
  生成的 Swagger 提供接口入参与响应示例，而 README 负责描述项目背景、部署方式和整体架构，两者内容互补而非重复。

## 常见问题

- **推送 GitHub 提示日志过大**：将 `web_app*.log` 移出仓库或使用 `.gitignore` 忽略，并通过 `git filter-repo` 清理历史。
- **端口占用/配置加载失败**：确保当前工作目录存在 `config.yaml`，并修改 `port` 避免冲突。
- **JWT 保护的路由 401**：先调用 `/api/v1/login` 获取 token，再以 `Authorization: Bearer <token>` 访问受保护资源。

如需更多运行截图或表结构，可将 SQL、ER 图等补充到 `docs/` 或 `项目介绍.md` 中。