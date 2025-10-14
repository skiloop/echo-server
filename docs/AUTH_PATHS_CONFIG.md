# 认证路径配置指南

## 概述

Echo Server 支持通过命令行参数、环境变量或代码配置需要 HMAC 认证的路径。这使得认证策略可以灵活调整而无需修改代码。

## 配置方式

### 1. 命令行参数（推荐）

使用 `--auth-paths` 参数指定需要认证的路径列表，多个路径用逗号分隔。

```bash
./echo-server --auth-paths="/upload,/upload/*,/api/private/*"
```

### 2. 环境变量

设置 `AUTH_PATHS` 环境变量：

```bash
export AUTH_PATHS="/upload,/upload/*,/api/*"
./echo-server
```

### 3. 默认值

如果不指定，默认为：`/upload,/upload/*`

## 路径匹配规则

支持以下匹配模式：

### 精确匹配
```bash
--auth-paths="/upload"
```
- ✅ 匹配：`/upload`
- ❌ 不匹配：`/upload/file`, `/uploadfile`

### 前缀通配符
```bash
--auth-paths="/upload/*"
```
- ✅ 匹配：`/upload/file`, `/upload/dir/file.txt`
- ❌ 不匹配：`/upload`, `/uploadfile`

### 后缀通配符
```bash
--auth-paths="/api*"
```
- ✅ 匹配：`/api`, `/api/v1`, `/api/v2/users`
- ❌ 不匹配：`/v1/api`

## 使用示例

### 示例1：仅保护上传接口

**命令行：**
```bash
./echo-server --auth-paths="/upload,/upload/*"
```

**环境变量：**
```bash
export AUTH_PATHS="/upload,/upload/*"
./echo-server
```

**效果：**
- `/upload` - 需要认证 🔒
- `/upload/file` - 需要认证 🔒
- `/api/data` - 不需要认证 ✅
- `/health` - 不需要认证 ✅

### 示例2：保护多个API路径

```bash
./echo-server --auth-paths="/upload/*,/api/private/*,/admin/*"
```

**效果：**
- `/upload/file` - 需要认证 🔒
- `/api/private/data` - 需要认证 🔒
- `/admin/users` - 需要认证 🔒
- `/api/public/info` - 不需要认证 ✅

### 示例3：保护所有API但排除公开路径

由于 `--auth-paths` 只支持包含路径，要排除特定路径需要在代码中配置：

```go
e.Use(esmw.HMACAuth(esmw.HMACAuthConfig{
    Paths:     []string{"/api/*"},
    SkipPaths: []string{"/api/public/*"},
}))
```

### 示例4：禁用认证

设置为空值：

```bash
./echo-server --auth-paths=""
```

或使用环境变量：

```bash
export AUTH_PATHS=""
./echo-server
```

## 配置优先级

```
命令行参数 > 环境变量 > 默认值
```

### 优先级示例

```bash
# 环境变量设置
export AUTH_PATHS="/api/*"

# 命令行参数会覆盖环境变量
./echo-server --auth-paths="/upload/*"

# 实际生效：/upload/*
```

## 实时查看配置

启动服务器时会显示启用的认证路径：

```bash
$ ./echo-server --auth-paths="/upload/*,/api/*"

{"time":"...","level":"INFO","message":"Enabling HMAC auth for paths: [/upload/* /api/*]"}
```

## 验证配置

### 测试方法

**1. 访问需要认证的路径（无认证）**

```bash
curl -X POST http://localhost:9012/upload -F "file=@test.txt"
```

**预期结果：**
```json
{
  "error": "authentication failed"
}
```

**2. 访问需要认证的路径（有认证）**

```bash
API_KEY="your-secret-api-key-here"
TIMESTAMP=$(date +%s)
SIGNATURE=$(echo -n "$TIMESTAMP" | openssl dgst -sha256 -hmac "$API_KEY" | awk '{print $2}')

curl -X POST http://localhost:9012/upload \
  -H "X-Timestamp: $TIMESTAMP" \
  -H "X-Signature: $SIGNATURE" \
  -F "file=@test.txt"
```

**预期结果：**
```json
{
  "success": true,
  "filename": "test.txt",
  ...
}
```

**3. 访问不需要认证的路径**

```bash
curl http://localhost:9012/health
```

**预期结果：**
```
ok
```

## 高级配置

### 在代码中配置（更灵活）

如果需要更复杂的配置（如跳过路径、自定义错误处理），可以在 `main.go` 中使用完整配置：

```go
e.Use(esmw.HMACAuth(esmw.HMACAuthConfig{
    ApiKey:         os.Getenv("AUTH_API_KEY"),
    TimestampValid: 300,
    Paths:          cli.AuthPaths,  // 使用命令行配置
    SkipPaths:      []string{"/upload/public"},  // 跳过特定路径
    ErrorHandler: func(c echo.Context, err error) error {
        return c.JSON(401, map[string]string{
            "error": "unauthorized",
            "detail": err.Error(),
        })
    },
}))
```

### 动态路径配置

使用环境变量实现动态配置：

**开发环境：**
```bash
export AUTH_PATHS="/upload/*"
./echo-server
```

**测试环境：**
```bash
export AUTH_PATHS="/upload/*,/api/test/*"
./echo-server
```

**生产环境：**
```bash
export AUTH_PATHS="/upload/*,/api/*,/admin/*"
./echo-server
```

## Docker 部署示例

### Dockerfile

```dockerfile
FROM golang:1.21 AS builder
WORKDIR /app
COPY . .
RUN go build -o echo-server .

FROM debian:bookworm-slim
COPY --from=builder /app/echo-server /usr/local/bin/
ENTRYPOINT ["echo-server"]
```

### Docker 运行

```bash
# 使用默认认证路径
docker run -p 9012:9012 echo-server

# 自定义认证路径
docker run -p 9012:9012 echo-server --auth-paths="/upload/*,/api/*"

# 使用环境变量
docker run -p 9012:9012 -e AUTH_PATHS="/upload/*,/api/*" echo-server
```

### Docker Compose

```yaml
version: '3.8'
services:
  echo-server:
    image: echo-server
    ports:
      - "9012:9012"
      - "9013:9013"
    environment:
      - AUTH_PATHS=/upload/*,/api/private/*
      - AUTH_API_KEY=your-secret-key
      - HTTP_ADDR=0.0.0.0:9012
      - HTTPS_ADDR=0.0.0.0:9013
    command: 
      - --debug=false
      - --cert=/certs/cert.pem
      - --key=/certs/key.pem
    volumes:
      - ./certs:/certs
```

## Kubernetes 部署示例

### ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: echo-server-config
data:
  AUTH_PATHS: "/upload/*,/api/private/*"
  AUTH_API_KEY: "your-secret-key"
```

### Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: echo-server
spec:
  replicas: 3
  selector:
    matchLabels:
      app: echo-server
  template:
    metadata:
      labels:
        app: echo-server
    spec:
      containers:
      - name: echo-server
        image: echo-server:latest
        ports:
        - containerPort: 9012
        envFrom:
        - configMapRef:
            name: echo-server-config
        args:
        - --debug=false
```

## 安全最佳实践

### 1. 最小权限原则

只为必要的路径启用认证：

```bash
# ✅ 好 - 只保护需要的路径
--auth-paths="/upload/*,/api/admin/*"

# ❌ 不好 - 过度保护
--auth-paths="/*"
```

### 2. 使用环境变量管理敏感信息

```bash
# ✅ 好 - 使用环境变量
export AUTH_API_KEY="$(openssl rand -hex 32)"
export AUTH_PATHS="/upload/*,/api/*"

# ❌ 不好 - 硬编码在命令行（可能被历史记录泄露）
./echo-server --auth-paths="/upload/*"
```

### 3. 定期审查认证路径

建立定期审查机制，确保认证配置符合安全要求。

### 4. 日志监控

监控认证失败的请求：

```bash
# 查看认证失败的日志
grep "auth failed" /var/log/echo-server.log
```

## 故障排查

### 问题1：认证不生效

**症状：** 访问应该需要认证的路径却不需要认证

**检查：**
```bash
# 1. 查看启动日志
./echo-server 2>&1 | grep "Enabling HMAC"

# 2. 验证配置
echo $AUTH_PATHS

# 3. 测试认证
curl -v http://localhost:9012/upload
# 应该返回 401
```

**解决：**
- 确认路径配置正确
- 检查路径匹配规则
- 查看服务器日志

### 问题2：路径不匹配

**症状：** 配置了认证路径但某些请求仍然不需要认证

**原因：** 路径匹配规则不正确

**示例：**
```bash
# 配置
--auth-paths="/upload"

# /upload - 需要认证 ✅
# /upload/file - 不需要认证 ❌

# 正确配置应该是
--auth-paths="/upload,/upload/*"
```

### 问题3：环境变量不生效

**检查步骤：**
```bash
# 1. 确认环境变量已设置
echo $AUTH_PATHS

# 2. 确认没有命令行参数覆盖
ps aux | grep echo-server

# 3. 重新启动服务
./echo-server
```

## 参考

- [HMAC 认证中间件文档](../middleware/AUTH_MIDDLEWARE.md)
- [Kong CLI 文档](KONG_CLI_REFACTORING.md)
- [上传 API 文档](UPLOAD_API.md)

## 总结

通过命令行参数配置认证路径提供了以下优势：

✅ **灵活性** - 无需修改代码即可调整认证策略  
✅ **可移植性** - 不同环境使用不同配置  
✅ **可维护性** - 集中管理认证配置  
✅ **安全性** - 环境变量管理敏感配置  
✅ **可观察性** - 启动时显示生效的配置  

建议在生产环境中使用环境变量配置，在开发和测试中使用命令行参数快速调整。

