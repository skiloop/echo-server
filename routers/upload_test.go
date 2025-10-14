package routers

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestValidateUploadPath 测试路径验证功能
func TestValidateUploadPath(t *testing.T) {
	tests := []struct {
		name        string
		uploadDir   string
		filename    string
		expectError bool
		description string
	}{
		{
			name:        "正常文件名",
			uploadDir:   "/tmp/uploads",
			filename:    "test.txt",
			expectError: false,
			description: "普通文件应该被允许",
		},
		{
			name:        "路径遍历攻击 - 使用../",
			uploadDir:   "/tmp/uploads",
			filename:    "../../../etc/passwd",
			expectError: false, // filepath.Base 会清理掉 ../
			description: "filepath.Base 会自动清理路径遍历",
		},
		{
			name:        "空文件名",
			uploadDir:   "/tmp/uploads",
			filename:    "",
			expectError: true,
			description: "空文件名应该被拒绝",
		},
		{
			name:        "当前目录",
			uploadDir:   "/tmp/uploads",
			filename:    ".",
			expectError: true,
			description: ". 应该被拒绝",
		},
		{
			name:        "父目录",
			uploadDir:   "/tmp/uploads",
			filename:    "..",
			expectError: true,
			description: ".. 应该被拒绝",
		},
		{
			name:        "文件名包含路径分隔符",
			uploadDir:   "/tmp/uploads",
			filename:    "subdir/test.txt",
			expectError: false,
			description: "filepath.Base 会只取文件名部分",
		},
		{
			name:        "正常的子目录路径",
			uploadDir:   "/tmp/uploads",
			filename:    "test.txt",
			expectError: false,
			description: "正常路径应该允许",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validateUploadPath(tt.uploadDir, tt.filename)

			if tt.expectError {
				if err == nil {
					t.Errorf("%s: 期望错误但成功了, result=%s", tt.description, result)
				}
			} else {
				if err != nil {
					t.Errorf("%s: 不期望错误但失败了, err=%v", tt.description, err)
				} else {
					// 验证结果路径在上传目录内
					absUploadDir, _ := filepath.Abs(tt.uploadDir)
					if !strings.HasPrefix(result, absUploadDir) {
						t.Errorf("%s: 结果路径 %s 不在上传目录 %s 内", tt.description, result, absUploadDir)
					}
				}
			}
		})
	}
}

// TestValidateUploadPathSecurity 测试路径遍历安全性
func TestValidateUploadPathSecurity(t *testing.T) {
	uploadDir := "/tmp/uploads"
	absUploadDir, _ := filepath.Abs(uploadDir)

	// 测试各种路径遍历攻击
	maliciousFilenames := []string{
		"../../../etc/passwd",
		"..\\..\\..\\windows\\system32\\config\\sam",
		"....//....//....//etc/passwd",
		"./../../../etc/passwd",
		"test/../../../etc/passwd",
	}

	for _, filename := range maliciousFilenames {
		t.Run("Malicious: "+filename, func(t *testing.T) {
			result, err := validateUploadPath(uploadDir, filename)

			if err != nil {
				// 如果返回错误，那很好，说明被拦截了
				t.Logf("✅ 恶意路径被正确拦截: %s -> error: %v", filename, err)
				return
			}

			// 如果没有返回错误，检查结果路径是否安全
			// filepath.Base 应该已经清理掉路径遍历
			if !strings.HasPrefix(result, absUploadDir+string(filepath.Separator)) {
				t.Errorf("❌ 安全漏洞！路径 %s 被转换为 %s，不在上传目录 %s 内", filename, result, absUploadDir)
			} else {
				t.Logf("✅ 路径已被安全清理: %s -> %s", filename, result)
			}
		})
	}
}

// TestGetUploadDir 测试获取上传目录的优先级
func TestGetUploadDir(t *testing.T) {
	// 这里可以添加对 getUploadDir 的测试
	// 需要 echo.Context mock
	t.Skip("需要 echo.Context mock，暂时跳过")
}
