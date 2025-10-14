package routers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

const (
	// 默认上传目录
	DefaultUploadDir = "./uploads"
	// 默认最大文件大小（字节）- 10MB
	DefaultMaxFileSize = 10 << 20 // 10MB
)

// getMaxFileSize 从环境变量获取最大文件大小配置
func getMaxFileSize() int64 {
	maxFileSizeStr := os.Getenv("UPLOAD_MAX_SIZE")
	if maxFileSizeStr != "" {
		if size, err := strconv.ParseInt(maxFileSizeStr, 10, 64); err == nil {
			return size
		}
	}
	return DefaultMaxFileSize
}

// validateUploadPath 验证上传路径是否安全
// 确保目标路径在允许的上传目录内，防止路径遍历攻击
func validateUploadPath(uploadDir, filename string) (string, error) {
	// 1. 标准化上传目录路径
	absUploadDir, err := filepath.Abs(uploadDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve upload directory: %w", err)
	}
	absUploadDir = filepath.Clean(absUploadDir)

	// 2. 清理文件名，移除任何路径分隔符
	cleanFilename := filepath.Base(filepath.Clean(filename))
	if cleanFilename == "." || cleanFilename == ".." || cleanFilename == "" {
		return "", errors.New("invalid filename")
	}

	// 3. 构造目标路径并标准化
	dstPath := filepath.Join(absUploadDir, cleanFilename)
	absDstPath, err := filepath.Abs(dstPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve destination path: %w", err)
	}
	absDstPath = filepath.Clean(absDstPath)

	// 4. 检查目标路径是否在上传目录内
	// 使用 strings.HasPrefix 并确保是目录边界
	if !strings.HasPrefix(absDstPath, absUploadDir+string(filepath.Separator)) &&
		absDstPath != absUploadDir {
		return "", fmt.Errorf("path traversal attempt detected: destination outside upload directory")
	}

	return absDstPath, nil
}

// UploadFile 处理文件上传的handler
// 注意：认证由中间件处理，这里不需要再验证
func UploadFile(c echo.Context) error {
	// 获取配置
	maxFileSize := getMaxFileSize()

	// 1. 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.Logger().Errorf("get form file error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "no file uploaded",
		})
	}

	if file.Size > maxFileSize {
		c.Logger().Errorf("file too large: %d bytes", file.Size)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("file too large, max size is %d bytes", maxFileSize),
		})
	}

	// 2. 获取并验证上传目录
	uploadDir := getUploadDir(c)

	// 3. 验证路径安全性
	dstPath, err := validateUploadPath(uploadDir, file.Filename)
	if err != nil {
		c.Logger().Errorf("path validation failed: %v, original filename: %s", err, file.Filename)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid file path",
		})
	}

	// 4. 确保上传目录存在
	uploadDirAbs, _ := filepath.Abs(uploadDir)
	if err := os.MkdirAll(uploadDirAbs, 0755); err != nil {
		c.Logger().Errorf("create upload dir error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to create upload directory",
		})
	}

	// 5. 打开上传的文件
	src, err := file.Open()
	if err != nil {
		c.Logger().Errorf("open file error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to open uploaded file",
		})
	}
	defer src.Close()

	// 6. 创建目标文件
	dst, err := os.Create(dstPath)
	if err != nil {
		c.Logger().Errorf("create destination file error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to create destination file",
		})
	}
	defer dst.Close()

	// 7. 复制文件内容
	size, err := io.Copy(dst, src)
	if err != nil {
		c.Logger().Errorf("copy file error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to save file",
		})
	}

	c.Logger().Infof("file uploaded successfully: %s, size: %d", dstPath, size)

	// 8. 返回成功响应
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success":  true,
		"filename": filepath.Base(dstPath),
		"size":     size,
		"path":     dstPath,
	})
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
