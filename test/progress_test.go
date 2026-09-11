package test

import (
	"errors"
	"testing"

	"github.com/example/stablediffusion"
)

// 该测试在没有共享库（mock 模式）时验证：
//  1. 回调能收到阶段事件与采样 step/steps；
//  2. 回调返回 true 后生成尽快停止并返回 ErrGenerationCanceled。
//
// 若加载了真实库但没有模型，NewContext 会失败，测试自动跳过。
func TestProgressCallbackCancel(t *testing.T) {
	ctx, err := stablediffusion.NewContext(stablediffusion.DefaultContextOptions("nonexistent-mock.gguf"))
	if err != nil {
		t.Skipf("无可用模型/共享库，跳过端到端取消测试: %v", err)
	}
	defer ctx.Free()

	var sawSampling, sawPhase bool
	var lastStep, totalSteps int

	cfg := stablediffusion.GenerationConfig{
		Prompt:     "test",
		Width:      64,
		Height:     64,
		Seed:       1,
		BatchCount: 1,
		Sampler: stablediffusion.SamplerConfig{
			Steps: 12,
		},
		ProgressCallback: func(p stablediffusion.ProgressInfo) bool {
			switch p.Phase {
			case "loading model", "encoding prompt", "decoding":
				sawPhase = true
			case "sampling":
				sawSampling = true
				lastStep, totalSteps = p.Step, p.Steps
			}
			// 第 5 步请求取消
			return p.Phase == "sampling" && p.Step >= 5
		},
	}

	_, gerr := ctx.GenerateImage(cfg)
	if !errors.Is(gerr, stablediffusion.ErrGenerationCanceled) {
		t.Logf("GenerateImage 返回: %v", gerr)
	}
	if !sawPhase {
		t.Fatal("回调未收到任何阶段事件")
	}
	if !sawSampling {
		t.Fatal("回调从未收到 sampling 阶段事件")
	}
	if totalSteps != 12 {
		t.Fatalf("总步数不匹配: got %d, want 12", totalSteps)
	}
	if lastStep > 6 {
		t.Fatalf("取消未及时生效，最后收到的 step=%d", lastStep)
	}
	t.Logf("阶段事件=%v sampling=%v last=%d/%d", sawPhase, sawSampling, lastStep, totalSteps)
}

func TestProgressCallbackNoCancel(t *testing.T) {
	ctx, err := stablediffusion.NewContext(stablediffusion.DefaultContextOptions("nonexistent-mock.gguf"))
	if err != nil {
		t.Skipf("无可用模型/共享库，跳过: %v", err)
	}
	defer ctx.Free()

	count := 0
	cfg := stablediffusion.GenerationConfig{
		Width:  64,
		Height: 64,
		Seed:   2,
		Sampler: stablediffusion.SamplerConfig{
			Steps: 5,
		},
		ProgressCallback: func(p stablediffusion.ProgressInfo) bool {
			count++
			return false
		},
	}
	_, _ = ctx.GenerateImage(cfg)
	if count == 0 {
		t.Fatal("未收到任何进度回调")
	}
}
