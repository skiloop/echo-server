package routers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

const (
	// 默认上传目录
	DefaultUploadDir = "./uploads"
	// 默认API Key - 在生产环境中应该从环境变量或配置文件读取
	DefaultApiKey = "your-secret-api-key-here"
	// 默认时间戳有效期（秒）- 防止重放攻击
	DefaultTimestampValidDuration = 300 // 5分钟
	// 默认最大文件大小（字节）- 10MB
	DefaultMaxFileSize = 10 << 20 // 10MB
)

// getConfig 从环境变量获取配置，如果不存在则使用默认值
func getConfig() (apiKey string, maxFileSize int64, timestampValid int64) {
	// 获取API Key
	apiKey = os.Getenv("UPLOAD_API_KEY")
	if apiKey == "" {
		apiKey = DefaultApiKey
	}

	// 获取最大文件大小
	maxFileSizeStr := os.Getenv("UPLOAD_MAX_SIZE")
	if maxFileSizeStr != "" {
		if size, err := strconv.ParseInt(maxFileSizeStr, 10, 64); err == nil {
			maxFileSize = size
		} else {
			maxFileSize = DefaultMaxFileSize
		}
	} else {
		maxFileSize = DefaultMaxFileSize
	}

	// 获取时间戳有效期
	timestampValidStr := os.Getenv("UPLOAD_TIMESTAMP_VALID")
	if timestampValidStr != "" {
		if valid, err := strconv.ParseInt(timestampValidStr, 10, 64); err == nil {
			timestampValid = valid
		} else {
			timestampValid = DefaultTimestampValidDuration
		}
	} else {
		timestampValid = DefaultTimestampValidDuration
	}

	return
}

// UploadFile 处理文件上传的handler
func UploadFile(c echo.Context) error {
	// 获取配置
	apiKey, maxFileSize, timestampValid := getConfig()

	// 1. 验证认证
	if err := validateAuth(c, apiKey, timestampValid); err != nil {
		c.Logger().Errorf("auth failed: %v", err)
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "authentication failed",
		})
	}

	// 2. 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.Logger().Errorf("get form file error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "no file uploaded",
		})
	}

	// 2.1 检查文件大小
	if file.Size > maxFileSize {
		c.Logger().Errorf("file too large: %d bytes", file.Size)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("file too large, max size is %d bytes", maxFileSize),
		})
	}

	// 2.2 清理文件名，防止路径遍历攻击
	filename := filepath.Base(filepath.Clean(file.Filename))
	if filename == "." || filename == ".." {
		c.Logger().Errorf("invalid filename: %s", file.Filename)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid filename",
		})
	}

	// 3. 打开上传的文件
	src, err := file.Open()
	if err != nil {
		c.Logger().Errorf("open file error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to open uploaded file",
		})
	}
	defer src.Close()

	// 4. 确保上传目录存在
	uploadDir := getUploadDir(c)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.Logger().Errorf("create upload dir error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to create upload directory",
		})
	}

	// 5. 创建目标文件（使用清理后的文件名）
	dstPath := filepath.Join(uploadDir, filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		c.Logger().Errorf("create destination file error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to create destination file",
		})
	}
	defer dst.Close()

	// 6. 复制文件内容
	size, err := io.Copy(dst, src)
	if err != nil {
		c.Logger().Errorf("copy file error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to save file",
		})
	}

	c.Logger().Infof("file uploaded successfully: %s, size: %d", dstPath, size)

	// 7. 返回成功响应
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success":  true,
		"filename": filename,
		"size":     size,
		"path":     dstPath,
	})
}

// validateAuth 验证API key认证
// 认证方式：客户端发送 timestamp 和 signature (HMAC-SHA256)
// signature = HMAC-SHA256(apikey, timestamp)
func validateAuth(c echo.Context, apiKey string, timestampValid int64) error {
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
	expectedSignature := calculateHMAC(apiKey, timestamp)

	// 比较签名
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

// calculateHMAC 计算HMAC-SHA256签名
func calculateHMAC(key, message string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

// getUploadDir 获取上传目录，优先级：query参数 > 环境变量 > 默认值
func getUploadDir(c echo.Context) string {
	// 优先使用query参数
	dir := c.QueryParam("dir")
	if dir == "" {
		// 其次使用环境变量
		dir = os.Getenv("UPLOAD_DIR")
	}
	if dir == "" {
		// 最后使用默认值
		dir = DefaultUploadDir
	}
	return dir
}

// SetUploadRouters 设置上传路由
func SetUploadRouters(e *echo.Echo) {
	e.POST("/upload", UploadFile)
}
