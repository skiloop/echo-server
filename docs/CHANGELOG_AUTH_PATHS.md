# 认证路径配置功能更新

## 版本信息

- **更新日期**: 2025-10-14
- **功能**: 支持从命令行配置认证路径
- **影响范围**: HMAC认证中间件

## 新特性

### 命令行参数 `--auth-paths`

现在可以通过命令行参数或环境变量配置需要 HMAC 认证的路径，无需修改代码。

**参数名称**: `--auth-paths`  
**环境变量**: `AUTH_PATHS`  
**默认值**: `/upload,/upload/*`  
**格式**: 逗号分隔的路径列表，支持通配符

## 使用方法

### 1. 命令行参数

```bash
# 自定义认证路径
./echo-server --auth-paths="/upload/*,/api/private/*,/admin/*"

# 禁用认证
./echo-server --auth-paths=""
```

### 2. 环境变量

```bash
export AUTH_PATHS="/upload/*,/api/*"
./echo-server
```

### 3. 组合使用

```bash
# 命令行参数优先级更高
AUTH_PATHS="/api/*" ./echo-server --auth-paths="/upload/*"
# 实际生效: /upload/*
```

## 路径匹配

支持以下匹配模式：

- **精确匹配**: `/upload` - 只匹配 `/upload`
- **前缀通配符**: `/upload/*` - 匹配 `/upload/` 开头的所有路径
- **后缀通配符**: `/api*` - 匹配 `/api` 开头的所有路径

## 示例

### 示例1：保护上传和私有API

```bash
./echo-server --auth-paths="/upload/*,/api/private/*"
```

**效果**:
- `/upload/file` → 需要认证 🔒
- `/api/private/data` → 需要认证 🔒
- `/api/public/info` → 不需要认证 ✅
- `/health` → 不需要认证 ✅

### 示例2：生产环境配置

```bash
export AUTH_PATHS="/upload/*,/api/*,/admin/*"
export AUTH_API_KEY="$(openssl rand -hex 32)"
./echo-server --debug=false
```

### 示例3：Docker 部署

```yaml
# docker-compose.yml
services:
  echo-server:
    image: echo-server
    environment:
      - AUTH_PATHS=/upload/*,/api/*
      - AUTH_API_KEY=${API_KEY}
    command: --debug=false
```

## 配置验证

启动服务器时会显示启用的认证路径：

```bash
$ ./echo-server --auth-paths="/upload/*,/api/*"

{"time":"...","level":"INFO","message":"Enabling HMAC auth for paths: [/upload/* /api/*]"}
```

## 兼容性

- ✅ **向后兼容**: 默认值保持不变
- ✅ **无破坏性变更**: 现有部署无需修改
- ✅ **灵活配置**: 支持运行时配置

## 代码变更

### main.go

**添加参数定义**:
```go
type CLI struct {
    // ... 其他参数
    AuthPaths []string `help:"Paths requiring HMAC authentication (supports wildcards)" default:"/upload,/upload/*" env:"AUTH_PATHS" sep:","`
}
```

**使用配置的路径**:
```go
if len(cli.AuthPaths) > 0 {
    e.Logger.Infof("Enabling HMAC auth for paths: %v", cli.AuthPaths)
    e.Use(esmw.HMACAuthForPaths(cli.AuthPaths...))
}
```

## 文档更新

- ✅ [README.md](../README.md) - 添加参数说明
- ✅ [docs/AUTH_PATHS_CONFIG.md](AUTH_PATHS_CONFIG.md) - 完整配置指南
- ✅ [docs/KONG_CLI_REFACTORING.md](KONG_CLI_REFACTORING.md) - Kong文档更新

## 测试

### 测试1: 默认配置

```bash
$ ./echo-server
# 输出: Enabling HMAC auth for paths: [/upload /upload/*]
```

✅ 通过

### 测试2: 自定义路径

```bash
$ ./echo-server --auth-paths="/api/*,/admin/*"
# 输出: Enabling HMAC auth for paths: [/api/* /admin/*]
```

✅ 通过

### 测试3: 环境变量

```bash
$ AUTH_PATHS="/private/*" ./echo-server
# 输出: Enabling HMAC auth for paths: [/private/*]
```

✅ 通过

### 测试4: 禁用认证

```bash
$ ./echo-server --auth-paths=""
# 无认证中间件启用日志
```

✅ 通过

## 好处

1. **灵活性** - 无需修改代码即可调整认证策略
2. **可移植性** - 不同环境使用不同配置
3. **可维护性** - 集中管理认证配置
4. **安全性** - 环境变量管理敏感配置
5. **可观察性** - 启动时显示生效的配置

## 迁移指南

### 从硬编码迁移

**之前**:
```go
// main.go
e.Use(esmw.HMACAuthForPaths("/upload", "/upload/*"))
```

**现在**:
```bash
# 命令行配置
./echo-server --auth-paths="/upload,/upload/*"

# 或环境变量
export AUTH_PATHS="/upload,/upload/*"
./echo-server
```

### 无需迁移

如果你满意默认配置（`/upload,/upload/*`），无需任何操作。

## 未来增强

计划中的改进：

- [ ] 支持正则表达式匹配
- [ ] 支持从配置文件读取
- [ ] 支持动态重载配置
- [ ] 支持路径黑名单
- [ ] 支持不同路径使用不同API Key

## 相关链接

- [认证路径配置完整指南](AUTH_PATHS_CONFIG.md)
- [HMAC认证中间件文档](../middleware/AUTH_MIDDLEWARE.md)
- [Kong CLI文档](KONG_CLI_REFACTORING.md)

## 总结

这次更新提供了更灵活的认证配置方式，使得 Echo Server 更适合在不同环境中部署，同时保持了完全的向后兼容性。

