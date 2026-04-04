package routers

import (
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	watermark "github.com/yyyoichi/watermark_zero"
	"github.com/yyyoichi/watermark_zero/mark"
)

type WatermarkTask struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"` // pending, processing, completed, failed
	Progress  int       `json:"progress"`
	Result    string    `json:"result,omitempty"`    // output path for embed, watermark text for detect
	Watermark string    `json:"watermark,omitempty"` // watermark text for embed
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type watermarkManager struct {
	mu    sync.RWMutex
	tasks map[string]*WatermarkTask
}

var wm = &watermarkManager{
	tasks: make(map[string]*WatermarkTask),
}

func generateTaskID() string {
	return uuid.New().String()[:8]
}

func createTask(watermark string) string {
	taskID := generateTaskID()
	wm.mu.Lock()
	wm.tasks[taskID] = &WatermarkTask{
		ID:        taskID,
		Status:    "pending",
		Progress:  0,
		Watermark: watermark,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	wm.mu.Unlock()
	return taskID
}

func getTask(taskID string) *WatermarkTask {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return wm.tasks[taskID]
}

func updateTask(taskID string, status string, progress int, result, errMsg string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	if task, ok := wm.tasks[taskID]; ok {
		task.Status = status
		task.Progress = progress
		task.UpdatedAt = time.Now()
		if result != "" {
			task.Result = result
		}
		if errMsg != "" {
			task.Error = errMsg
		}
	}
}

func isLikelyWatermark(s string) bool {
	if len(s) == 0 {
		return false
	}
	printable := 0
	for _, r := range s {
		if r >= 32 && r < 127 {
			printable++
		}
	}
	return float64(printable)/float64(len(s)) > 0.7
}

func EmbedWatermarkAsync(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "no file uploaded",
		})
	}

	watermarkText := c.FormValue("watermark")
	if watermarkText == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "watermark text is required",
		})
	}

	taskID := createTask(watermarkText)

	go processWatermarkEmbed(taskID, file)

	return c.JSON(http.StatusAccepted, map[string]interface{}{
		"task_id":  taskID,
		"status":   "pending",
		"message":  "Watermark embedding task created",
		"check_at": "/watermark/status/" + taskID,
	})
}

func processWatermarkEmbed(taskID string, file *multipart.FileHeader) {
	updateTask(taskID, "processing", 10, "", "")

	src, err := file.Open()
	if err != nil {
		updateTask(taskID, "failed", 0, "", "failed to open uploaded file")
		return
	}
	defer src.Close()

	img, format, err := image.Decode(src)
	if err != nil {
		updateTask(taskID, "failed", 0, "", fmt.Sprintf("failed to decode image: %v", err))
		return
	}
	updateTask(taskID, "processing", 30, "", "")

	w, err := watermark.New(
		watermark.WithBlockShape(8, 6),
		watermark.WithD1D2(36, 20),
	)
	if err != nil {
		updateTask(taskID, "failed", 0, "", fmt.Sprintf("failed to create watermark instance: %v", err))
		return
	}

	m := mark.NewString(getTask(taskID).Watermark)
	updateTask(taskID, "processing", 50, "", "")

	markedImg, err := w.Embed(context.Background(), img, m)
	if err != nil {
		updateTask(taskID, "failed", 0, "", fmt.Sprintf("failed to embed watermark: %v", err))
		return
	}
	updateTask(taskID, "processing", 80, "", "")

	outputDir := filepath.Join(GetRootUploadDir(), "watermarked")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		updateTask(taskID, "failed", 0, "", fmt.Sprintf("failed to create output directory: %v", err))
		return
	}

	outputFilename := fmt.Sprintf("%s_%s.%s", taskID, sanitizeFilename(file.Filename), format)
	outputPath := filepath.Join(outputDir, outputFilename)

	f, err := os.Create(outputPath)
	if err != nil {
		updateTask(taskID, "failed", 0, "", fmt.Sprintf("failed to create output file: %v", err))
		return
	}
	defer f.Close()

	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		err = jpeg.Encode(f, markedImg, &jpeg.Options{Quality: 95})
	default:
		err = png.Encode(f, markedImg)
	}

	if err != nil {
		updateTask(taskID, "failed", 0, "", fmt.Sprintf("failed to save image: %v", err))
		return
	}

	updateTask(taskID, "completed", 100, outputPath, "")
}

func GetWatermarkStatus(c echo.Context) error {
	taskID := c.Param("taskId")
	if taskID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "task ID is required",
		})
	}

	task := getTask(taskID)
	if task == nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "task not found",
		})
	}

	return c.JSON(http.StatusOK, task)
}

func DetectWatermark(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "no file uploaded",
		})
	}

	sizeStr := c.FormValue("size")
	size := 256
	if sizeStr != "" {
		fmt.Sscanf(sizeStr, "%d", &size)
	}

	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "failed to open uploaded file",
		})
	}
	defer src.Close()

	img, _, err := image.Decode(src)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("failed to decode image: %v", err),
		})
	}

	w, err := watermark.New(
		watermark.WithBlockShape(8, 6),
		watermark.WithD1D2(36, 20),
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("failed to create watermark instance: %v", err),
		})
	}

	exM := mark.NewExtract(size)
	decoded, err := w.Extract(context.Background(), img, exM)
	if err != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"detected": false,
			"message":  "no watermark detected",
		})
	}

	text := decoded.DecodeToString()
	text = strings.TrimRight(text, "\x00")
	text = strings.TrimSpace(text)

	if text == "" || !isLikelyWatermark(text) {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"detected": false,
			"message":  "no watermark detected",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"detected":  true,
		"watermark": text,
	})
}

func sanitizeFilename(name string) string {
	ext := filepath.Ext(name)
	nameWithoutExt := strings.TrimSuffix(name, ext)
	nameWithoutExt = strings.ReplaceAll(nameWithoutExt, "/", "_")
	nameWithoutExt = strings.ReplaceAll(nameWithoutExt, "\\", "_")
	nameWithoutExt = strings.ReplaceAll(nameWithoutExt, "..", "_")
	return nameWithoutExt + ext
}

func DownloadWatermarked(c echo.Context) error {
	taskID := c.Param("taskId")
	if taskID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "task ID is required",
		})
	}

	task := getTask(taskID)
	if task == nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "task not found",
		})
	}

	if task.Status != "completed" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "task not completed yet",
		})
	}

	if task.Result == "" {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "file not found",
		})
	}

	return c.File(task.Result)
}

func SetWatermarkRouters(e *echo.Echo) {
	e.POST("/watermark/embed", EmbedWatermarkAsync)
	e.GET("/watermark/status/:taskId", GetWatermarkStatus)
	e.GET("/watermark/download/:taskId", DownloadWatermarked)
	e.POST("/watermark/detect", DetectWatermark)
}
