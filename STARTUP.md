## 项目启动说明（bell-best）

本说明用于记录如何在本地启动 `bell-best` 项目，避免以后再次忘记启动方式。

### 1. 环境依赖

- **Go 版本**：建议 Go **1.22 及以上**
- **MySQL 数据库**
  - 已创建好数据库（名称你可自定义，与配置文件一致即可）
- **Redis**

### 2. 配置文件 `config.yaml`

项目启动时会在当前工作目录读取 `config.yaml`（见 `setting/settings.go` 中的 `viper.SetConfigFile("config.yaml")`），因此需要在项目根目录 `bell-best` 下创建该文件。

示例配置（根据你自己的环境调整）：

```yaml
name: "bell-best"
mode: "dev"          # dev / release
version: "v1.0"
start_time: "2025-01-01"
machine_id: 1
port: 8081           # HTTP 监听端口，对应 main.go 中的 setting.Conf.Port

log:
  level: "debug"
  filename: "app.log"
  max_size: 128
  max_age: 30
  max_backups: 7

mysql:
  host: "127.0.0.1"
  port: 3306
  user: "root"
  password: "your_password"
  dbname: "your_db_name"
  max_open_conns: 200
  max_idle_conns: 50

redis:
  host: "127.0.0.1"
  port: 6379
  password: ""
  db: 0
  pool_size: 100
```

> 注意：**MySQL** 和 **Redis** 都需要先启动，并确保上面的连接信息配置正确，否则程序在初始化阶段会直接退出。

### 3. 启动步骤（Windows PowerShell 示例）

1. 打开 PowerShell，进入项目目录：

   ```powershell
   cd D:\goproject\bell-best
   ```

2. （可选，首次或依赖更新时）拉取依赖：

   ```powershell
   go mod tidy
   ```

3. 启动服务：

   ```powershell
   go run main.go
   ```

   程序会：

   - 加载 `config.yaml` 配置
   - 初始化日志、MySQL、Redis、雪花算法等
   - 启动 Gin HTTP 服务，监听端口为 `config.yaml` 中的 `port`（例如 `8081`）

### 4. 启动成功后的验证方式

- 浏览器访问：
  - `http://127.0.0.1:8081/ping`（测试接口）
  - `http://127.0.0.1:8081/swagger/index.html`（如果已生成并开启 swagger）
- 或使用 curl：

  ```powershell
  curl http://127.0.0.1:8081/ping
  ```

如返回正常字符串或 JSON，即表示服务启动成功。

### 5. 常见启动问题排查

- **提示找不到 `config.yaml`**：确认该文件位于项目根目录 `bell-best`，并且启动时当前工作目录就是该目录。
- **MySQL/Redis 连接失败**：检查对应服务是否已启动、端口是否正确、用户/密码是否匹配。
- **端口占用**：如果 `8081` 被占用，可以在 `config.yaml` 中修改 `port`，然后重新启动。


