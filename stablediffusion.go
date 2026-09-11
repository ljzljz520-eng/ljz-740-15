package stablediffusion

import (
	"fmt"
	"unsafe"

	"github.com/example/stablediffusion/bindings"
)

// Context 表示stable-diffusion的上下文
type Context struct {
	ctx *bindings.SdCtx
}

// Image 表示生成的图像
type Image struct {
	Width   uint32
	Height  uint32
	Channel uint32
	Data    []byte
}

// Lora 表示LoRA模型
type Lora struct {
	Path        string
	Multiplier  float32
	IsHighNoise bool
}

// SamplerConfig 表示采样器配置
type SamplerConfig struct {
	Scheduler    bindings.Scheduler
	Method       bindings.SampleMethod
	Steps        int
	Eta          float32
	TxtCfg       float32
	ImgCfg       float32
	DistilledCfg float32
}

// TilingParams 表示平铺渲染参数
type TilingParams struct {
	Enabled       bool
	TileSizeX     int
	TileSizeY     int
	TargetOverlap float32
	RelSizeX      float32
	RelSizeY      float32
}

// CacheParams 表示缓存配置参数
type CacheParams struct {
	Mode                     bindings.SdCacheMode
	ReuseThreshold           float32
	StartPercent             float32
	EndPercent               float32
	ErrorDecayRate           float32
	UseRelativeThreshold     bool
	ResetErrorOnCompute      bool
	FnComputeBlocks          int
	BnComputeBlocks          int
	ResidualDiffThreshold    float32
	MaxWarmupSteps           int
	MaxCachedSteps           int
	MaxContinuousCachedSteps int
	TaylorseerNDerivatives   int
	TaylorseerSkipInterval   int
	ScmMask                  string
	ScmPolicyDynamic         bool
}

// GenerationConfig 表示图像生成配置
type GenerationConfig struct {
	Prompt             string
	NegativePrompt     string
	Width              int
	Height             int
	Seed               int64
	Strength           float32
	BatchCount         int
	ClipSkip           int
	Loras              []Lora
	ControlImage       *Image
	ControlStrength    float32
	InitImage          *Image
	MaskImage          *Image
	RefImages          []Image
	AutoResizeRefImage bool
	IncreaseRefIndex   bool
	Sampler            SamplerConfig
	VaeTilingParams    TilingParams
	Cache              CacheParams
}

// Upscaler 表示超分辨率器
type Upscaler struct {
	ctx *bindings.UpscalerCtx
}

// ContextOptions 定义了stable-diffusion上下文的完整配置参数
type ContextOptions struct {
	ModelPath                   string
	ClipLPath                   string
	ClipGPath                   string
	ClipVisionPath              string
	T5xxlPath                   string
	LlmPath                     string
	LlmVisionPath               string
	DiffusionModelPath          string
	HighNoiseDiffusionModelPath string
	VaePath                     string
	TaesdPath                   string
	ControlNetPath              string
	PhotoMakerPath              string
	TensorTypeRules             string
	VaeDecodeOnly               bool
	FreeParamsImmediately       bool
	NThreads                    int
	Wtype                       bindings.SdType
	RngType                     bindings.RngType
	SamplerRngType              bindings.RngType
	Prediction                  bindings.Prediction
	LoraApplyMode               bindings.LoraApplyMode
	OffloadParamsToCpu          bool
	EnableMmap                  bool
	KeepClipOnCpu               bool
	KeepControlNetOnCpu         bool
	KeepVaeOnCpu                bool
	FlashAttn                   bool
	DiffusionFlashAttn          bool
	TaePreviewOnly              bool
	DiffusionConvDirect         bool
	VaeConvDirect               bool
	CircularX                   bool
	CircularY                   bool
	ForceSdxlVaeConvScale       bool
	ChromaUseDitMask            bool
	ChromaUseT5Mask             bool
	ChromaT5MaskPad             int
	QwenImageZeroCondT          bool
	FlowShift                   float32
	TilingParams                TilingParams
	CacheParams                 CacheParams
}

// DefaultContextOptions 返回具有默认参数的上下文选项
func DefaultContextOptions(modelPath string) ContextOptions {
	return ContextOptions{
		ModelPath: modelPath,
		NThreads:  -1, // 自动
		Wtype:     bindings.SD_TYPE_F16,
		EnableMmap: true,
	}
}

// NewContext 创建一个新的stable-diffusion上下文
func NewContext(options ContextOptions) (*Context, error) {
	// 初始化上下文参数
	params := &bindings.SdCtxParams{}
	bindings.SdCtxParamsInit(params)

	if options.ModelPath != "" {
		params.ModelPath = bindings.CString(options.ModelPath)
	}
	if options.ClipLPath != "" {
		params.ClipLPath = bindings.CString(options.ClipLPath)
	}
	if options.ClipGPath != "" {
		params.ClipGPath = bindings.CString(options.ClipGPath)
	}
	if options.ClipVisionPath != "" {
		params.ClipVisionPath = bindings.CString(options.ClipVisionPath)
	}
	if options.T5xxlPath != "" {
		params.T5xxlPath = bindings.CString(options.T5xxlPath)
	}
	if options.LlmPath != "" {
		params.LlmPath = bindings.CString(options.LlmPath)
	}
	if options.LlmVisionPath != "" {
		params.LlmVisionPath = bindings.CString(options.LlmVisionPath)
	}
	if options.DiffusionModelPath != "" {
		params.DiffusionModelPath = bindings.CString(options.DiffusionModelPath)
	}
	if options.HighNoiseDiffusionModelPath != "" {
		params.HighNoiseDiffusionModelPath = bindings.CString(options.HighNoiseDiffusionModelPath)
	}
	if options.VaePath != "" {
		params.VaePath = bindings.CString(options.VaePath)
	}
	if options.TaesdPath != "" {
		params.TaesdPath = bindings.CString(options.TaesdPath)
	}
	if options.ControlNetPath != "" {
		params.ControlNetPath = bindings.CString(options.ControlNetPath)
	}
	if options.PhotoMakerPath != "" {
		params.PhotoMakerPath = bindings.CString(options.PhotoMakerPath)
	}
	if options.TensorTypeRules != "" {
		params.TensorTypeRules = bindings.CString(options.TensorTypeRules)
	}

	params.VaeDecodeOnly = options.VaeDecodeOnly
	params.FreeParamsImmediately = options.FreeParamsImmediately
	params.NThreads = options.NThreads
	params.Wtype = options.Wtype
	params.RngType = options.RngType
	params.SamplerRngType = options.SamplerRngType
	params.Prediction = options.Prediction
	params.LoraApplyMode = options.LoraApplyMode
	params.OffloadParamsToCpu = options.OffloadParamsToCpu
	params.EnableMmap = options.EnableMmap
	params.KeepClipOnCpu = options.KeepClipOnCpu
	params.KeepControlNetOnCpu = options.KeepControlNetOnCpu
	params.KeepVaeOnCpu = options.KeepVaeOnCpu
	params.FlashAttn = options.FlashAttn
	params.DiffusionFlashAttn = options.DiffusionFlashAttn
	params.TaePreviewOnly = options.TaePreviewOnly
	params.DiffusionConvDirect = options.DiffusionConvDirect
	params.VaeConvDirect = options.VaeConvDirect
	params.CircularX = options.CircularX
	params.CircularY = options.CircularY
	params.ForceSdxlVaeConvScale = options.ForceSdxlVaeConvScale
	params.ChromaUseDitMask = options.ChromaUseDitMask
	params.ChromaUseT5Mask = options.ChromaUseT5Mask
	params.ChromaT5MaskPad = options.ChromaT5MaskPad
	params.QwenImageZeroCondT = options.QwenImageZeroCondT
	params.FlowShift = options.FlowShift

	// 设置 Tiling 和 Cache
	params.CircularX = options.CircularX
	params.CircularY = options.CircularY

	// 注意：bindings 层的某些复杂结构体初始化可能需要显式处理

	// 创建上下文
	ctx := bindings.CreateSdCtx(params)
	if ctx == nil {
		return nil, fmt.Errorf("failed to create context")
	}

	return &Context{ctx: ctx}, nil
}

// Free 释放上下文
func (c *Context) Free() {
	if c.ctx != nil {
		bindings.FreeSdCtx(c.ctx)
		c.ctx = nil
	}
}

// NewUpscaler 创建一个新的超分辨率器
func NewUpscaler(modelPath string) (*Upscaler, error) {
	ctx := bindings.CreateUpscalerCtx(modelPath, false, true, -1, 0)
	if ctx == nil {
		return nil, fmt.Errorf("failed to create upscaler")
	}

	return &Upscaler{ctx: ctx}, nil
}

// Free 释放超分辨率器
func (u *Upscaler) Free() {
	if u.ctx != nil {
		bindings.FreeUpscalerCtx(u.ctx)
		u.ctx = nil
	}
}

// Upscale 执行超分辨率
func (u *Upscaler) Upscale(img *Image, factor uint32) (*Image, error) {
	// 转换为绑定的图像类型
	input := bindings.SdImage{
		Width:   img.Width,
		Height:  img.Height,
		Channel: img.Channel,
		Data:    (*uint8)(unsafe.Pointer(&img.Data[0])),
	}

	// 执行超分辨率
	result := bindings.Upscale(u.ctx, input, factor)

	// 转换回Go图像类型
	data := make([]byte, int(result.Width)*int(result.Height)*int(result.Channel))
	copy(data, unsafe.Slice(result.Data, len(data)))

	return &Image{
		Width:   result.Width,
		Height:  result.Height,
		Channel: result.Channel,
		Data:    data,
	}, nil
}

// GetUpscaleFactor 获取超分辨率因子
func (u *Upscaler) GetUpscaleFactor() int {
	return bindings.GetUpscaleFactor(u.ctx)
}

// GenerateImage 生成图像
func (c *Context) GenerateImage(cfg GenerationConfig) ([]*Image, error) {
	// 初始化图像生成参数
	params := &bindings.SdImgGenParams{}
	bindings.SdImgGenParamsInit(params)

	// 设置基本参数
	params.Prompt = bindings.CString(cfg.Prompt)
	params.NegativePrompt = bindings.CString(cfg.NegativePrompt)
	params.Width = cfg.Width
	params.Height = cfg.Height
	params.Seed = cfg.Seed
	params.Strength = cfg.Strength
	params.BatchCount = cfg.BatchCount
	params.ClipSkip = cfg.ClipSkip

	// 设置采样器参数
	params.SampleParams.Scheduler = cfg.Sampler.Scheduler
	params.SampleParams.SampleMethod = cfg.Sampler.Method
	params.SampleParams.SampleSteps = cfg.Sampler.Steps
	params.SampleParams.Eta = cfg.Sampler.Eta
	params.SampleParams.Guidance.TxtCfg = cfg.Sampler.TxtCfg
	params.SampleParams.Guidance.ImgCfg = cfg.Sampler.ImgCfg
	params.SampleParams.Guidance.DistilledGuidance = cfg.Sampler.DistilledCfg

	// 设置LoRA
	if len(cfg.Loras) > 0 {
		loras := make([]bindings.SdLora, len(cfg.Loras))
		for i, lora := range cfg.Loras {
			loras[i] = bindings.SdLora{
				Path:        bindings.CString(lora.Path),
				Multiplier:  lora.Multiplier,
				IsHighNoise: lora.IsHighNoise,
			}
		}
		params.Loras = &loras[0]
		params.LoraCount = uint32(len(loras))
	}

	// 设置控制图像
	if cfg.ControlImage != nil {
		params.ControlImage = bindings.SdImage{
			Width:   cfg.ControlImage.Width,
			Height:  cfg.ControlImage.Height,
			Channel: cfg.ControlImage.Channel,
			Data:    (*uint8)(unsafe.Pointer(&cfg.ControlImage.Data[0])),
		}
		params.ControlStrength = cfg.ControlStrength
	}

	// 设置初始图像
	if cfg.InitImage != nil {
		params.InitImage = bindings.SdImage{
			Width:   cfg.InitImage.Width,
			Height:  cfg.InitImage.Height,
			Channel: cfg.InitImage.Channel,
			Data:    (*uint8)(unsafe.Pointer(&cfg.InitImage.Data[0])),
		}
	}

	// 设置掩码图像
	if cfg.MaskImage != nil {
		params.MaskImage = bindings.SdImage{
			Width:   cfg.MaskImage.Width,
			Height:  cfg.MaskImage.Height,
			Channel: cfg.MaskImage.Channel,
			Data:    (*uint8)(unsafe.Pointer(&cfg.MaskImage.Data[0])),
		}
	}

	// 设置参考图像
	if len(cfg.RefImages) > 0 {
		refImages := make([]bindings.SdImage, len(cfg.RefImages))
		for i, img := range cfg.RefImages {
			refImages[i] = bindings.SdImage{
				Width:   img.Width,
				Height:  img.Height,
				Channel: img.Channel,
				Data:    (*uint8)(unsafe.Pointer(&img.Data[0])),
			}
		}
		params.RefImages = &refImages[0]
		params.RefImagesCount = len(refImages)
		params.AutoResizeRefImage = cfg.AutoResizeRefImage
		params.IncreaseRefIndex = cfg.IncreaseRefIndex
	}

	// 设置 Tiling 和 Cache 参数
	params.VaeTilingParams = bindings.SdTilingParams{
		Enabled:       cfg.VaeTilingParams.Enabled,
		TileSizeX:     cfg.VaeTilingParams.TileSizeX,
		TileSizeY:     cfg.VaeTilingParams.TileSizeY,
		TargetOverlap: cfg.VaeTilingParams.TargetOverlap,
		RelSizeX:      cfg.VaeTilingParams.RelSizeX,
		RelSizeY:      cfg.VaeTilingParams.RelSizeY,
	}

	params.Cache = bindings.SdCacheParams{
		Mode:                     cfg.Cache.Mode,
		ReuseThreshold:           cfg.Cache.ReuseThreshold,
		StartPercent:             cfg.Cache.StartPercent,
		EndPercent:               cfg.Cache.EndPercent,
		ErrorDecayRate:           cfg.Cache.ErrorDecayRate,
		UseRelativeThreshold:     cfg.Cache.UseRelativeThreshold,
		ResetErrorOnCompute:      cfg.Cache.ResetErrorOnCompute,
		FnComputeBlocks:          cfg.Cache.FnComputeBlocks,
		BnComputeBlocks:          cfg.Cache.BnComputeBlocks,
		ResidualDiffThreshold:    cfg.Cache.ResidualDiffThreshold,
		MaxWarmupSteps:           cfg.Cache.MaxWarmupSteps,
		MaxCachedSteps:           cfg.Cache.MaxCachedSteps,
		MaxContinuousCachedSteps: cfg.Cache.MaxContinuousCachedSteps,
		TaylorseerNDerivatives:   cfg.Cache.TaylorseerNDerivatives,
		TaylorseerSkipInterval:   cfg.Cache.TaylorseerSkipInterval,
		ScmPolicyDynamic:         cfg.Cache.ScmPolicyDynamic,
	}
	if cfg.Cache.ScmMask != "" {
		params.Cache.ScmMask = bindings.CString(cfg.Cache.ScmMask)
	}

	// 生成图像
	result := bindings.GenerateImage(c.ctx, params)
	if result == nil {
		return nil, fmt.Errorf("failed to generate image")
	}

	// 转换结果
	images := make([]*Image, cfg.BatchCount)
	for i := 0; i < cfg.BatchCount; i++ {
		img := (*bindings.SdImage)(unsafe.Pointer(uintptr(unsafe.Pointer(result)) + uintptr(i)*unsafe.Sizeof(bindings.SdImage{})))
		data := make([]byte, int(img.Width)*int(img.Height)*int(img.Channel))
		copy(data, unsafe.Slice(img.Data, len(data)))

		images[i] = &Image{
			Width:   img.Width,
			Height:  img.Height,
			Channel: img.Channel,
			Data:    data,
		}
	}

	return images, nil
}

// ConvertModel 转换模型格式
func ConvertModel(inputPath, vaePath, outputPath string, outputType bindings.SdType) error {
	success := bindings.Convert(inputPath, vaePath, outputPath, outputType, "", true)
	if !success {
		return fmt.Errorf("failed to convert model")
	}
	return nil
}

// GetDefaultSampleMethod 获取默认采样方法
func (c *Context) GetDefaultSampleMethod() bindings.SampleMethod {
	return bindings.GetDefaultSampleMethod(c.ctx)
}

// GetDefaultScheduler 获取默认调度器
func (c *Context) GetDefaultScheduler(method bindings.SampleMethod) bindings.Scheduler {
	return bindings.GetDefaultScheduler(c.ctx, method)
}

// GetTypeName 获取类型名称
func GetTypeName(t bindings.SdType) string {
	return bindings.GetTypeName(t)
}

// GetSystemInfo 获取系统信息
func GetSystemInfo() string {
	return bindings.GetSystemInfo()
}

// GetVersion 获取版本信息
func GetVersion() string {
	return bindings.GetVersion()
}

// GetCommit 获取提交信息
func GetCommit() string {
	return bindings.GetCommit()
}

// PreprocessCanny 预处理Canny边缘检测
func PreprocessCanny(img *Image, highThreshold, lowThreshold, weak, strong float32, inverse bool) error {
	input := bindings.SdImage{
		Width:   img.Width,
		Height:  img.Height,
		Channel: img.Channel,
		Data:    (*uint8)(unsafe.Pointer(&img.Data[0])),
	}

	success := bindings.PreprocessCanny(input, highThreshold, lowThreshold, weak, strong, inverse)
	if !success {
		return fmt.Errorf("failed to preprocess canny")
	}

	// 更新图像数据
	img.Data = unsafe.Slice(input.Data, int(input.Width)*int(input.Height)*int(input.Channel))

	return nil
}

// VideoGenerationConfig 表示视频生成配置
type VideoGenerationConfig struct {
	Prompt                string
	NegativePrompt        string
	Width                 int
	Height                int
	Seed                  int64
	Strength              float32
	ClipSkip              int
	Loras                 []Lora
	ControlFrames         []Image
	InitImage             *Image
	EndImage              *Image
	VideoFrames           int
	MoeBoundary           float32
	VaceStrength          float32
	Sampler               SamplerConfig
	HighNoiseSampler      SamplerConfig
}

// GenerateVideo 生成视频
func (c *Context) GenerateVideo(cfg VideoGenerationConfig) ([]*Image, error) {
	params := &bindings.SdVidGenParams{}
	bindings.SdVidGenParamsInit(params)

	params.Prompt = bindings.CString(cfg.Prompt)
	params.NegativePrompt = bindings.CString(cfg.NegativePrompt)
	params.Width = cfg.Width
	params.Height = cfg.Height
	params.Seed = cfg.Seed
	params.Strength = cfg.Strength
	params.ClipSkip = cfg.ClipSkip
	params.VideoFrames = cfg.VideoFrames
	params.MoeBoundary = cfg.MoeBoundary
	params.VaceStrength = cfg.VaceStrength

	// 设置采样器参数
	params.SampleParams.Scheduler = cfg.Sampler.Scheduler
	params.SampleParams.SampleMethod = cfg.Sampler.Method
	params.SampleParams.SampleSteps = cfg.Sampler.Steps
	params.SampleParams.Eta = cfg.Sampler.Eta
	params.SampleParams.Guidance.TxtCfg = cfg.Sampler.TxtCfg
	params.SampleParams.Guidance.ImgCfg = cfg.Sampler.ImgCfg
	params.SampleParams.Guidance.DistilledGuidance = cfg.Sampler.DistilledCfg

	// 设置高噪声采样器参数
	params.HighNoiseSampleParams.Scheduler = cfg.HighNoiseSampler.Scheduler
	params.HighNoiseSampleParams.SampleMethod = cfg.HighNoiseSampler.Method
	params.HighNoiseSampleParams.SampleSteps = cfg.HighNoiseSampler.Steps
	params.HighNoiseSampleParams.Eta = cfg.HighNoiseSampler.Eta
	params.HighNoiseSampleParams.Guidance.TxtCfg = cfg.HighNoiseSampler.TxtCfg
	params.HighNoiseSampleParams.Guidance.ImgCfg = cfg.HighNoiseSampler.ImgCfg
	params.HighNoiseSampleParams.Guidance.DistilledGuidance = cfg.HighNoiseSampler.DistilledCfg

	// 设置LoRA
	if len(cfg.Loras) > 0 {
		loras := make([]bindings.SdLora, len(cfg.Loras))
		for i, lora := range cfg.Loras {
			loras[i] = bindings.SdLora{
				Path:        bindings.CString(lora.Path),
				Multiplier:  lora.Multiplier,
				IsHighNoise: lora.IsHighNoise,
			}
		}
		params.Loras = &loras[0]
		params.LoraCount = uint32(len(loras))
	}

	// 设置初始和结束图像
	if cfg.InitImage != nil {
		params.InitImage = bindings.SdImage{
			Width:   cfg.InitImage.Width,
			Height:  cfg.InitImage.Height,
			Channel: cfg.InitImage.Channel,
			Data:    (*uint8)(unsafe.Pointer(&cfg.InitImage.Data[0])),
		}
	}

	if cfg.EndImage != nil {
		params.EndImage = bindings.SdImage{
			Width:   cfg.EndImage.Width,
			Height:  cfg.EndImage.Height,
			Channel: cfg.EndImage.Channel,
			Data:    (*uint8)(unsafe.Pointer(&cfg.EndImage.Data[0])),
		}
	}

	// 设置控制帧
	if len(cfg.ControlFrames) > 0 {
		controlFrames := make([]bindings.SdImage, len(cfg.ControlFrames))
		for i, img := range cfg.ControlFrames {
			controlFrames[i] = bindings.SdImage{
				Width:   img.Width,
				Height:  img.Height,
				Channel: img.Channel,
				Data:    (*uint8)(unsafe.Pointer(&img.Data[0])),
			}
		}
		params.ControlFrames = &controlFrames[0]
		params.ControlFramesSize = len(controlFrames)
	}

	// 生成视频
	var numFramesOut int
	result := bindings.GenerateVideo(c.ctx, params, &numFramesOut)
	if result == nil || numFramesOut == 0 {
		return nil, fmt.Errorf("failed to generate video")
	}

	// 转换结果
	images := make([]*Image, numFramesOut)
	for i := 0; i < numFramesOut; i++ {
		img := (*bindings.SdImage)(unsafe.Pointer(uintptr(unsafe.Pointer(result)) + uintptr(i)*unsafe.Sizeof(bindings.SdImage{})))
		data := make([]byte, int(img.Width)*int(img.Height)*int(img.Channel))
		copy(data, unsafe.Slice(img.Data, len(data)))

		images[i] = &Image{
			Width:   img.Width,
			Height:  img.Height,
			Channel: img.Channel,
			Data:    data,
		}
	}

	return images, nil
}

// GetNumPhysicalCores 获取物理核心数
func GetNumPhysicalCores() int32 {
	return bindings.GetNumPhysicalCores()
}

// SetLogCallback 设置日志回调
func SetLogCallback(cb bindings.SdLogCb, data unsafe.Pointer) {
	bindings.SetLogCallback(cb, data)
}

// SetProgressCallback 设置进度回调
func SetProgressCallback(cb bindings.SdProgressCb, data unsafe.Pointer) {
	bindings.SetProgressCallback(cb, data)
}

// SetPreviewCallback 设置预览回调
func SetPreviewCallback(cb bindings.SdPreviewCb, mode bindings.Preview, interval int, denoised, noisy bool, data unsafe.Pointer) {
	bindings.SetPreviewCallback(cb, mode, interval, denoised, noisy, data)
}

// --- 类型名称 / 字符串互转（完整覆盖 stable-diffusion.h） ---

// StrToSdType 将字符串转换为 SdType
func StrToSdType(s string) bindings.SdType {
	return bindings.StrToSdType(s)
}

// GetRngTypeName 获取 RNG 类型名称
func GetRngTypeName(t bindings.RngType) string {
	return bindings.GetRngTypeName(t)
}

// StrToRngType 将字符串转换为 RngType
func StrToRngType(s string) bindings.RngType {
	return bindings.StrToRngType(s)
}

// GetSampleMethodName 获取采样方法名称
func GetSampleMethodName(t bindings.SampleMethod) string {
	return bindings.GetSampleMethodName(t)
}

// StrToSampleMethod 将字符串转换为 SampleMethod
func StrToSampleMethod(s string) bindings.SampleMethod {
	return bindings.StrToSampleMethod(s)
}

// GetSchedulerName 获取调度器名称
func GetSchedulerName(t bindings.Scheduler) string {
	return bindings.GetSchedulerName(t)
}

// StrToScheduler 将字符串转换为 Scheduler
func StrToScheduler(s string) bindings.Scheduler {
	return bindings.StrToScheduler(s)
}

// GetPredictionName 获取预测类型名称
func GetPredictionName(t bindings.Prediction) string {
	return bindings.GetPredictionName(t)
}

// StrToPrediction 将字符串转换为 Prediction
func StrToPrediction(s string) bindings.Prediction {
	return bindings.StrToPrediction(s)
}

// GetPreviewName 获取预览类型名称
func GetPreviewName(t bindings.Preview) string {
	return bindings.GetPreviewName(t)
}

// StrToPreview 将字符串转换为 Preview
func StrToPreview(s string) bindings.Preview {
	return bindings.StrToPreview(s)
}

// GetLoraApplyModeName 获取 LoRA 应用模式名称
func GetLoraApplyModeName(t bindings.LoraApplyMode) string {
	return bindings.GetLoraApplyModeName(t)
}

// StrToLoraApplyMode 将字符串转换为 LoraApplyMode
func StrToLoraApplyMode(s string) bindings.LoraApplyMode {
	return bindings.StrToLoraApplyMode(s)
}
