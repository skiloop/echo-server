package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	// API配置 - 需要与服务器端的ApiKey一致
	ApiKey    = "your-secret-api-key-here"
	ServerURL = "http://localhost:9012/upload" // 或 https://localhost:9013/upload
)

// UploadResponse 服务器响应结构
type UploadResponse struct {
	Success  bool   `json:"success"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	Path     string `json:"path"`
	Error    string `json:"error,omitempty"`
}

// calculateHMAC 计算HMAC-SHA256签名
func calculateHMAC(key, message string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

// uploadFile 上传文件到服务器
func uploadFile(filePath string, uploadDir string) error {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	// 创建multipart表单
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加文件字段
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("创建表单文件字段失败: %w", err)
	}

	// 复制文件内容
	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("复制文件内容失败: %w", err)
	}

	// 关闭writer
	if err := writer.Close(); err != nil {
		return fmt.Errorf("关闭writer失败: %w", err)
	}

	// 生成时间戳
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	// 计算签名
	signature := calculateHMAC(ApiKey, timestamp)

	// 构建URL
	url := ServerURL
	if uploadDir != "" {
		url = fmt.Sprintf("%s?dir=%s", ServerURL, uploadDir)
	}

	// 创建请求
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Timestamp", timestamp)
	req.Header.Set("X-Signature", signature)

	// 发送请求
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	var uploadResp UploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	// 处理响应
	if resp.StatusCode == http.StatusOK && uploadResp.Success {
		fmt.Printf("✅ 上传成功!\n")
		fmt.Printf("   文件名: %s\n", uploadResp.Filename)
		fmt.Printf("   大小: %d bytes\n", uploadResp.Size)
		fmt.Printf("   相对路径: %s\n", uploadResp.Path)
		fmt.Printf("   (相对于根上传目录)\n")
		return nil
	} else {
		return fmt.Errorf("上传失败 (状态码: %d): %s", resp.StatusCode, uploadResp.Error)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("使用方法: go run upload_client_example.go <file_path> [upload_dir]")
		fmt.Println("示例: go run upload_client_example.go test.txt")
		fmt.Println("示例: go run upload_client_example.go test.txt ./custom_uploads")
		os.Exit(1)
	}

	filePath := os.Args[1]
	uploadDir := ""
	if len(os.Args) > 2 {
		uploadDir = os.Args[2]
	}

	if err := uploadFile(filePath, uploadDir); err != nil {
		fmt.Printf("❌ 错误: %v\n", err)
		os.Exit(1)
	}
}
