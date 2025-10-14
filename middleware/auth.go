package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

const (
	// 默认API Key - 在生产环境中应该从环境变量读取
	DefaultHMACApiKey = "your-secret-api-key-here"
	// 默认时间戳有效期（秒）- 防止重放攻击
	DefaultHMACTimestampValid = 300 // 5分钟
)

// HMACAuthConfig HMAC认证中间件配置
type HMACAuthConfig struct {
	// API Key，如果为空则从环境变量 AUTH_API_KEY 读取
	ApiKey string
	// 时间戳有效期（秒），如果为0则使用默认值
	TimestampValid int64
	// 需要认证的路径列表，支持前缀匹配
	// 例如: []string{"/upload", "/api/private"}
	// 如果为空，则对所有路径进行认证
	Paths []string
	// 跳过认证的路径列表，优先级高于Paths
	// 例如: []string{"/upload/public"}
	SkipPaths []string
	// 自定义错误处理函数
	ErrorHandler func(c echo.Context, err error) error
}

// HMACAuth 返回HMAC认证中间件
func HMACAuth(config HMACAuthConfig) echo.MiddlewareFunc {
	// 设置默认值
	if config.ApiKey == "" {
		config.ApiKey = os.Getenv("AUTH_API_KEY")
		if config.ApiKey == "" {
			config.ApiKey = DefaultHMACApiKey
		}
	}

	if config.TimestampValid == 0 {
		timestampValidStr := os.Getenv("ECHO_AUTH_TIMESTAMP_VALID")
		if timestampValidStr != "" {
			if valid, err := strconv.ParseInt(timestampValidStr, 10, 64); err == nil {
				config.TimestampValid = valid
			} else {
				config.TimestampValid = DefaultHMACTimestampValid
			}
		} else {
			config.TimestampValid = DefaultHMACTimestampValid
		}
	}

	// 默认错误处理
	if config.ErrorHandler == nil {
		config.ErrorHandler = func(c echo.Context, err error) error {
			c.Logger().Errorf("HMAC auth failed: %v", err)
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "authentication failed",
			})
		}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path

			// 检查是否在跳过列表中
			if shouldSkip(path, config.SkipPaths) {
				return next(c)
			}

			// 检查是否需要认证
			if len(config.Paths) > 0 && !shouldAuth(path, config.Paths) {
				return next(c)
			}

			// 执行认证
			if err := validateHMACAuth(c, config.ApiKey, config.TimestampValid); err != nil {
				return config.ErrorHandler(c, err)
			}

			return next(c)
		}
	}
}

// shouldSkip 检查路径是否应该跳过认证
func shouldSkip(path string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if matchPath(path, skipPath) {
			return true
		}
	}
	return false
}

// shouldAuth 检查路径是否需要认证
func shouldAuth(path string, authPaths []string) bool {
	for _, authPath := range authPaths {
		if matchPath(path, authPath) {
			return true
		}
	}
	return false
}

// matchPath 路径匹配，支持前缀匹配和精确匹配
// 支持通配符 * 表示任意字符
func matchPath(path, pattern string) bool {
	// 精确匹配
	if path == pattern {
		return true
	}

	// 前缀匹配
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		return strings.HasPrefix(path, prefix)
	}

	// 后缀通配符
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(path, prefix)
	}

	return false
}

// validateHMACAuth 验证HMAC认证
// 认证方式：客户端发送 timestamp 和 signature (HMAC-SHA256)
// signature = HMAC-SHA256(apikey, timestamp)
func validateHMACAuth(c echo.Context, apiKey string, timestampValid int64) error {
	// 从header中获取认证信息
	timestamp := c.Request().Header.Get("X-Timestamp")
	signature := c.Request().Header.Get("X-Signature")

	if timestamp == "" || signature == "" {
		return fmt.Errorf("missing authentication headers")
	}

	// 验证时间戳
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid timestamp format")
	}

	now := time.Now().Unix()
	if now-ts > timestampValid || now-ts < -timestampValid {
		return fmt.Errorf("timestamp expired or invalid")
	}

	// 计算期望的签名
	expectedSignature := calculateHMACSignature(apiKey, timestamp)
	c.Logger().Debugf("expected signature: %s", expectedSignature)
	c.Logger().Debugf("signature: %s", signature)
	// 比较签名
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

// calculateHMACSignature 计算HMAC-SHA256签名
func calculateHMACSignature(key, message string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

// HMACAuthForPaths 为指定路径创建HMAC认证中间件的便捷函数
func HMACAuthForPaths(paths ...string) echo.MiddlewareFunc {
	return HMACAuth(HMACAuthConfig{
		Paths: paths,
	})
}

// HMACAuthWithConfig 使用自定义配置创建HMAC认证中间件的便捷函数
func HMACAuthWithConfig(apiKey string, timestampValid int64, paths []string) echo.MiddlewareFunc {
	return HMACAuth(HMACAuthConfig{
		ApiKey:         apiKey,
		TimestampValid: timestampValid,
		Paths:          paths,
	})
}
