package routers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestValidateUploadPath 测试路径验证功能
func TestValidateUploadPath(t *testing.T) {
	// 设置测试用的根目录
	rootUploadDir = "/tmp/test_uploads"
	defer func() { rootUploadDir = "" }() // 测试完成后重置

	tests := []struct {
		name        string
		subDir      string
		filename    string
		expectError bool
		expectedIn  string // 期望路径包含的部分
		description string
	}{
		{
			name:        "正常文件名-无子目录",
			subDir:      "",
			filename:    "test.txt",
			expectError: false,
			expectedIn:  "/tmp/test_uploads/test.txt",
			description: "普通文件应该被允许",
		},
		{
			name:        "正常文件名-有子目录",
			subDir:      "user1",
			filename:    "test.txt",
			expectError: false,
			expectedIn:  "/tmp/test_uploads/user1/test.txt",
			description: "子目录应该被允许",
		},
		{
			name:        "多级子目录",
			subDir:      "user1/photos/2024",
			filename:    "image.jpg",
			expectError: false,
			expectedIn:  "/tmp/test_uploads/user1/photos/2024/image.jpg",
			description: "多级子目录应该被允许",
		},
		{
			name:        "子目录路径遍历攻击",
			subDir:      "../../../etc",
			filename:    "passwd",
			expectError: false,
			expectedIn:  "/tmp/test_uploads/etc/passwd", // .. 被清理掉
			description: "路径遍历应该被清理",
		},
		{
			name:        "文件名路径遍历",
			subDir:      "",
			filename:    "../../../etc/passwd",
			expectError: false,
			expectedIn:  "/tmp/test_uploads/passwd", // 只保留文件名
			description: "文件名中的路径遍历应该被清理",
		},
		{
			name:        "空文件名",
			subDir:      "",
			filename:    "",
			expectError: true,
			description: "空文件名应该被拒绝",
		},
		{
			name:        "特殊文件名 - 点",
			subDir:      "",
			filename:    ".",
			expectError: true,
			description: ". 应该被拒绝",
		},
		{
			name:        "特殊文件名 - 双点",
			subDir:      "",
			filename:    "..",
			expectError: true,
			description: ".. 应该被拒绝",
		},
		{
			name:        "绝对路径子目录",
			subDir:      "/etc/config",
			filename:    "test.txt",
			expectError: false,
			expectedIn:  "/tmp/test_uploads/etc/config/test.txt", // 开头的 / 被移除
			description: "绝对路径应该被转换为相对路径",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validateUploadPath(tt.subDir, tt.filename)

			if tt.expectError {
				if err == nil {
					t.Errorf("%s: 期望错误但成功了, result=%s", tt.description, result)
				}
			} else {
				if err != nil {
					t.Errorf("%s: 不期望错误但失败了, err=%v", tt.description, err)
				} else {
					// 验证结果路径
					if result != tt.expectedIn {
						t.Errorf("%s: 路径不匹配\n期望: %s\n实际: %s", tt.description, tt.expectedIn, result)
					}

					// 验证在根目录内
					if !strings.HasPrefix(result, rootUploadDir) {
						t.Errorf("%s: 结果路径 %s 不在根目录 %s 内", tt.description, result, rootUploadDir)
					}
				}
			}
		})
	}
}

// TestValidateUploadPathSecurity 测试路径遍历安全性
func TestValidateUploadPathSecurity(t *testing.T) {
	// 设置测试根目录
	rootUploadDir = "/tmp/test_uploads_security"
	defer func() { rootUploadDir = "" }()

	// 测试各种路径遍历攻击场景
	tests := []struct {
		subDir      string
		filename    string
		description string
	}{
		{
			subDir:      "",
			filename:    "../../../etc/passwd",
			description: "文件名路径遍历",
		},
		{
			subDir:      "../../../etc",
			filename:    "passwd",
			description: "子目录路径遍历",
		},
		{
			subDir:      "../../..",
			filename:    "etc/passwd",
			description: "子目录多级遍历",
		},
		{
			subDir:      "/etc/config",
			filename:    "test.txt",
			description: "子目录绝对路径",
		},
		{
			subDir:      "user/../../../etc",
			filename:    "passwd",
			description: "子目录混合遍历",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result, err := validateUploadPath(tt.subDir, tt.filename)

			if err != nil {
				t.Logf("✅ 恶意路径被拦截: subDir=%s, filename=%s, error=%v", tt.subDir, tt.filename, err)
				return
			}

			// 验证结果路径必须在根目录内
			if !strings.HasPrefix(result, rootUploadDir+string(filepath.Separator)) &&
				result != rootUploadDir {
				t.Errorf("❌ 安全漏洞！路径逃逸根目录\n  subDir=%s\n  filename=%s\n  result=%s\n  root=%s",
					tt.subDir, tt.filename, result, rootUploadDir)
			} else {
				t.Logf("✅ 路径安全: %s (subDir=%s, filename=%s)", result, tt.subDir, tt.filename)
			}
		})
	}
}

// TestInitUploadDir 测试上传目录初始化
func TestInitUploadDir(t *testing.T) {
	// 保存原始值
	origDir := rootUploadDir
	defer func() { rootUploadDir = origDir }()

	// 清空环境变量
	oldEnv := os.Getenv("UPLOAD_DIR")
	os.Unsetenv("UPLOAD_DIR")
	defer func() {
		if oldEnv != "" {
			os.Setenv("UPLOAD_DIR", oldEnv)
		}
	}()

	// 测试1: 使用默认值
	rootUploadDir = ""
	err := InitUploadDir()
	if err != nil {
		t.Fatalf("初始化失败: %v", err)
	}
	if !strings.Contains(rootUploadDir, "uploads") {
		t.Errorf("默认目录不正确: %s", rootUploadDir)
	}
	t.Logf("✅ 默认目录: %s", rootUploadDir)

	// 测试2: 使用环境变量
	os.Setenv("UPLOAD_DIR", "/tmp/custom_uploads")
	rootUploadDir = ""
	err = InitUploadDir()
	if err != nil {
		t.Fatalf("初始化失败: %v", err)
	}
	if !strings.Contains(rootUploadDir, "custom_uploads") {
		t.Errorf("环境变量目录不正确: %s", rootUploadDir)
	}
	t.Logf("✅ 环境变量目录: %s", rootUploadDir)
}
