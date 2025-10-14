# 命令行参数解析重构 - 使用 Kong

## 变更概述

将命令行参数解析从标准库 `flag` 迁移到 `github.com/alecthomas/kong`，提供更好的命令行体验。

## 为什么使用 Kong？

### Kong 的优势

✅ **更好的用户体验**
- 自动生成美观的帮助信息
- 支持环境变量绑定
- 更清晰的参数描述

✅ **更简洁的代码**
- 使用结构体标签定义参数
- 自动类型转换
- 减少样板代码

✅ **功能更强大**
- 支持子命令
- 支持参数验证
- 支持默认值
- 支持环境变量

### 与 flag 库对比

**之前 (flag):**
```go
var keyFile = flag.String("key", "", "tls key file path, empty then env TLS_KEY_FILE will apply")
var certFile = flag.String("cert", "", "tls cert file path, empty then TLS_CERT_FILE will apply")
var httpAddr = flag.String("http", "0.0.0.0:9012", "http bind addr")
var httpsAddr = flag.String("https", "0.0.0.0:9013", "https bind addr")

func main() {
    flag.Parse()
    
    if *keyFile == "" {
        *keyFile = os.Getenv("TLS_KEY_FILE")
    }
    // ... 手动处理环境变量
}
```

**现在 (kong):**
```go
type CLI struct {
    HTTP  string `help:"HTTP bind address" default:"0.0.0.0:9012" env:"HTTP_ADDR"`
    HTTPS string `help:"HTTPS bind address" default:"0.0.0.0:9013" env:"HTTPS_ADDR"`
    Cert  string `help:"TLS certificate file path" env:"TLS_CERT_FILE"`
    Key   string `help:"TLS key file path" env:"TLS_KEY_FILE"`
    Debug bool   `help:"Enable debug logging" default:"true"`
}

func main() {
    cli := CLI{}
    _ = kong.Parse(&cli,
        kong.Name("echo-server"),
        kong.Description("A versatile echo server with JA3 fingerprinting and file upload support"),
        kong.UsageOnError(),
    )
    
    // 直接使用 cli.Key, cli.Cert 等
}
```

## 参数说明

### 命令行参数

| 参数 | 说明 | 默认值 | 环境变量 |
|------|------|--------|----------|
| `--http` | HTTP 绑定地址 | `0.0.0.0:9012` | `HTTP_ADDR` |
| `--https` | HTTPS 绑定地址 | `0.0.0.0:9013` | `HTTPS_ADDR` |
| `--cert` | TLS 证书文件路径 | - | `TLS_CERT_FILE` |
| `--key` | TLS 密钥文件路径 | - | `TLS_KEY_FILE` |
| `--debug` | 启用调试日志 | `true` | - |
| `--auth-paths` | 需要 HMAC 认证的路径 | `/upload,/upload/*` | `AUTH_PATHS` |

### 使用示例

#### 1. 查看帮助信息

```bash
./echo-server --help
```

输出：
```
Usage: echo-server [flags]

A versatile echo server with JA3 fingerprinting and file upload support

Flags:
  -h, --help                    Show context-sensitive help.
      --http="0.0.0.0:9012"     HTTP bind address ($HTTP_ADDR)
      --https="0.0.0.0:9013"    HTTPS bind address ($HTTPS_ADDR)
      --cert=STRING             TLS certificate file path ($TLS_CERT_FILE)
      --key=STRING              TLS key file path ($TLS_KEY_FILE)
      --debug                   Enable debug logging
      --auth-paths=/upload,/upload/*,...
                                Paths requiring HMAC authentication (supports
                                wildcards) ($AUTH_PATHS)
```

#### 2. 仅启动 HTTP 服务器

```bash
./echo-server --http 0.0.0.0:8080
```

或使用环境变量：

```bash
HTTP_ADDR=0.0.0.0:8080 ./echo-server
```

#### 3. 启动 HTTPS 服务器

```bash
./echo-server \
  --cert /path/to/cert.pem \
  --key /path/to/key.pem \
  --https 0.0.0.0:9443
```

或使用环境变量：

```bash
export TLS_CERT_FILE=/path/to/cert.pem
export TLS_KEY_FILE=/path/to/key.pem
export HTTPS_ADDR=0.0.0.0:9443

./echo-server
```

#### 4. 同时启动 HTTP 和 HTTPS

```bash
./echo-server \
  --http 0.0.0.0:8080 \
  --https 0.0.0.0:8443 \
  --cert cert.pem \
  --key key.pem
```

#### 5. 禁用调试日志

```bash
./echo-server --debug=false
```

#### 6. 配置认证路径

```bash
# 自定义需要认证的路径
./echo-server --auth-paths="/upload/*,/api/private/*,/admin/*"

# 使用环境变量
export AUTH_PATHS="/upload/*,/api/*"
./echo-server

# 禁用所有认证
./echo-server --auth-paths=""
```

## 配置优先级

参数值的优先级（从高到低）：

```
命令行参数 > 环境变量 > 默认值
```

### 示例

```bash
# 1. 使用默认值
./echo-server
# HTTP: 0.0.0.0:9012, HTTPS: 0.0.0.0:9013

# 2. 环境变量覆盖默认值
HTTP_ADDR=127.0.0.1:8080 ./echo-server
# HTTP: 127.0.0.1:8080, HTTPS: 0.0.0.0:9013

# 3. 命令行参数优先级最高
HTTP_ADDR=127.0.0.1:8080 ./echo-server --http 0.0.0.0:9999
# HTTP: 0.0.0.0:9999, HTTPS: 0.0.0.0:9013
```

## 技术细节

### CLI 结构体定义

```go
type CLI struct {
    HTTP      string   `help:"HTTP bind address" default:"0.0.0.0:9012" env:"HTTP_ADDR"`
    HTTPS     string   `help:"HTTPS bind address" default:"0.0.0.0:9013" env:"HTTPS_ADDR"`
    Cert      string   `help:"TLS certificate file path" env:"TLS_CERT_FILE"`
    Key       string   `help:"TLS key file path" env:"TLS_KEY_FILE"`
    Debug     bool     `help:"Enable debug logging" default:"true"`
    AuthPaths []string `help:"Paths requiring HMAC authentication (supports wildcards)" default:"/upload,/upload/*" env:"AUTH_PATHS" sep:","`
}
```

### 标签说明

- `help`: 参数描述，显示在帮助信息中
- `default`: 默认值
- `env`: 关联的环境变量名
- `sep`: 数组分隔符（用于切分字符串为数组）
- 其他支持的标签：
  - `short`: 短选项，如 `-h`
  - `required`: 标记为必需参数
  - `enum`: 限制可选值
  - `placeholder`: 帮助信息中的占位符

### Kong 配置选项

```go
kong.Parse(&cli,
    kong.Name("echo-server"),           // 程序名称
    kong.Description("..."),             // 程序描述
    kong.UsageOnError(),                 // 参数错误时显示用法
)
```

其他有用的选项：
- `kong.ConfigureHelp()`: 自定义帮助格式
- `kong.Vars{}`: 设置变量
- `kong.Writers()`: 自定义输出流
- `kong.Exit()`: 自定义退出函数

## 向后兼容性

### 参数名称映射

| 旧参数 (flag) | 新参数 (kong) | 说明 |
|---------------|---------------|------|
| `-http` | `--http` | 功能相同 |
| `-https` | `--https` | 功能相同 |
| `-cert` | `--cert` | 功能相同 |
| `-key` | `--key` | 功能相同 |
| - | `--debug` | 新增 |

### 环境变量映射

| 旧环境变量 | 新环境变量 | 说明 |
|------------|------------|------|
| `TLS_CERT_FILE` | `TLS_CERT_FILE` | 保持不变 |
| `TLS_KEY_FILE` | `TLS_KEY_FILE` | 保持不变 |
| - | `HTTP_ADDR` | 新增 |
| - | `HTTPS_ADDR` | 新增 |
| `BIND_ADDR_TLS` | `HTTPS_ADDR` | 重命名建议 |

### 迁移指南

**如果你之前使用：**

```bash
./echo-server -http 0.0.0.0:8080 -cert cert.pem -key key.pem
```

**现在应该使用：**

```bash
./echo-server --http 0.0.0.0:8080 --cert cert.pem --key key.pem
```

或者简写（kong 支持）：

```bash
./echo-server --http=0.0.0.0:8080 --cert=cert.pem --key=key.pem
```

**注意：** kong 同时支持 `-` 和 `--` 前缀，但推荐使用 `--` 以保持一致性。

## 代码改动摘要

### 新增依赖

```go
import "github.com/alecthomas/kong"
```

### 删除代码

```go
// 删除全局变量
var keyFile = flag.String(...)
var certFile = flag.String(...)
// ...

// 删除手动环境变量处理
if "" == *keyFile {
    *keyFile = os.Getenv("TLS_KEY_FILE")
}
```

### 新增代码

```go
// 定义 CLI 结构体
type CLI struct {
    HTTP  string `help:"..." default:"..." env:"..."`
    // ...
}

// 解析参数
cli := CLI{}
_ = kong.Parse(&cli, ...)

// 直接使用
serve(e, cli.HTTP, cli.HTTPS, cli.Cert, cli.Key)
```

## 扩展性

### 添加新参数

只需在 CLI 结构体中添加新字段：

```go
type CLI struct {
    // 现有参数...
    
    // 新增参数
    LogLevel  string `help:"Log level" default:"info" enum:"debug,info,warn,error"`
    MaxConns  int    `help:"Maximum concurrent connections" default:"1000"`
    Timeout   int    `help:"Request timeout in seconds" default:"30"`
}
```

### 添加子命令

Kong 支持子命令结构：

```go
type CLI struct {
    Server ServerCmd `cmd:"" help:"Start server (default)"`
    Test   TestCmd   `cmd:"" help:"Run tests"`
    Config ConfigCmd `cmd:"" help:"Manage configuration"`
}

type ServerCmd struct {
    HTTP  string `help:"HTTP bind address"`
    // ...
}

type TestCmd struct {
    Pattern string `arg:"" help:"Test pattern"`
}

func main() {
    cli := CLI{}
    ctx := kong.Parse(&cli)
    
    switch ctx.Command() {
    case "server":
        // 启动服务器
    case "test":
        // 运行测试
    case "config":
        // 管理配置
    }
}
```

### 添加参数验证

```go
type CLI struct {
    Port int `help:"Server port" default:"8080" range:"1024:65535"`
}
```

## 测试

### 单元测试示例

```go
func TestCLIParsing(t *testing.T) {
    cli := CLI{}
    parser, err := kong.New(&cli)
    if err != nil {
        t.Fatal(err)
    }
    
    _, err = parser.Parse([]string{"--http", "127.0.0.1:8080"})
    if err != nil {
        t.Fatal(err)
    }
    
    assert.Equal(t, "127.0.0.1:8080", cli.HTTP)
}
```

### 集成测试

```bash
# 测试帮助信息
./echo-server --help

# 测试默认参数
./echo-server

# 测试自定义参数
./echo-server --http localhost:9999

# 测试环境变量
HTTP_ADDR=localhost:8888 ./echo-server

# 测试参数优先级
HTTP_ADDR=localhost:8888 ./echo-server --http localhost:9999
```

## 最佳实践

### 1. 使用有意义的默认值

```go
HTTP  string `default:"0.0.0.0:9012"` // ✅ 好
HTTP  string `default:""`              // ❌ 不好
```

### 2. 提供清晰的帮助文本

```go
Cert string `help:"TLS certificate file path"` // ✅ 好
Cert string `help:"cert"`                       // ❌ 不好
```

### 3. 合理使用环境变量

```go
// ✅ 好 - 敏感信息使用环境变量
Key string `env:"TLS_KEY_FILE"`

// ✅ 好 - 常用配置提供环境变量
HTTP string `env:"HTTP_ADDR"`
```

### 4. 使用枚举限制选项

```go
LogLevel string `enum:"debug,info,warn,error"`
```

### 5. 标记必需参数

```go
APIKey string `required:"" help:"API key for authentication"`
```

## 故障排查

### 问题：参数解析失败

**错误信息：**
```
kong: unknown flag --unknown
```

**解决方案：**
检查参数名是否正确，查看 `--help` 获取可用参数列表。

### 问题：环境变量不生效

**原因：**
环境变量名不匹配或未正确设置。

**解决方案：**
```bash
# 检查环境变量是否设置
echo $HTTP_ADDR

# 确保环境变量名与 CLI 结构体中的 env 标签一致
```

### 问题：默认值未生效

**原因：**
命令行参数或环境变量覆盖了默认值。

**解决方案：**
按优先级检查：命令行参数 > 环境变量 > 默认值

## 参考资料

- [Kong 官方文档](https://github.com/alecthomas/kong)
- [Kong 示例](https://github.com/alecthomas/kong/tree/master/_examples)
- [Go 命令行最佳实践](https://github.com/golang-standards/project-layout)

## 总结

使用 Kong 重构命令行参数解析带来了以下改进：

✅ **更好的用户体验** - 自动生成美观的帮助信息  
✅ **更简洁的代码** - 减少样板代码，提高可维护性  
✅ **更强大的功能** - 支持环境变量、参数验证等  
✅ **更好的可扩展性** - 易于添加新参数和子命令  
✅ **类型安全** - 编译时类型检查  

这次重构为项目的长期维护和功能扩展奠定了良好的基础。

