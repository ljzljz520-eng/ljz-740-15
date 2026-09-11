package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"image"
	"image/color"
	"image/png"
	"bytes"
	"syscall"
	"time"

	"github.com/example/stablediffusion"
	"github.com/example/stablediffusion/bindings"
)

// 定义请求和响应结构
type GenerateRequest struct {
	Prompt            string  `json:"prompt"`
	NegativePrompt    string  `json:"negative_prompt"`
	Width             int     `json:"width"`
	Height            int     `json:"height"`
	Seed              int64   `json:"seed"`
	Steps             int     `json:"steps"`
	GuidanceScale     float32 `json:"guidance_scale"`
	BatchCount        int     `json:"batch_count"`
}

type GenerateResponse struct {
	Images []ImageInfo `json:"images"`
	Error  string      `json:"error,omitempty"`
}

type ImageInfo struct {
	Width   uint32 `json:"width"`
	Height  uint32 `json:"height"`
	Channel uint32 `json:"channel"`
	Data    string `json:"data"` // Base64 encoded
}

// 全局变量
var (
	sdCtx *stablediffusion.Context
	// 并发控制：最多同时处理 1 个生成请求 (SD 通常是单流的)
	sem = make(chan struct{}, 1)
)

// 初始化函数
func init() {
	// 从环境变量获取模型路径
	modelPath := os.Getenv("MODEL_PATH")
	if modelPath == "" {
		modelPath = "models/model.gguf"
	}

	// 创建上下文
	var err error
	options := stablediffusion.DefaultContextOptions(modelPath)
	sdCtx, err = stablediffusion.NewContext(options)
	if err != nil {
		log.Printf("Warning: Failed to create context: %v", err)
		log.Println("Server will start but image generation will fail without a valid model")
	}
}

// 健康检查端点
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status":  "ok",
		"version": stablediffusion.GetVersion(),
		"commit":  stablediffusion.GetCommit(),
	}
	json.NewEncoder(w).Encode(response)
}

// 系统信息端点
func systemInfoHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"system_info": stablediffusion.GetSystemInfo(),
		"version":     stablediffusion.GetVersion(),
		"commit":      stablediffusion.GetCommit(),
	}
	json.NewEncoder(w).Encode(response)
}

// 图像生成端点
func generateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 解析请求
	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// 检查上下文是否初始化
	if sdCtx == nil {
		response := GenerateResponse{
			Error: "Context not initialized: missing model file",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// 并发控制
	select {
	case sem <- struct{}{}:
		defer func() { <-sem }()
	default:
		http.Error(w, "Server is busy", http.StatusServiceUnavailable)
		return
	}

	// 设置默认值
	if req.Width == 0 {
		req.Width = 512
	}
	if req.Height == 0 {
		req.Height = 512
	}
	if req.Steps == 0 {
		req.Steps = 20
	}
	if req.GuidanceScale == 0 {
		req.GuidanceScale = 7.5
	}
	if req.BatchCount == 0 {
		req.BatchCount = 1
	}

	// 创建生成配置
	cfg := stablediffusion.GenerationConfig{
		Prompt:         req.Prompt,
		NegativePrompt: req.NegativePrompt,
		Width:          req.Width,
		Height:         req.Height,
		Seed:           req.Seed,
		BatchCount:     req.BatchCount,
		Sampler: stablediffusion.SamplerConfig{
			Scheduler:    bindings.KARRAS_SCHEDULER,
			Method:       bindings.EULER_A_SAMPLE_METHOD,
			Steps:        req.Steps,
			TxtCfg:       req.GuidanceScale,
			ImgCfg:       1.0,
			DistilledCfg: 0.0,
		},
	}

	// 生成图像，已经用 sem 保证了串行访问
	images, err := sdCtx.GenerateImage(cfg)

	if err != nil {
		response := GenerateResponse{
			Error: fmt.Sprintf("Failed to generate image: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// 转换为响应格式
	imageInfos := make([]ImageInfo, len(images))
	for i, img := range images {
		// 将原始 RGB 数据转换为 PNG
		pngData, err := encodeToPNG(img)
		if err != nil {
			log.Printf("Failed to encode image %d: %v", i, err)
			continue
		}
		imageInfos[i] = ImageInfo{
			Width:   img.Width,
			Height:  img.Height,
			Channel: img.Channel,
			Data:    base64.StdEncoding.EncodeToString(pngData),
		}
	}

	// 返回响应
	response := GenerateResponse{
		Images: imageInfos,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	// 获取端口
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 获取主机
	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0"
	}

	mux := http.NewServeMux()
	// 注册路由
	mux.HandleFunc("/health", healthCheckHandler)
	mux.HandleFunc("/system-info", systemInfoHandler)
	
	// 给生成接口加上超时机制
	generateTimeout := 5 * time.Minute
	if t := os.Getenv("GENERATE_TIMEOUT"); t != "" {
		if d, err := time.ParseDuration(t); err == nil {
			generateTimeout = d
		}
	}
	
	generateTimeoutHandler := http.TimeoutHandler(http.HandlerFunc(generateHandler), generateTimeout, `{"error":"Generation timeout"}`)
	mux.Handle("/generate", generateTimeoutHandler)

	// 启动服务器
	serverAddr := fmt.Sprintf("%s:%s", host, port)
	srv := &http.Server{
		Addr:    serverAddr,
		Handler: mux,
	}

	log.Printf("Starting server on %s", serverAddr)
	log.Printf("Health check: http://localhost:%s/health", port)
	log.Printf("System info: http://localhost:%s/system-info", port)
	log.Printf("Generate endpoint: POST http://localhost:%s/generate", port)

	// 在后台启动服务器
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 优雅停机处理
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// 停止接收新的请求并等待现有请求完成（最多等待1分钟）
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	// 释放模型资源
	if sdCtx != nil {
		log.Println("Freeing stable-diffusion context...")
		sdCtx.Free()
	}

	log.Println("Server exiting")
}

// 辅助函数：将原始 RGB 数据转换为 PNG
func encodeToPNG(img *stablediffusion.Image) ([]byte, error) {
	rect := image.Rect(0, 0, int(img.Width), int(img.Height))
	rgba := image.NewRGBA(rect)

	// 假设输入是 RGB (3 channels)
	for y := 0; y < int(img.Height); y++ {
		for x := 0; x < int(img.Width); x++ {
			pos := (y*int(img.Width) + x) * int(img.Channel)
			if int(img.Channel) == 3 {
				rgba.SetRGBA(x, y, color.RGBA{
					R: img.Data[pos],
					G: img.Data[pos+1],
					B: img.Data[pos+2],
					A: 255,
				})
			} else if int(img.Channel) == 4 {
				rgba.SetRGBA(x, y, color.RGBA{
					R: img.Data[pos],
					G: img.Data[pos+1],
					B: img.Data[pos+2],
					A: img.Data[pos+3],
				})
			}
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, rgba); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
