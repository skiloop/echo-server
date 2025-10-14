# 文件上传API使用说明

## 功能概述

这个API提供了安全的文件上传功能，包括：
- 文件上传到指定目录
- 基于HMAC-SHA256的API Key认证
- 时间戳防重放攻击保护
- 可自定义上传目录

## API端点

```
POST /upload
```

## 认证机制

### 认证方式
使用HMAC-SHA256增强hash进行认证，客户端需要在请求头中提供：

- `X-Timestamp`: Unix时间戳（秒）
- `X-Signature`: HMAC-SHA256签名

### 签名计算方法

```
signature = HMAC-SHA256(api_key, timestamp)
```

### 防重放攻击
- 时间戳有效期：5分钟（300秒）
- 超过有效期的请求会被拒绝

## 请求格式

### 请求头
```
Content-Type: multipart/form-data
X-Timestamp: <unix_timestamp>
X-Signature: <hmac_sha256_signature>
```

### 请求体
使用 `multipart/form-data` 格式上传文件：
- 字段名：`file`
- 字段值：文件内容

### 查询参数（可选）
- `dir`: 指定上传目录，默认为 `./uploads`

## 响应格式

### 成功响应 (200 OK)
```json
{
  "success": true,
  "filename": "example.txt",
  "size": 1024,
  "path": "./uploads/example.txt"
}
```

### 失败响应

#### 认证失败 (401 Unauthorized)
```json
{
  "error": "authentication failed"
}
```

#### 无文件上传 (400 Bad Request)
```json
{
  "error": "no file uploaded"
}
```

#### 服务器错误 (500 Internal Server Error)
```json
{
  "error": "failed to save file"
}
```

## 配置

### 服务器端配置

支持两种配置方式：环境变量（推荐）或代码中的默认值。

#### 方式1: 使用环境变量（推荐）

```bash
# 复制示例配置文件
cp upload.env.example upload.env

# 编辑配置文件
vi upload.env

# 加载环境变量
source upload.env

# 或者直接设置
export AUTH_API_KEY="your-secret-api-key-here"
export UPLOAD_DIR="./uploads"
export UPLOAD_MAX_SIZE="10485760"  # 10MB
export ECHO_AUTH_TIMESTAMP_VALID="300"  # 5分钟
```

支持的环境变量：
- `AUTH_API_KEY`: API密钥（必须设置）
- `UPLOAD_DIR`: 上传目录（可选，默认 `./uploads`）
- `UPLOAD_MAX_SIZE`: 最大文件大小，单位字节（可选，默认 10MB）
- `ECHO_AUTH_TIMESTAMP_VALID`: 时间戳有效期，单位秒（可选，默认 300秒）

#### 方式2: 修改默认值

在 `routers/upload.go` 中修改以下常量：

```go
const (
    // 默认上传目录
    DefaultUploadDir = "./uploads"
    
    // 默认API Key
    DefaultApiKey = "your-secret-api-key-here"
    
    // 默认时间戳有效期（秒）
    DefaultTimestampValidDuration = 300
    
    // 默认最大文件大小（字节）
    DefaultMaxFileSize = 10 << 20  // 10MB
)
```

**生产环境建议：**
- ✅ 优先使用环境变量配置
- ✅ 使用强随机字符串作为 API Key（至少32字符）
- ✅ 定期轮换 API Key
- ✅ 根据需要调整时间戳有效期和文件大小限制
- ⚠️ 不要将敏感信息提交到代码仓库

## 使用示例

### Go客户端示例

示例代码位于 `examples/upload_client.go`

```bash
# 编译并运行
go run examples/upload_client.go <file_path> [upload_dir]

# 示例
go run examples/upload_client.go test.txt
go run examples/upload_client.go test.txt ./custom_uploads
```

### Python客户端示例

示例代码位于 `examples/upload_client.py`

```bash
# 安装依赖
pip install requests

# 运行
python examples/upload_client.py <file_path> [upload_dir]

# 示例
python examples/upload_client.py test.txt
python examples/upload_client.py test.txt ./custom_uploads
```

### curl示例

```bash
# 1. 生成时间戳
TIMESTAMP=$(date +%s)

# 2. 计算签名（需要安装openssl）
API_KEY="your-secret-api-key-here"
SIGNATURE=$(echo -n "$TIMESTAMP" | openssl dgst -sha256 -hmac "$API_KEY" | awk '{print $2}')

# 3. 上传文件
curl -X POST http://localhost:9012/upload \
  -H "X-Timestamp: $TIMESTAMP" \
  -H "X-Signature: $SIGNATURE" \
  -F "file=@test.txt"

# 4. 指定上传目录
curl -X POST "http://localhost:9012/upload?dir=./my_uploads" \
  -H "X-Timestamp: $TIMESTAMP" \
  -H "X-Signature: $SIGNATURE" \
  -F "file=@test.txt"
```

## 安全考虑

1. **API Key安全**
   - 不要将API Key硬编码在代码中
   - 使用环境变量或安全的配置管理
   - 定期轮换API Key

2. **HTTPS**
   - 生产环境必须使用HTTPS
   - 防止中间人攻击和签名泄露

3. **时间同步**
   - 确保客户端和服务器时间同步
   - 使用NTP服务

4. **文件验证**
   - 可以添加文件类型验证
   - 可以添加文件大小限制
   - 可以添加病毒扫描

5. **目录安全**
   - 验证上传目录路径，防止路径遍历攻击
   - 设置适当的文件权限

## 扩展功能建议

1. **文件覆盖处理**
   - 自动重命名重复文件
   - 或返回错误

2. **多文件上传**
   - 支持批量上传

3. **文件类型限制**
   - 白名单/黑名单

4. **进度回调**
   - 大文件上传进度

5. **断点续传**
   - 支持大文件断点续传

