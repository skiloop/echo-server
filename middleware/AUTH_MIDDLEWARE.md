# HMAC认证中间件使用指南

## 概述

`HMACAuth` 是一个可复用的Echo中间件，提供基于HMAC-SHA256的API认证功能。支持灵活配置需要认证的路径，可用于保护API端点。

## 特性

✅ **HMAC-SHA256认证** - 安全的签名验证机制  
✅ **时间戳防重放** - 防止重放攻击  
✅ **灵活的路径配置** - 支持指定需要/跳过认证的路径  
✅ **路径匹配** - 支持精确匹配和通配符匹配  
✅ **环境变量配置** - 支持从环境变量读取配置  
✅ **自定义错误处理** - 可自定义认证失败的响应

## 快速开始

### 基本用法

```go
import (
    "github.com/labstack/echo/v4"
    esmw "github.com/skiloop/echo-server/middleware"
)

func main() {
    e := echo.New()
    
    // 对所有路径启用认证
    e.Use(esmw.HMACAuth(esmw.HMACAuthConfig{}))
    
    // ... 设置路由
}
```

### 指定路径认证

```go
// 仅对 /upload 路径启用认证
e.Use(esmw.HMACAuthForPaths("/upload"))

// 对多个路径启用认证
e.Use(esmw.HMACAuthForPaths("/upload", "/api/private/*"))
```

### 完整配置

```go
e.Use(esmw.HMACAuth(esmw.HMACAuthConfig{
    // API Key，留空则从环境变量 AUTH_API_KEY 读取
    ApiKey: "your-secret-api-key",
    
    // 时间戳有效期（秒），0则使用默认300秒
    TimestampValid: 300,
    
    // 需要认证的路径
    Paths: []string{
        "/upload",
        "/api/private/*",
    },
    
    // 跳过认证的路径（优先级更高）
    SkipPaths: []string{
        "/upload/public",
    },
    
    // 自定义错误处理
    ErrorHandler: func(c echo.Context, err error) error {
        return c.JSON(401, map[string]string{
            "error": "unauthorized",
            "detail": err.Error(),
        })
    },
}))
```

## 路径匹配规则

### 精确匹配
```go
Paths: []string{"/upload"}
```
只匹配 `/upload`，不匹配 `/upload/file` 或 `/uploadfile`

### 前缀通配符
```go
Paths: []string{"/upload/*"}
```
匹配 `/upload/` 开头的所有路径：
- ✅ `/upload/file`
- ✅ `/upload/image/test.jpg`
- ❌ `/upload`
- ❌ `/uploadfile`

### 后缀通配符
```go
Paths: []string{"/api*"}
```
匹配 `/api` 开头的所有路径：
- ✅ `/api`
- ✅ `/api/v1`
- ✅ `/api/v2/users`

## 配置优先级

### API Key
```
HMACAuthConfig.ApiKey > 环境变量 AUTH_API_KEY > DefaultHMACApiKey
```

### 时间戳有效期
```
HMACAuthConfig.TimestampValid > 环境变量 ECHO_AUTH_TIMESTAMP_VALID > DefaultHMACTimestampValid (300秒)
```

### 路径认证逻辑
1. 如果路径在 `SkipPaths` 中 → 跳过认证
2. 如果 `Paths` 为空 → 所有路径都需要认证
3. 如果路径在 `Paths` 中 → 需要认证
4. 否则 → 跳过认证

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `AUTH_API_KEY` | API密钥 | `your-secret-api-key-here` |
| `ECHO_AUTH_TIMESTAMP_VALID` | 时间戳有效期（秒） | `300` |

## 客户端实现

### 认证流程

1. 生成当前Unix时间戳
2. 使用API Key和时间戳计算HMAC-SHA256签名
3. 在请求头中包含 `X-Timestamp` 和 `X-Signature`

### Go客户端示例

```go
import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "strconv"
    "time"
)

func makeAuthHeaders(apiKey string) map[string]string {
    // 生成时间戳
    timestamp := strconv.FormatInt(time.Now().Unix(), 10)
    
    // 计算签名
    h := hmac.New(sha256.New, []byte(apiKey))
    h.Write([]byte(timestamp))
    signature := hex.EncodeToString(h.Sum(nil))
    
    return map[string]string{
        "X-Timestamp": timestamp,
        "X-Signature": signature,
    }
}

// 使用
headers := makeAuthHeaders("your-secret-api-key")
// 在HTTP请求中设置headers
```

### Python客户端示例

```python
import hmac
import hashlib
import time

def make_auth_headers(api_key):
    # 生成时间戳
    timestamp = str(int(time.time()))
    
    # 计算签名
    signature = hmac.new(
        api_key.encode(),
        timestamp.encode(),
        hashlib.sha256
    ).hexdigest()
    
    return {
        'X-Timestamp': timestamp,
        'X-Signature': signature
    }

# 使用
headers = make_auth_headers('your-secret-api-key')
# 在HTTP请求中设置headers
```

### curl示例

```bash
#!/bin/bash
API_KEY="your-secret-api-key"
TIMESTAMP=$(date +%s)
SIGNATURE=$(echo -n "$TIMESTAMP" | openssl dgst -sha256 -hmac "$API_KEY" | awk '{print $2}')

curl -X POST http://localhost:9012/upload \
  -H "X-Timestamp: $TIMESTAMP" \
  -H "X-Signature: $SIGNATURE" \
  -F "file=@test.txt"
```

## 使用场景

### 场景1: 保护上传接口

```go
// 仅对上传路径启用认证
e.Use(esmw.HMACAuthForPaths("/upload", "/upload/*"))

// 设置路由
e.POST("/upload", uploadHandler)
e.GET("/files", listFilesHandler)  // 不需要认证
```

### 场景2: 保护私有API

```go
// 对所有 /api/private 开头的路径启用认证
e.Use(esmw.HMACAuthForPaths("/api/private/*"))

// 设置路由
e.GET("/api/public/users", publicUsersHandler)    // 不需要认证
e.GET("/api/private/admin", privateAdminHandler)  // 需要认证
```

### 场景3: 除了特定路径都需要认证

```go
// 对所有路径启用认证，但跳过公开路径
e.Use(esmw.HMACAuth(esmw.HMACAuthConfig{
    Paths: []string{},  // 空表示所有路径
    SkipPaths: []string{
        "/",
        "/health",
        "/api/public/*",
    },
}))
```

### 场景4: 自定义错误响应

```go
e.Use(esmw.HMACAuth(esmw.HMACAuthConfig{
    Paths: []string{"/upload"},
    ErrorHandler: func(c echo.Context, err error) error {
        // 记录认证失败日志
        c.Logger().Warnf("Auth failed: %v, IP: %s", err, c.RealIP())
        
        // 返回自定义错误
        return c.JSON(http.StatusUnauthorized, map[string]interface{}{
            "code":    401,
            "message": "Authentication failed",
            "detail":  err.Error(),
            "timestamp": time.Now().Unix(),
        })
    },
}))
```

## 安全建议

### ✅ 推荐做法

1. **使用强密钥**
   ```bash
   # 生成32字节随机密钥
   openssl rand -hex 32
   ```

2. **使用环境变量**
   ```bash
   export AUTH_API_KEY="$(openssl rand -hex 32)"
   ```

3. **使用HTTPS**
   - 生产环境必须使用HTTPS
   - 防止签名在传输中被窃取

4. **时间同步**
   - 使用NTP保持服务器时间准确
   - 客户端也要保持时间同步

5. **合理的时间窗口**
   - 默认5分钟是合理的
   - 太短：网络延迟可能导致失败
   - 太长：重放攻击风险增加

### ❌ 避免的做法

1. **不要硬编码API Key**
   ```go
   // ❌ 不好
   ApiKey: "my-secret-key"
   
   // ✅ 好
   ApiKey: os.Getenv("AUTH_API_KEY")
   ```

2. **不要在HTTP中使用**
   - 生产环境必须使用HTTPS

3. **不要使用弱密钥**
   - 不要使用短密钥或可预测的密钥
   - 至少32字符随机生成

## 便捷函数

### HMACAuthForPaths

为指定路径快速创建认证中间件：

```go
// 使用默认配置，仅指定路径
e.Use(esmw.HMACAuthForPaths("/upload", "/api/private/*"))
```

等价于：

```go
e.Use(esmw.HMACAuth(esmw.HMACAuthConfig{
    Paths: []string{"/upload", "/api/private/*"},
}))
```

### HMACAuthWithConfig

使用自定义配置快速创建认证中间件：

```go
e.Use(esmw.HMACAuthWithConfig(
    "your-api-key",
    600,  // 10分钟
    []string{"/upload"},
))
```

等价于：

```go
e.Use(esmw.HMACAuth(esmw.HMACAuthConfig{
    ApiKey:         "your-api-key",
    TimestampValid: 600,
    Paths:          []string{"/upload"},
}))
```

## 错误处理

### 认证失败情况

| 错误 | 原因 | 解决方案 |
|------|------|----------|
| `missing authentication headers` | 缺少 X-Timestamp 或 X-Signature | 确保请求头包含认证信息 |
| `invalid timestamp format` | 时间戳格式错误 | 使用Unix时间戳（秒） |
| `timestamp expired or invalid` | 时间戳过期 | 检查时间同步，重新生成时间戳 |
| `invalid signature` | 签名不匹配 | 检查API Key是否一致，签名算法是否正确 |

### 默认错误响应

```json
{
  "error": "authentication failed"
}
```

HTTP状态码：`401 Unauthorized`

## 测试

### 单元测试示例

```go
func TestHMACAuth(t *testing.T) {
    e := echo.New()
    req := httptest.NewRequest(http.MethodGet, "/upload", nil)
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)
    
    // 设置认证头
    timestamp := strconv.FormatInt(time.Now().Unix(), 10)
    signature := calculateHMACSignature("test-key", timestamp)
    req.Header.Set("X-Timestamp", timestamp)
    req.Header.Set("X-Signature", signature)
    
    // 创建中间件
    middleware := HMACAuth(HMACAuthConfig{
        ApiKey: "test-key",
        Paths:  []string{"/upload"},
    })
    
    handler := middleware(func(c echo.Context) error {
        return c.String(http.StatusOK, "success")
    })
    
    // 执行测试
    err := handler(c)
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, rec.Code)
}
```

## 常见问题

### Q: 如何为不同的路径设置不同的API Key？

A: 可以创建多个中间件实例：

```go
// 上传路径使用一个key
uploadAuth := esmw.HMACAuth(esmw.HMACAuthConfig{
    ApiKey: "upload-key",
    Paths:  []string{"/upload/*"},
})

// 管理路径使用另一个key
adminAuth := esmw.HMACAuth(esmw.HMACAuthConfig{
    ApiKey: "admin-key",
    Paths:  []string{"/admin/*"},
})

e.Use(uploadAuth)
e.Use(adminAuth)
```

### Q: 如何调试认证失败？

A: 使用自定义错误处理记录详细信息：

```go
ErrorHandler: func(c echo.Context, err error) error {
    c.Logger().Errorf("Auth failed: %v", err)
    c.Logger().Errorf("Timestamp: %s", c.Request().Header.Get("X-Timestamp"))
    c.Logger().Errorf("Signature: %s", c.Request().Header.Get("X-Signature"))
    return c.JSON(401, map[string]string{"error": err.Error()})
}
```

### Q: 中间件的性能如何？

A: HMAC-SHA256计算非常快，对性能影响minimal。主要开销是时间戳验证和签名比较，都是O(1)操作。

## 参考

- [Echo中间件文档](https://echo.labstack.com/middleware)
- [HMAC标准 (RFC 2104)](https://www.rfc-editor.org/rfc/rfc2104)
- [Go crypto/hmac文档](https://pkg.go.dev/crypto/hmac)

