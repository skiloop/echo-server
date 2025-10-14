# 文件上传API - 快速启动指南

## 📦 已实现功能

✅ **文件上传到指定目录**
- 支持通过query参数或环境变量指定上传目录
- 自动创建上传目录
- 文件名清理，防止路径遍历攻击

✅ **API Key认证（独立中间件）**  
- 使用HMAC-SHA256增强hash认证
- 时间戳验证，防重放攻击
- 支持环境变量配置
- **可配置哪些路径需要认证**
- **可复用于其他需要认证的API**

✅ **安全特性**
- 文件大小限制（默认10MB）
- 路径遍历攻击防护
- 时间戳过期验证（默认5分钟）
- 敏感配置支持环境变量

## 🚀 快速开始

### 1. 查看上传路由和认证中间件（已添加）

上传路由和HMAC认证中间件已经在 `main.go` 中配置：

```go
// main.go

// HMAC认证中间件 - 仅对指定路径生效
e.Use(esmw.HMACAuthForPaths("/upload", "/upload/*"))

// file upload routes (需要HMAC认证)
routers.SetUploadRouters(e)
```

**说明：**
- 认证已抽取为独立中间件 `middleware/auth.go`
- 可以灵活配置哪些路径需要认证
- 详见 [认证中间件文档](../middleware/AUTH_MIDDLEWARE.md)

### 2. 配置API Key

**开发环境：**
```bash
# 使用默认配置（不安全，仅供测试）
# API Key: "your-secret-api-key-here"
```

**生产环境（推荐）：**
```bash
# 设置环境变量
export AUTH_API_KEY="your-strong-random-api-key-32-chars-or-more"
export UPLOAD_DIR="./uploads"
export UPLOAD_MAX_SIZE="10485760"  # 10MB
```

或使用配置文件：
```bash
cp upload.env.example upload.env
vi upload.env  # 修改配置
source upload.env
```

### 3. 启动服务器

```bash
# HTTP模式
go run main.go -http 0.0.0.0:9012

# 或者 HTTPS模式（需要证书）
go run main.go -http 0.0.0.0:9012 -https 0.0.0.0:9013 -cert cert.pem -key key.pem
```

### 4. 测试上传

#### 方式1: 使用测试脚本（最快）

```bash
cd examples
./test_upload.sh
```

#### 方式2: 使用Go客户端

```bash
go run examples/upload_client.go examples/test.txt
```

#### 方式3: 使用Python客户端

```bash
pip install requests
python examples/upload_client.py examples/test.txt
```

#### 方式4: 使用curl

```bash
# 设置变量
API_KEY="your-secret-api-key-here"
TIMESTAMP=$(date +%s)
SIGNATURE=$(echo -n "$TIMESTAMP" | openssl dgst -sha256 -hmac "$API_KEY" | awk '{print $2}')

# 上传文件
curl -X POST http://localhost:9012/upload \
  -H "X-Timestamp: $TIMESTAMP" \
  -H "X-Signature: $SIGNATURE" \
  -F "file=@examples/test.txt"
```

## 📋 API端点

### POST /upload

**请求头：**
- `X-Timestamp`: Unix时间戳（秒）
- `X-Signature`: HMAC-SHA256(api_key, timestamp)

**请求体：**
- `multipart/form-data` 格式
- 字段名：`file`

**查询参数（可选）：**
- `dir`: 指定上传目录

**成功响应 (200)：**
```json
{
  "success": true,
  "filename": "test.txt",
  "size": 1024,
  "path": "./uploads/test.txt"
}
```

## 🔐 认证流程

### 客户端步骤：

1. 生成当前Unix时间戳
2. 使用API Key和时间戳计算HMAC-SHA256签名
3. 在请求头中包含时间戳和签名
4. 上传文件

### 签名计算示例：

**Go:**
```go
timestamp := strconv.FormatInt(time.Now().Unix(), 10)
h := hmac.New(sha256.New, []byte(apiKey))
h.Write([]byte(timestamp))
signature := hex.EncodeToString(h.Sum(nil))
```

**Python:**
```python
import hmac, hashlib, time
timestamp = str(int(time.time()))
signature = hmac.new(
    api_key.encode(), 
    timestamp.encode(), 
    hashlib.sha256
).hexdigest()
```

**Bash:**
```bash
TIMESTAMP=$(date +%s)
SIGNATURE=$(echo -n "$TIMESTAMP" | openssl dgst -sha256 -hmac "$API_KEY" | awk '{print $2}')
```

## 📁 项目文件结构

```
echo-server/
├── middleware/
│   ├── auth.go                # HMAC认证中间件 ⭐ 新增
│   └── AUTH_MIDDLEWARE.md     # 认证中间件文档 ⭐ 新增
├── routers/
│   └── upload.go              # 上传API实现（已简化）
├── examples/
│   ├── upload_client.go       # Go客户端示例
│   ├── upload_client.py       # Python客户端示例
│   ├── test_upload.sh         # Bash测试脚本
│   ├── test.txt               # 测试文件
│   └── README.md              # 示例说明
├── main.go                    # 服务器入口（已配置认证中间件）
├── upload.env.example         # 环境变量配置示例
├── UPLOAD_API.md              # 详细API文档
└── UPLOAD_QUICKSTART.md       # 本文件
```

## ⚙️ 配置选项

### 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `AUTH_API_KEY` | API密钥 | `your-secret-api-key-here` |
| `UPLOAD_DIR` | 上传目录 | `./uploads` |
| `UPLOAD_MAX_SIZE` | 最大文件大小（字节） | `10485760` (10MB) |
| `ECHO_AUTH_TIMESTAMP_VALID` | 时间戳有效期（秒） | `300` (5分钟) |

### 代码常量

在 `routers/upload.go` 中：
- `DefaultUploadDir`: 默认上传目录
- `DefaultApiKey`: 默认API Key
- `DefaultTimestampValidDuration`: 默认时间戳有效期
- `DefaultMaxFileSize`: 默认最大文件大小

## 🔒 安全建议

### ⚠️ 必须做的：

1. **更改默认API Key**
   ```bash
   export AUTH_API_KEY="$(openssl rand -hex 32)"
   ```

2. **使用HTTPS**
   - 生产环境必须使用HTTPS
   - 防止签名在传输中被窃取

3. **时间同步**
   - 确保服务器和客户端时间同步
   - 使用NTP服务

### 💡 建议做的：

1. **添加文件类型验证**
2. **添加病毒扫描**
3. **实施访问日志**
4. **定期轮换API Key**
5. **设置合理的文件大小限制**

## 📝 常见问题

### Q: 认证失败怎么办？
A: 检查：
1. API Key是否一致（服务器和客户端）
2. 时间戳是否在有效期内
3. 签名计算是否正确

### Q: 如何上传到自定义目录？
A: 三种方式：
1. Query参数：`/upload?dir=./my_dir`
2. 环境变量：`export UPLOAD_DIR=./my_dir`
3. 修改代码：`DefaultUploadDir = "./my_dir"`

### Q: 如何增加文件大小限制？
A: 设置环境变量：
```bash
export UPLOAD_MAX_SIZE="52428800"  # 50MB
```

### Q: 如何查看上传日志？
A: 服务器会输出详细日志：
```
INFO file uploaded successfully: ./uploads/test.txt, size: 1024
```

## 📚 更多文档

- [UPLOAD_API.md](UPLOAD_API.md) - 完整API文档
- [middleware/AUTH_MIDDLEWARE.md](middleware/AUTH_MIDDLEWARE.md) - 认证中间件文档 ⭐
- [examples/README.md](examples/README.md) - 客户端示例说明
- [upload.env.example](upload.env.example) - 环境变量配置示例

## 🔧 自定义认证路径

如果你想为其他API也添加HMAC认证，可以在 `main.go` 中配置：

```go
// 为多个路径启用认证
e.Use(esmw.HMACAuthForPaths("/upload", "/api/private/*", "/admin/*"))

// 或使用更灵活的配置
e.Use(esmw.HMACAuth(esmw.HMACAuthConfig{
    ApiKey:         os.Getenv("AUTH_API_KEY"),
    TimestampValid: 300,
    Paths:          []string{"/upload", "/api/private/*"},
    SkipPaths:      []string{"/upload/public"},
}))
```

详细用法请参考 [middleware/AUTH_MIDDLEWARE.md](middleware/AUTH_MIDDLEWARE.md)

## 🎉 完成！

现在你已经成功设置了文件上传API，可以开始上传文件了！

有任何问题，请查阅详细文档或检查服务器日志。

