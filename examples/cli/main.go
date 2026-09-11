// 命令行示例：展示生成过程中的进度回调与取消。
//
// 在没有构建出 libstable-diffusion 共享库时（例如 CI/开发机上），
// 绑定层会自动进入 mock 模式：NewContext 返回 nil 错误，但 GenerateImage
// 内部会用模拟的进度事件驱动回调，因此仍能完整演示进度与取消流程。
//
// 用法：
//
//	# 真实模型（需要先构建共享库并设置模型路径）
//	go run ./examples/cli -m models/model.safetensors -p "a cat" -o outputs
//
//	# mock 模式（无模型也可验证回调）
//	go run ./examples/cli -mock -steps 30
//
//	# 只跑前几步后主动取消，验证取消路径
//	go run ./examples/cli -mock -steps 30 -cancel-after 8
package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"os/signal"
	"path/filepath"
	"sync/atomic"
	"syscall"
	"time"

	sd "github.com/example/stablediffusion"
	"github.com/example/stablediffusion/bindings"
)

func main() {
	modelPath := flag.String("m", "", "模型文件路径（GGUF/safetensors）")
	prompt := flag.String("p", "a lovely cat sitting on a windowsill", "正向提示词")
	negative := flag.String("n", "", "反向提示词")
	width := flag.Int("W", 512, "图像宽度")
	height := flag.Int("H", 512, "图像高度")
	steps := flag.Int("steps", 20, "采样步数")
	seed := flag.Int64("seed", 42, "随机种子")
	cfgScale := flag.Float64("cfg", 7.0, "CFG guidance 系数")
	outDir := flag.String("o", "outputs", "输出目录")
	mock := flag.Bool("mock", false, "仅演示进度回调流程（不创建真实上下文、不生成图像）")
	every := flag.Int("progress-every", 3, "采样阶段每隔多少步打印一次进度")
	cancelAfter := flag.Int("cancel-after", 0, ">0 时在该步后由回调主动返回取消标记")
	flag.Parse()

	// 取消标记：Ctrl+C 置位；-cancel-after 演示也会置位。
	var cancelRequested atomic.Bool
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		s := <-sigCh
		fmt.Printf("\n[cli] 收到信号 %v，请求取消生成……\n", s)
		cancelRequested.Store(true)
	}()

	progressCb := func(p sd.ProgressInfo) bool {
		// 阶段切换事件（Steps 为 0）始终打印；采样阶段按 -progress-every 节流。
		if p.Steps == 0 {
			fmt.Printf("[阶段] %s\n", p.Phase)
		} else if p.Step == p.Steps || p.Step%*every == 0 {
			line := fmt.Sprintf("[进度] %-22s %d/%d", p.Phase, p.Step, p.Steps)
			if p.Time > 0 && p.Step > 0 {
				line += fmt.Sprintf("  (%.2fs/it)", p.Time)
			}
			fmt.Println(line)
		}

		// 演示：跑到指定步数后由回调自身请求取消。
		if *cancelAfter > 0 && p.Step >= *cancelAfter && p.Phase == "sampling" {
			fmt.Printf("[cli] 演示取消：在第 %d 步请求停止\n", p.Step)
			return true
		}
		// 外部信号（Ctrl+C）触发取消。
		return cancelRequested.Load()
	}

	if *mock {
		runMockDemo(*steps, progressCb)
		return
	}

	if *modelPath == "" {
		fmt.Println("错误：请用 -m 指定模型路径，或加 -mock 查看回调演示")
		os.Exit(2)
	}

	fmt.Printf("系统信息: %s\n", sd.GetSystemInfo())
	fmt.Printf("创建上下文并加载模型: %s\n", *modelPath)

	// 模型加载阶段（NewContext 内部）的进度通过全局回调接收，
	// 底层 C 接口的进度回调是进程级单例。
	sd.SetProgressCallback(progressCb)

	opts := sd.DefaultContextOptions(*modelPath)
	opts.Wtype = bindings.SD_TYPE_F16
	ctx, err := sd.NewContext(opts)
	if err != nil {
		fmt.Printf("创建上下文失败: %v\n", err)
		os.Exit(1)
	}
	defer ctx.Free()
	// 后续生成阶段使用每个 config 自带的回调（语义相同，结束后自动清理）。
	sd.SetProgressCallback(nil)

	scheduler := ctx.GetDefaultScheduler(bindings.EULER_A_SAMPLE_METHOD)
	genCfg := sd.GenerationConfig{
		Prompt:         *prompt,
		NegativePrompt: *negative,
		Width:          *width,
		Height:         *height,
		Seed:           *seed,
		BatchCount:     1,
		Sampler: sd.SamplerConfig{
			Scheduler: scheduler,
			Method:    bindings.EULER_A_SAMPLE_METHOD,
			Steps:     *steps,
			TxtCfg:    float32(*cfgScale),
		},
		ProgressCallback: progressCb,
	}

	t0 := time.Now()
	images, gerr := ctx.GenerateImage(genCfg)
	if gerr != nil {
		if errors.Is(gerr, sd.ErrGenerationCanceled) {
			fmt.Printf("生成已取消（耗时 %s），退出。\n", time.Since(t0).Round(time.Millisecond))
			os.Exit(130)
		}
		fmt.Printf("生成失败: %v\n", gerr)
		os.Exit(1)
	}
	fmt.Printf("生成完成，耗时 %s，共 %d 张图像\n", time.Since(t0).Round(time.Millisecond), len(images))

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fmt.Printf("创建输出目录失败: %v\n", err)
		os.Exit(1)
	}
	for i, img := range images {
		path := filepath.Join(*outDir, fmt.Sprintf("output_%d.png", i))
		if err := savePNG(path, img); err != nil {
			fmt.Printf("保存 %s 失败: %v\n", path, err)
			continue
		}
		fmt.Printf("已保存: %s (%dx%d)\n", path, img.Width, img.Height)
	}
}

// runMockDemo 通过绑定层的 mock 流程演示进度回调与取消。
func runMockDemo(steps int, cb sd.ProgressCallback) {
	fmt.Println("== mock 模式：模拟生成流程 ==")
	t0 := time.Now()

	// 安装全局回调；mock 实现会在“生成”时依次触发
	// loading model / encoding prompt / sampling 1..N / decoding 事件。
	sd.SetProgressCallback(cb)
	defer sd.SetProgressCallback(nil)

	cfg := sd.GenerationConfig{
		Prompt:     "mock prompt",
		Width:      64,
		Height:     64,
		Seed:       1,
		BatchCount: 1,
		Sampler: sd.SamplerConfig{
			Method: bindings.EULER_A_SAMPLE_METHOD,
			Steps:  steps,
			TxtCfg: 1,
		},
		ProgressCallback: cb,
	}

	// mock 模式下没有真实上下文，直接调用绑定层的 mock 生成入口：
	// 高层 API 需要 Context，所以这里用 bindings 的 mock 路径模拟。
	params := &bindings.SdImgGenParams{}
	bindings.SdImgGenParamsInit(params)
	params.Width = cfg.Width
	params.Height = cfg.Height
	params.Seed = cfg.Seed
	params.BatchCount = cfg.BatchCount
	params.SampleParams.SampleMethod = cfg.Sampler.Method
	params.SampleParams.SampleSteps = cfg.Sampler.Steps
	params.SampleParams.Guidance.TxtCfg = cfg.Sampler.TxtCfg
	result := bindings.GenerateImage(nil, params)
	if result != nil {
		fmt.Println("mock 不应产生图像（此分支仅为完整性保留）")
	}
	fmt.Printf("== mock 流程结束，耗时 %s ==\n", time.Since(t0).Round(time.Millisecond))
}

func savePNG(path string, img *sd.Image) error {
	if img.Channel != 3 {
		return fmt.Errorf("暂只支持 3 通道图像，实际 %d 通道", img.Channel)
	}
	w, h := int(img.Width), int(img.Height)
	rgba := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			src := (y*w + x) * 3
			dst := (y*w + x) * 4
			rgba.Pix[dst+0] = img.Data[src+0]
			rgba.Pix[dst+1] = img.Data[src+1]
			rgba.Pix[dst+2] = img.Data[src+2]
			rgba.Pix[dst+3] = 255
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, rgba)
}
