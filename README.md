# GO
## 1. 环境准备

### 1.1 安装 Go 语言

- 访问 [golang.org/dl](https://golang.org/dl) 下载 Windows 版本
- 安装后打开 PowerShell，输入 `go version` 验证安装成功

### 1.2 安装 MongoDB

- 访问 [mongodb.com/try/download/community](https://www.mongodb.com/try/download/community) 下载
- 安装时选择 "Complete" 安装
- 安装完成后，MongoDB 服务会自动启动

## 2. 配置修改

### 2.1 修改 JWT 密钥

打开 `configs/config.yaml` 文件，找到 `jwt.secret` 部分：

```yaml
jwt:
  # JWT 签名密钥（生产环境请使用安全的随机值）
  secret: "your_super_secret_jwt_key_here"  # 修改这一行
  # 访问令牌有效期（分钟）
  access_ttl_minutes: 30
  # 刷新令牌有效期（天）
  refresh_ttl_days: 14
```

**重要**：将 `"your_super_secret_jwt_key_here"` 替换为一个复杂的字符串，比如：

- `"my_roleplay_app_secret_key_2024"`
- `"super_secret_jwt_key_for_roleplay_backend"`

### 2.2 检查其他配置

```yaml
server:
  # 服务监听端口（如果 8080 被占用，可以改为 8081、8082 等）
  port: 8080

mongo:
  # MongoDB 连接字符串（如果 MongoDB 安装在其他机器，需要修改 IP）
  uri: "mongodb://localhost:27017"
  # 数据库名（可以保持默认）
  database: "roleplay"

sms:
  # 开发环境保持 true 即可
  enabled: true
  # Mock 验证码（测试时使用）
  mock_code: "123456"
```

## 3. 运行步骤

### 3.1 打开 PowerShell

- 按 `Win + R`，输入 `powershell`，回车
- 或者按 `Win + X`，选择 "Windows PowerShell"

### 3.2 进入项目目录

```powershell
cd "你的项目目录"
```

### 3.3 安装依赖

```powershell
go mod tidy
```

这个命令会：

- 下载项目需要的所有 Go 包
- 可能需要几分钟时间
- 如果看到下载进度条，说明正在工作

注意：国内可以换源：
```
 # 设置七牛云镜像（推荐）
 go env -w GOPROXY=https://goproxy.cn,direct
 # 设置阿里云镜像（备选）
 go env -w GOPROXY=https://mirrors.aliyun.com/goproxy/,direct
```

### 3.4 启动服务

```powershell
go run ./cmd/server
```


## 4. 测试服务

### 4.1 健康检查

打开浏览器，访问：`http://localhost:8080/healthz`
应该看到：`{"status":"ok"}`

### 4.2 发送验证码（测试）

使用 Postman 或者 PowerShell 的 `Invoke-RestMethod`：

```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/user/send_code" -Method POST -ContentType "application/json" -Body '{"phone":"13800138000"}'
```

应该返回：

```json
{
  "code": 200,
  "message": "验证码已发送",
  "data": {
    "mock_code": "123456"
  }
}
```

### 4.3 登录测试

```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/user/login" -Method POST -ContentType "application/json" -Body '{"phone":"13800138000","code":"123456"}'
```

### 4.4 参考测试脚本
位于：scripts/
```
 powershell -ExecutionPolicy Bypass -File scripts/test_all_fixed.ps1
```

## 5. 常见问题

### 5.1 端口被占用

如果看到 "端口已被使用" 错误：

- 修改 `configs/config.yaml` 中的 `port: 8080` 为 `port: 8081`
- 或者找到占用 8080 端口的程序并关闭

### 5.2 MongoDB 连接失败

如果看到 "MongoDB 连接失败"：

- 确保 MongoDB 服务已启动
- 检查 `configs/config.yaml` 中的 `mongo.uri` 是否正确
- 可以尝试重启 MongoDB 服务

### 5.3 Go 命令找不到

如果提示 `go: command not found`：

- 重新安装 Go
- 或者重启 PowerShell（让环境变量生效）

## 6. 开发建议

### 6.1 保持服务运行

- 在开发过程中，保持 `go run ./cmd/server` 运行
- 如果需要修改代码，按 `Ctrl + C` 停止服务，然后重新运行

### 6.2 查看日志

服务运行时会输出日志，包括：

- 配置加载状态
- MongoDB 连接状态
- 索引创建状态
- HTTP 请求日志

### 6.3 修改配置后

每次修改 `configs/config.yaml` 后，需要：

1. 停止服务（`Ctrl + C`）
2. 重新运行 `go run ./cmd/server`

### 6.4 支持的API
见GO/docs/openapi.yaml
待完善：一键登录、上传用户头像等等

