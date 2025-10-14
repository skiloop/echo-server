# 文件上传示例

这个目录包含文件上传API的使用示例和测试工具。

## 文件说明

### 客户端示例

1. **upload_client.go** - Go语言客户端示例
   - 演示如何使用Go上传文件
   - 包含HMAC-SHA256签名计算
   - 支持自定义上传目录

2. **upload_client.py** - Python客户端示例
   - 演示如何使用Python上传文件
   - 使用requests库
   - 包含完整的认证流程

### 测试工具

3. **test_upload.sh** - Bash测试脚本
   - 使用curl测试上传功能
   - 自动计算签名
   - 显示详细的请求和响应信息

4. **test.txt** - 测试文件
   - 用于测试上传的示例文件

## 快速开始

### 1. 启动服务器

确保echo-server正在运行：

```bash
# 在项目根目录
go run main.go -http 0.0.0.0:9012
```

### 2. 配置API Key

在使用客户端之前，确保API Key与服务器端一致。

**服务器端配置** (在 `routers/upload.go`):
```go
ApiKey = "your-secret-api-key-here"
```

**客户端配置** (在各个客户端文件中修改对应的API_KEY):
- `upload_client.go`: 第12行
- `upload_client.py`: 第11行
- `test_upload.sh`: 第5行

### 3. 运行测试

#### 使用Bash脚本测试

```bash
cd examples
./test_upload.sh
```

#### 使用Go客户端

```bash
# 上传到默认目录
go run examples/upload_client.go examples/test.txt

# 上传到指定目录
go run examples/upload_client.go examples/test.txt ./my_uploads
```

#### 使用Python客户端

```bash
# 安装依赖
pip install requests

# 上传文件
python examples/upload_client.py examples/test.txt

# 上传到指定目录
python examples/upload_client.py examples/test.txt ./my_uploads
```

## API使用说明

详细的API文档请参考项目根目录的 [UPLOAD_API.md](../UPLOAD_API.md)

### 核心认证流程

1. 生成Unix时间戳
2. 使用HMAC-SHA256计算签名: `signature = HMAC(api_key, timestamp)`
3. 在请求头中包含:
   - `X-Timestamp`: 时间戳
   - `X-Signature`: 签名
4. 使用multipart/form-data上传文件

### 示例请求

```bash
POST /upload HTTP/1.1
Host: localhost:9012
X-Timestamp: 1697212800
X-Signature: abc123def456...
Content-Type: multipart/form-data; boundary=----WebKitFormBoundary

------WebKitFormBoundary
Content-Disposition: form-data; name="file"; filename="test.txt"
Content-Type: text/plain

[文件内容]
------WebKitFormBoundary--
```

## 安全提示

⚠️ **重要提示**:

1. **不要在生产环境使用默认API Key**
   - 使用环境变量或配置文件管理API Key
   - 定期轮换API Key

2. **使用HTTPS**
   - 生产环境必须使用HTTPS加密传输
   - 防止签名在传输过程中被窃取

3. **时间同步**
   - 确保客户端和服务器时间同步
   - 使用NTP服务保持时间准确

4. **文件验证**
   - 建议添加文件类型白名单
   - 添加文件大小限制
   - 考虑添加病毒扫描

## 故障排查

### 🔧 快速修复

如果上传失败，请先阅读：

📖 **[QUICK_FIX.md](QUICK_FIX.md)** - 认证问题快速修复指南

或运行诊断工具：

```bash
./debug_auth.sh
```

### 常见问题

#### 认证失败 (401)

**症状**：
```
HTTP状态码: 401
响应内容: {"error":"authentication failed"}
```

**快速解决**：
```bash
# 1. 运行诊断
./debug_auth.sh

# 2. 确保 API Key 一致
export AUTH_API_KEY="your-secret-api-key-here"

# 3. 重新测试
./test_upload.sh
```

**详细排查**：参见 [TROUBLESHOOTING.md](TROUBLESHOOTING.md)

#### Python 成功但 Shell 失败

**原因**：API Key 不一致

**解决**：
```bash
# 确保环境变量设置
export AUTH_API_KEY="your-secret-api-key-here"

# 验证
echo $AUTH_API_KEY

# 测试
./test_upload.sh
```

#### 文件未找到 (400)

- 确保使用正确的表单字段名 "file"
- 检查文件路径是否正确

#### 服务器错误 (500)

- 检查上传目录权限
- 查看服务器日志获取详细错误信息

### 📚 完整文档

- [QUICK_FIX.md](QUICK_FIX.md) - 快速修复指南 ⭐
- [TROUBLESHOOTING.md](TROUBLESHOOTING.md) - 详细故障排查

## 扩展开发

基于这些示例，你可以：

1. 添加文件类型验证
2. 实现文件大小限制
3. 添加上传进度显示
4. 实现批量上传
5. 添加断点续传功能
6. 集成到你的应用中

