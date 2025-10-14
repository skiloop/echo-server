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

var (
	// 全局根上传目录，服务启动时初始化
	rootUploadDir string
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

// InitUploadDir 初始化全局上传根目录（应该在服务启动时调用）
func InitUploadDir() error {
	dir := os.Getenv("UPLOAD_DIR")
	if dir == "" {
		dir = DefaultUploadDir
	}

	// 标准化为绝对路径
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("failed to resolve upload directory: %w", err)
	}
	rootUploadDir = filepath.Clean(absDir)

	// 确保目录存在
	if err := os.MkdirAll(rootUploadDir, 0755); err != nil {
		return fmt.Errorf("failed to create upload directory: %w", err)
	}

	return nil
}

// GetRootUploadDir 获取全局根上传目录
func GetRootUploadDir() string {
	if rootUploadDir == "" {
		// 如果未初始化，使用默认值
		_ = InitUploadDir()
	}
	return rootUploadDir
}

// validateUploadPath 验证上传路径是否安全
// subDir: 用户指定的子目录（可能包含恶意路径）
// filename: 文件名（可能包含恶意路径）
// 返回: 安全的绝对路径，确保在根上传目录内
func validateUploadPath(subDir, filename string) (string, error) {
	// 1. 获取根上传目录
	rootDir := GetRootUploadDir()

	// 2. 清理子目录路径，防止路径遍历
	var targetDir string
	if subDir != "" {
		// 清理子目录路径
		cleanSubDir := filepath.Clean(subDir)
		// 移除开头的 / 或 ..
		cleanSubDir = strings.TrimPrefix(cleanSubDir, "/")
		cleanSubDir = strings.TrimPrefix(cleanSubDir, string(filepath.Separator))

		// 移除所有的 ..
		parts := strings.Split(cleanSubDir, string(filepath.Separator))
		cleanParts := make([]string, 0, len(parts))
		for _, part := range parts {
			if part != "" && part != "." && part != ".." {
				cleanParts = append(cleanParts, part)
			}
		}
		cleanSubDir = filepath.Join(cleanParts...)

		// 构造目标目录
		targetDir = filepath.Join(rootDir, cleanSubDir)
	} else {
		targetDir = rootDir
	}

	// 3. 标准化目标目录
	absTargetDir, err := filepath.Abs(targetDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve target directory: %w", err)
	}
	absTargetDir = filepath.Clean(absTargetDir)

	// 4. 验证目标目录在根目录内
	if !strings.HasPrefix(absTargetDir, rootDir+string(filepath.Separator)) &&
		absTargetDir != rootDir {
		return "", fmt.Errorf("invalid subdirectory: outside root upload directory")
	}

	// 5. 清理文件名，移除任何路径分隔符
	cleanFilename := filepath.Base(filepath.Clean(filename))
	if cleanFilename == "." || cleanFilename == ".." || cleanFilename == "" {
		return "", errors.New("invalid filename")
	}

	// 6. 构造最终路径
	finalPath := filepath.Join(absTargetDir, cleanFilename)
	absFinalPath, err := filepath.Abs(finalPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve final path: %w", err)
	}
	absFinalPath = filepath.Clean(absFinalPath)

	// 7. 最终验证：确保在根目录内
	if !strings.HasPrefix(absFinalPath, rootDir+string(filepath.Separator)) &&
		absFinalPath != rootDir {
		return "", fmt.Errorf("path traversal detected: final path outside root directory")
	}

	return absFinalPath, nil
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

	// 2. 获取用户指定的子目录（可选）
	subDir := c.QueryParam("dir")

	// 3. 验证路径安全性 - 确保在根上传目录内
	dstPath, err := validateUploadPath(subDir, file.Filename)
	if err != nil {
		c.Logger().Errorf("path validation failed: %v, subdir: %s, filename: %s", err, subDir, file.Filename)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid file path",
		})
	}

	// 4. 确保目标目录存在
	dstDir := filepath.Dir(dstPath)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		c.Logger().Errorf("create target dir error: %v", err)
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": "Permission denied",
		})
	}

	// 5. 打开上传的文件
	src, err := file.Open()
	if err != nil {
		c.Logger().Errorf("open file error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "failed to open uploaded file",
		})
	}
	defer src.Close()

	// 6. 创建目标文件
	dst, err := os.Create(dstPath)
	if err != nil {
		c.Logger().Errorf("create destination file error: %v", err)
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": "Permission denied",
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

	// 8. 计算相对路径（不暴露服务器实际路径）
	rootDir := GetRootUploadDir()
	relativePath, err := filepath.Rel(rootDir, dstPath)
	if err != nil {
		// 如果计算相对路径失败，只返回文件名
		relativePath = filepath.Base(dstPath)
	}

	// 9. 返回成功响应
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success":  true,
		"filename": filepath.Base(dstPath),
		"size":     size,
		"path":     relativePath, // 只返回相对路径
	})
}

// SetUploadRouters 设置上传路由
func SetUploadRouters(e *echo.Echo) {
	// 初始化根上传目录
	if err := InitUploadDir(); err != nil {
		e.Logger.Errorf("Failed to initialize upload directory: %v", err)
	} else {
		e.Logger.Infof("Upload root directory: %s", GetRootUploadDir())
	}

	e.POST("/upload", UploadFile)
}
