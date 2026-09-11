package bindings

import (
	"fmt"
	"os"
	"runtime"
	"unsafe"

	"github.com/ebitengine/purego"
)

// 平台相关的库名
var libNames = map[string]string{
	"darwin":  "libstable-diffusion.dylib",
	"linux":   "libstable-diffusion.so",
	"windows": "stable-diffusion.dll", // 与 CMakeLists.txt 一致
}

// 全局变量，存储加载的库
var lib uintptr

// 保持回调函数的引用，防止被GC
var (
	currentLogCallback      uintptr
	currentProgressCallback uintptr
	currentPreviewCallback  uintptr
)

// 初始化函数，加载对应平台的库
func init() {
	// 优先检查环境变量
	libName := os.Getenv("SD_LIB_PATH")
	if libName == "" {
		name, ok := libNames[runtime.GOOS]
		if ok {
			// 尝试在几个可能的路径搜索
			searchPaths := []string{
				name,
				"./" + name,
				"./stable-diffusion.cpp/build/bin/" + name,
				"./stable-diffusion.cpp/build/bin/Release/" + name,
				"/usr/local/lib/" + name,
				"/usr/lib/" + name,
			}
			for _, p := range searchPaths {
				if _, err := os.Stat(p); err == nil {
					libName = p
					break
				}
			}
			if libName == "" {
				libName = name // 回退到默认库名，让 Dlopen 尝试系统路径
			}
		} else {
			fmt.Printf("Warning: Unsupported platform: %s, using mock implementation\n", runtime.GOOS)
			InitFuncs()
			return
		}
	}

	// 尝试加载库
	var err error
	lib, err = purego.Dlopen(libName, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		fmt.Printf("Warning: Failed to load library: %v, using mock implementation\n", err)
	}

	// 初始化函数指针
	InitFuncs()
}

// 枚举类型定义

type RngType int
type SampleMethod int
type Scheduler int
type Prediction int
type SdType int
type SdLogLevel int
type Preview int
type LoraApplyMode int
type SdCacheMode int

// 常量定义
const (
	// RngType
	STD_DEFAULT_RNG RngType = iota
	CUDA_RNG
	CPU_RNG
	RNG_TYPE_COUNT

	// SampleMethod
	EULER_SAMPLE_METHOD SampleMethod = iota
	EULER_A_SAMPLE_METHOD
	HEUN_SAMPLE_METHOD
	DPM2_SAMPLE_METHOD
	DPMPP2S_A_SAMPLE_METHOD
	DPMPP2M_SAMPLE_METHOD
	DPMPP2Mv2_SAMPLE_METHOD
	IPNDM_SAMPLE_METHOD
	IPNDM_V_SAMPLE_METHOD
	LCM_SAMPLE_METHOD
	DDIM_TRAILING_SAMPLE_METHOD
	TCD_SAMPLE_METHOD
	RES_MULTISTEP_SAMPLE_METHOD
	RES_2S_SAMPLE_METHOD
	SAMPLE_METHOD = iota // Added to match header logic if needed, but SAMPLE_METHOD_COUNT is usually last
	SAMPLE_METHOD_COUNT = RES_2S_SAMPLE_METHOD + 1

	// Scheduler
	DISCRETE_SCHEDULER Scheduler = iota
	KARRAS_SCHEDULER
	EXPONENTIAL_SCHEDULER
	AYS_SCHEDULER
	GITS_SCHEDULER
	SGM_UNIFORM_SCHEDULER
	SIMPLE_SCHEDULER
	SMOOTHSTEP_SCHEDULER
	KL_OPTIMAL_SCHEDULER
	LCM_SCHEDULER
	BONG_TANGENT_SCHEDULER
	SCHEDULER_COUNT

	// Prediction
	EPS_PRED Prediction = iota
	V_PRED
	EDM_V_PRED
	FLOW_PRED
	FLUX_FLOW_PRED
	FLUX2_FLOW_PRED
	PREDICTION_COUNT

	// SdType
	SD_TYPE_F32     SdType = 0
	SD_TYPE_F16     SdType = 1
	SD_TYPE_Q4_0    SdType = 2
	SD_TYPE_Q4_1    SdType = 3
	SD_TYPE_Q5_0    SdType = 6
	SD_TYPE_Q5_1    SdType = 7
	SD_TYPE_Q8_0    SdType = 8
	SD_TYPE_Q8_1    SdType = 9
	SD_TYPE_Q2_K    SdType = 10
	SD_TYPE_Q3_K    SdType = 11
	SD_TYPE_Q4_K    SdType = 12
	SD_TYPE_Q5_K    SdType = 13
	SD_TYPE_Q6_K    SdType = 14
	SD_TYPE_Q8_K    SdType = 15
	SD_TYPE_IQ2_XXS SdType = 16
	SD_TYPE_IQ2_XS  SdType = 17
	SD_TYPE_IQ3_XXS SdType = 18
	SD_TYPE_IQ1_S   SdType = 19
	SD_TYPE_IQ4_NL  SdType = 20
	SD_TYPE_IQ3_S   SdType = 21
	SD_TYPE_IQ2_S   SdType = 22
	SD_TYPE_IQ4_XS  SdType = 23
	SD_TYPE_I8      SdType = 24
	SD_TYPE_I16     SdType = 25
	SD_TYPE_I32     SdType = 26
	SD_TYPE_I64     SdType = 27
	SD_TYPE_F64     SdType = 28
	SD_TYPE_IQ1_M   SdType = 29
	SD_TYPE_BF16    SdType = 30
	SD_TYPE_TQ1_0   SdType = 34
	SD_TYPE_TQ2_0   SdType = 35
	SD_TYPE_MXFP4   SdType = 39
	SD_TYPE_COUNT   SdType = 40

	// SdLogLevel
	SD_LOG_DEBUG SdLogLevel = iota
	SD_LOG_INFO
	SD_LOG_WARN
	SD_LOG_ERROR

	// Preview
	PREVIEW_NONE Preview = iota
	PREVIEW_PROJ
	PREVIEW_TAE
	PREVIEW_VAE
	PREVIEW_COUNT

	// LoraApplyMode
	LORA_APPLY_AUTO LoraApplyMode = iota
	LORA_APPLY_IMMEDIATELY
	LORA_APPLY_AT_RUNTIME
	LORA_APPLY_MODE_COUNT

	// SdCacheMode
	SD_CACHE_DISABLED SdCacheMode = iota
	SD_CACHE_EASYCACHE
	SD_CACHE_UCACHE
	SD_CACHE_DBCACHE
	SD_CACHE_TAYLORSEER
	SD_CACHE_CACHE_DIT
)

// 结构体定义
type SdTilingParams struct {
	Enabled        bool
	TileSizeX      int
	TileSizeY      int
	TargetOverlap  float32
	RelSizeX       float32
	RelSizeY       float32
}

type SdEmbedding struct {
	Name *byte
	Path *byte
}

type SdCtxParams struct {
	ModelPath                   *byte
	ClipLPath                   *byte
	ClipGPath                   *byte
	ClipVisionPath              *byte
	T5xxlPath                   *byte
	LlmPath                     *byte
	LlmVisionPath               *byte
	DiffusionModelPath          *byte
	HighNoiseDiffusionModelPath *byte
	VaePath                     *byte
	TaesdPath                   *byte
	ControlNetPath              *byte
	Embeddings                  *SdEmbedding
	EmbeddingCount              uint32
	PhotoMakerPath              *byte
	TensorTypeRules             *byte
	VaeDecodeOnly               bool
	FreeParamsImmediately       bool
	NThreads                    int
	Wtype                       SdType
	RngType                     RngType
	SamplerRngType              RngType
	Prediction                  Prediction
	LoraApplyMode               LoraApplyMode
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
}

type SdImage struct {
	Width   uint32
	Height  uint32
	Channel uint32
	Data    *uint8
}

type SdSlgParams struct {
	Layers     *int
	LayerCount uintptr
	LayerStart float32
	LayerEnd   float32
	Scale      float32
}

type SdGuidanceParams struct {
	TxtCfg            float32
	ImgCfg            float32
	DistilledGuidance float32
	Slg               SdSlgParams
}

type SdSampleParams struct {
	Guidance          SdGuidanceParams
	Scheduler         Scheduler
	SampleMethod      SampleMethod
	SampleSteps       int
	Eta               float32
	ShiftedTimestep   int
	CustomSigmas      *float32
	CustomSigmasCount int
}

type SdPmParams struct {
	IdImages      *SdImage
	IdImagesCount int
	IdEmbedPath   *byte
	StyleStrength float32
}

type SdCacheParams struct {
	Mode                     SdCacheMode
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
	ScmMask                  *byte
	ScmPolicyDynamic         bool
}

type SdLora struct {
	IsHighNoise bool
	Multiplier  float32
	Path        *byte
}

type SdImgGenParams struct {
	Loras              *SdLora
	LoraCount          uint32
	Prompt             *byte
	NegativePrompt     *byte
	ClipSkip           int
	InitImage          SdImage
	RefImages          *SdImage
	RefImagesCount     int
	AutoResizeRefImage bool
	IncreaseRefIndex   bool
	MaskImage          SdImage
	Width              int
	Height             int
	SampleParams       SdSampleParams
	Strength           float32
	Seed               int64
	BatchCount         int
	ControlImage       SdImage
	ControlStrength    float32
	PmParams           SdPmParams
	VaeTilingParams    SdTilingParams
	Cache              SdCacheParams
}

type SdVidGenParams struct {
	Loras                 *SdLora
	LoraCount             uint32
	Prompt                *byte
	NegativePrompt        *byte
	ClipSkip              int
	InitImage             SdImage
	EndImage              SdImage
	ControlFrames         *SdImage
	ControlFramesSize     int
	Width                 int
	Height                int
	SampleParams          SdSampleParams
	HighNoiseSampleParams SdSampleParams
	MoeBoundary           float32
	Strength              float32
	Seed                  int64
	VideoFrames           int
	VaceStrength          float32
	VaeTilingParams       SdTilingParams
	Cache                 SdCacheParams
}

type SdCtx struct{}
type UpscalerCtx struct{}

// 回调函数类型
type SdLogCb func(level SdLogLevel, text *byte, data unsafe.Pointer)
type SdProgressCb func(step, steps int, time float32, data unsafe.Pointer)
type SdPreviewCb func(step, frameCount int, frames *SdImage, isNoisy bool, data unsafe.Pointer)

// 函数原型定义
var (
	// 设置日志回调
	sdSetLogCallback func(cb uintptr, data unsafe.Pointer)

	// 设置进度回调
	sdSetProgressCallback func(cb uintptr, data unsafe.Pointer)

	// 设置预览回调
	sdSetPreviewCallback func(cb uintptr, mode Preview, interval int, denoised, noisy bool, data unsafe.Pointer)

	// 获取物理核心数
	sdGetNumPhysicalCores func() int32

	// 获取系统信息
	sdGetSystemInfo func() *byte

	// 获取类型名称
	sdTypeName          func(t SdType) *byte
	strToSdType         func(s *byte) SdType
	sdRngTypeName       func(t RngType) *byte
	strToRngType        func(s *byte) RngType
	sdSampleMethodName  func(t SampleMethod) *byte
	strToSampleMethod   func(s *byte) SampleMethod
	sdSchedulerName     func(t Scheduler) *byte
	strToScheduler      func(s *byte) Scheduler
	sdPredictionName    func(t Prediction) *byte
	strToPrediction     func(s *byte) Prediction
	sdPreviewName       func(t Preview) *byte
	strToPreview        func(s *byte) Preview
	sdLoraApplyModeName func(t LoraApplyMode) *byte
	strToLoraApplyMode  func(s *byte) LoraApplyMode

	// 缓存参数初始化
	sdCacheParamsInit func(params *SdCacheParams)

	// 上下文参数初始化
	sdCtxParamsInit  func(params *SdCtxParams)
	sdCtxParamsToStr func(params *SdCtxParams) *byte

	// 创建和释放上下文
	newSdCtx  func(params *SdCtxParams) *SdCtx
	freeSdCtx func(ctx *SdCtx)

	// 采样参数初始化
	sdSampleParamsInit   func(params *SdSampleParams)
	sdSampleParamsToStr  func(params *SdSampleParams) *byte

	// 获取默认采样方法和调度器
	sdGetDefaultSampleMethod func(ctx *SdCtx) SampleMethod
	sdGetDefaultScheduler    func(ctx *SdCtx, method SampleMethod) Scheduler

	// 图像生成参数初始化
	sdImgGenParamsInit  func(params *SdImgGenParams)
	sdImgGenParamsToStr func(params *SdImgGenParams) *byte
	generateImage       func(ctx *SdCtx, params *SdImgGenParams) *SdImage

	// 视频生成参数初始化
	sdVidGenParamsInit func(params *SdVidGenParams)
	generateVideo      func(ctx *SdCtx, params *SdVidGenParams, numFramesOut *int) *SdImage

	// 超分辨率
	newUpscalerCtx   func(esrganPath *byte, offloadParamsToCpu, direct bool, nThreads, tileSize int) *UpscalerCtx
	freeUpscalerCtx  func(ctx *UpscalerCtx)
	upscale          func(ctx *UpscalerCtx, input SdImage, factor uint32) SdImage
	getUpscaleFactor func(ctx *UpscalerCtx) int

	// 转换
	convert func(inputPath, vaePath, outputPath *byte, outputType SdType, tensorTypeRules *byte, convertName bool) bool

	// 预处理
	preprocessCanny func(image SdImage, highThreshold, lowThreshold, weak, strong float32, inverse bool) bool

	// 获取版本信息
	sdCommit  func() *byte
	sdVersion func() *byte
)

// 初始化所有函数指针
func InitFuncs() {
	if lib == 0 {
		// 设置模拟实现
		setMockImplementations()
		return
	}

	purego.RegisterLibFunc(&sdSetLogCallback, lib, "sd_set_log_callback")
	purego.RegisterLibFunc(&sdSetProgressCallback, lib, "sd_set_progress_callback")
	purego.RegisterLibFunc(&sdSetPreviewCallback, lib, "sd_set_preview_callback")
	purego.RegisterLibFunc(&sdGetNumPhysicalCores, lib, "sd_get_num_physical_cores")
	purego.RegisterLibFunc(&sdGetSystemInfo, lib, "sd_get_system_info")
	purego.RegisterLibFunc(&sdTypeName, lib, "sd_type_name")
	purego.RegisterLibFunc(&strToSdType, lib, "str_to_sd_type")
	purego.RegisterLibFunc(&sdRngTypeName, lib, "sd_rng_type_name")
	purego.RegisterLibFunc(&strToRngType, lib, "str_to_rng_type")
	purego.RegisterLibFunc(&sdSampleMethodName, lib, "sd_sample_method_name")
	purego.RegisterLibFunc(&strToSampleMethod, lib, "str_to_sample_method")
	purego.RegisterLibFunc(&sdSchedulerName, lib, "sd_scheduler_name")
	purego.RegisterLibFunc(&strToScheduler, lib, "str_to_scheduler")
	purego.RegisterLibFunc(&sdPredictionName, lib, "sd_prediction_name")
	purego.RegisterLibFunc(&strToPrediction, lib, "str_to_prediction")
	purego.RegisterLibFunc(&sdPreviewName, lib, "sd_preview_name")
	purego.RegisterLibFunc(&strToPreview, lib, "str_to_preview")
	purego.RegisterLibFunc(&sdLoraApplyModeName, lib, "sd_lora_apply_mode_name")
	purego.RegisterLibFunc(&strToLoraApplyMode, lib, "str_to_lora_apply_mode")
	purego.RegisterLibFunc(&sdCacheParamsInit, lib, "sd_cache_params_init")
	purego.RegisterLibFunc(&sdCtxParamsInit, lib, "sd_ctx_params_init")
	purego.RegisterLibFunc(&sdCtxParamsToStr, lib, "sd_ctx_params_to_str")
	purego.RegisterLibFunc(&newSdCtx, lib, "new_sd_ctx")
	purego.RegisterLibFunc(&freeSdCtx, lib, "free_sd_ctx")
	purego.RegisterLibFunc(&sdSampleParamsInit, lib, "sd_sample_params_init")
	purego.RegisterLibFunc(&sdSampleParamsToStr, lib, "sd_sample_params_to_str")
	purego.RegisterLibFunc(&sdGetDefaultSampleMethod, lib, "sd_get_default_sample_method")
	purego.RegisterLibFunc(&sdGetDefaultScheduler, lib, "sd_get_default_scheduler")
	purego.RegisterLibFunc(&sdImgGenParamsInit, lib, "sd_img_gen_params_init")
	purego.RegisterLibFunc(&sdImgGenParamsToStr, lib, "sd_img_gen_params_to_str")
	purego.RegisterLibFunc(&generateImage, lib, "generate_image")
	purego.RegisterLibFunc(&sdVidGenParamsInit, lib, "sd_vid_gen_params_init")
	purego.RegisterLibFunc(&generateVideo, lib, "generate_video")
	purego.RegisterLibFunc(&newUpscalerCtx, lib, "new_upscaler_ctx")
	purego.RegisterLibFunc(&freeUpscalerCtx, lib, "free_upscaler_ctx")
	purego.RegisterLibFunc(&upscale, lib, "upscale")
	purego.RegisterLibFunc(&getUpscaleFactor, lib, "get_upscale_factor")
	purego.RegisterLibFunc(&convert, lib, "convert")
	purego.RegisterLibFunc(&preprocessCanny, lib, "preprocess_canny")
	purego.RegisterLibFunc(&sdCommit, lib, "sd_commit")
	purego.RegisterLibFunc(&sdVersion, lib, "sd_version")
}

// 设置模拟实现
func setMockImplementations() {
	// 模拟系统信息函数
	sdGetSystemInfo = func() *byte {
		return CString("CPU: Mock CPU, Cores: 10, Version: 1.0.0")
	}

	// 模拟版本信息函数
	sdVersion = func() *byte {
		return CString("1.0.0")
	}

	sdCommit = func() *byte {
		return CString("mock-commit-hash")
	}

	// 模拟核心数函数
	sdGetNumPhysicalCores = func() int32 {
		return 10
	}

	// 模拟上下文创建函数
	newSdCtx = func(params *SdCtxParams) *SdCtx {
		return nil
	}

	// 模拟其他函数
	freeSdCtx = func(ctx *SdCtx) {}
	sdGetDefaultSampleMethod = func(ctx *SdCtx) SampleMethod {
		return EULER_A_SAMPLE_METHOD
	}
	sdGetDefaultScheduler = func(ctx *SdCtx, method SampleMethod) Scheduler {
		return KARRAS_SCHEDULER
	}
	generateImage = func(ctx *SdCtx, params *SdImgGenParams) *SdImage {
		return nil
	}
	generateVideo = func(ctx *SdCtx, params *SdVidGenParams, numFramesOut *int) *SdImage {
		return nil
	}
	newUpscalerCtx = func(esrganPath *byte, offloadParamsToCpu, direct bool, nThreads, tileSize int) *UpscalerCtx {
		return nil
	}
	freeUpscalerCtx = func(ctx *UpscalerCtx) {}
	upscale = func(ctx *UpscalerCtx, input SdImage, factor uint32) SdImage {
		return SdImage{}
	}
	getUpscaleFactor = func(ctx *UpscalerCtx) int {
		return 2
	}
	convert = func(inputPath, vaePath, outputPath *byte, outputType SdType, tensorTypeRules *byte, convertName bool) bool {
		return false
	}
	preprocessCanny = func(image SdImage, highThreshold, lowThreshold, weak, strong float32, inverse bool) bool {
		return false
	}

	// 初始化参数函数
	sdCacheParamsInit = func(params *SdCacheParams) {}
	sdCtxParamsInit = func(params *SdCtxParams) {}
	sdCtxParamsToStr = func(params *SdCtxParams) *byte {
		return CString("mock ctx params")
	}
	sdSampleParamsInit = func(params *SdSampleParams) {}
	sdSampleParamsToStr = func(params *SdSampleParams) *byte {
		return CString("mock sample params")
	}
	sdImgGenParamsInit = func(params *SdImgGenParams) {}
	sdImgGenParamsToStr = func(params *SdImgGenParams) *byte {
		return CString("mock img gen params")
	}
	sdVidGenParamsInit = func(params *SdVidGenParams) {}

	// 模拟类型转换函数
	sdTypeName = func(t SdType) *byte {
		return CString("mock type")
	}
	strToSdType = func(s *byte) SdType {
		return SD_TYPE_F16
	}
	sdRngTypeName = func(t RngType) *byte {
		return CString("mock rng type")
	}
	strToRngType = func(s *byte) RngType {
		return CPU_RNG
	}
	sdSampleMethodName = func(t SampleMethod) *byte {
		return CString("mock sample method")
	}
	strToSampleMethod = func(s *byte) SampleMethod {
		return EULER_A_SAMPLE_METHOD
	}
	sdSchedulerName = func(t Scheduler) *byte {
		return CString("mock scheduler")
	}
	strToScheduler = func(s *byte) Scheduler {
		return KARRAS_SCHEDULER
	}
	sdPredictionName = func(t Prediction) *byte {
		return CString("mock prediction")
	}
	strToPrediction = func(s *byte) Prediction {
		return EPS_PRED
	}
	sdPreviewName = func(t Preview) *byte {
		return CString("mock preview")
	}
	strToPreview = func(s *byte) Preview {
		return PREVIEW_NONE
	}
	sdLoraApplyModeName = func(t LoraApplyMode) *byte {
		return CString("mock lora apply mode")
	}
	strToLoraApplyMode = func(s *byte) LoraApplyMode {
		return LORA_APPLY_AUTO
	}

	// 模拟回调设置函数
	sdSetLogCallback = func(cb uintptr, data unsafe.Pointer) {}
	sdSetProgressCallback = func(cb uintptr, data unsafe.Pointer) {}
	sdSetPreviewCallback = func(cb uintptr, mode Preview, interval int, denoised, noisy bool, data unsafe.Pointer) {}
}

// 辅助函数：将Go字符串转换为C字符串
func CString(s string) *byte {
	// 关键修复：添加 NULL 结尾
	b := append([]byte(s), 0)
	return &b[0]
}

// 辅助函数：将C字符串转换为Go字符串
func GoString(c *byte) string {
	if c == nil {
		return ""
	}
	// 关键修复：正确计算字符串长度直到 NULL
	var length int
	p := uintptr(unsafe.Pointer(c))
	for {
		if *(*byte)(unsafe.Pointer(p)) == 0 {
			break
		}
		length++
		p++
	}
	return unsafe.String(c, length)
}

// 公开的API函数

// SetLogCallback 设置日志回调
func SetLogCallback(cb SdLogCb, data unsafe.Pointer) {
	// 使用 purego.NewCallback 包装 Go 函数
	currentLogCallback = purego.NewCallback(func(level SdLogLevel, text *byte, d unsafe.Pointer) {
		cb(level, text, d)
	})
	sdSetLogCallback(currentLogCallback, data)
}

// SetProgressCallback 设置进度回调
func SetProgressCallback(cb SdProgressCb, data unsafe.Pointer) {
	currentProgressCallback = purego.NewCallback(func(step, steps int, time float32, d unsafe.Pointer) {
		cb(step, steps, time, d)
	})
	sdSetProgressCallback(currentProgressCallback, data)
}

// SetPreviewCallback 设置预览回调
func SetPreviewCallback(cb SdPreviewCb, mode Preview, interval int, denoised, noisy bool, data unsafe.Pointer) {
	currentPreviewCallback = purego.NewCallback(func(step, frameCount int, frames *SdImage, isNoisy bool, d unsafe.Pointer) {
		cb(step, frameCount, frames, isNoisy, d)
	})
	sdSetPreviewCallback(currentPreviewCallback, mode, interval, denoised, noisy, data)
}

// GetNumPhysicalCores 获取物理核心数
func GetNumPhysicalCores() int32 {
	return sdGetNumPhysicalCores()
}

// GetSystemInfo 获取系统信息
func GetSystemInfo() string {
	return GoString(sdGetSystemInfo())
}

// GetCommit 获取提交信息
func GetCommit() string {
	return GoString(sdCommit())
}

// GetVersion 获取版本信息
func GetVersion() string {
	return GoString(sdVersion())
}

// CreateSdCtx 创建上下文
func CreateSdCtx(params *SdCtxParams) *SdCtx {
	return newSdCtx(params)
}

// FreeSdCtx 释放上下文
func FreeSdCtx(ctx *SdCtx) {
	freeSdCtx(ctx)
}

// GenerateImage 生成图像
func GenerateImage(ctx *SdCtx, params *SdImgGenParams) *SdImage {
	return generateImage(ctx, params)
}

// GenerateVideo 生成视频
func GenerateVideo(ctx *SdCtx, params *SdVidGenParams, numFramesOut *int) *SdImage {
	return generateVideo(ctx, params, numFramesOut)
}

// CreateUpscalerCtx 创建超分辨率上下文
func CreateUpscalerCtx(esrganPath string, offloadParamsToCpu, direct bool, nThreads, tileSize int) *UpscalerCtx {
	return newUpscalerCtx(CString(esrganPath), offloadParamsToCpu, direct, nThreads, tileSize)
}

// FreeUpscalerCtx 释放超分辨率上下文
func FreeUpscalerCtx(ctx *UpscalerCtx) {
	freeUpscalerCtx(ctx)
}

// Upscale 执行超分辨率
func Upscale(ctx *UpscalerCtx, input SdImage, factor uint32) SdImage {
	return upscale(ctx, input, factor)
}

// GetUpscaleFactor 获取超分辨率因子
func GetUpscaleFactor(ctx *UpscalerCtx) int {
	return getUpscaleFactor(ctx)
}

// Convert 转换模型
func Convert(inputPath, vaePath, outputPath string, outputType SdType, tensorTypeRules string, convertName bool) bool {
	return convert(CString(inputPath), CString(vaePath), CString(outputPath), outputType, CString(tensorTypeRules), convertName)
}

// PreprocessCanny 预处理Canny边缘检测
func PreprocessCanny(image SdImage, highThreshold, lowThreshold, weak, strong float32, inverse bool) bool {
	return preprocessCanny(image, highThreshold, lowThreshold, weak, strong, inverse)
}

// SdImgGenParamsInit 初始化图像生成参数
func SdImgGenParamsInit(params *SdImgGenParams) {
	sdImgGenParamsInit(params)
}

// SdCtxParamsInit 初始化上下文参数
func SdCtxParamsInit(params *SdCtxParams) {
	sdCtxParamsInit(params)
}

// SdSampleParamsInit 初始化采样参数
func SdSampleParamsInit(params *SdSampleParams) {
	sdSampleParamsInit(params)
}

// SdVidGenParamsInit 初始化视频生成参数
func SdVidGenParamsInit(params *SdVidGenParams) {
	sdVidGenParamsInit(params)
}

// SdCacheParamsInit 初始化缓存参数
func SdCacheParamsInit(params *SdCacheParams) {
	sdCacheParamsInit(params)
}

// GetDefaultSampleMethod 获取默认采样方法
func GetDefaultSampleMethod(ctx *SdCtx) SampleMethod {
	return sdGetDefaultSampleMethod(ctx)
}

// GetDefaultScheduler 获取默认调度器
func GetDefaultScheduler(ctx *SdCtx, method SampleMethod) Scheduler {
	return sdGetDefaultScheduler(ctx, method)
}

// GetTypeName 获取类型名称
func GetTypeName(t SdType) string {
	return GoString(sdTypeName(t))
}

// StrToSdType 字符串转 SdType
func StrToSdType(s string) SdType {
	return strToSdType(CString(s))
}

// GetRngTypeName 获取 RNG 类型名称
func GetRngTypeName(t RngType) string {
	return GoString(sdRngTypeName(t))
}

// StrToRngType 字符串转 RngType
func StrToRngType(s string) RngType {
	return strToRngType(CString(s))
}

// GetSampleMethodName 获取采样方法名称
func GetSampleMethodName(t SampleMethod) string {
	return GoString(sdSampleMethodName(t))
}

// StrToSampleMethod 字符串转 SampleMethod
func StrToSampleMethod(s string) SampleMethod {
	return strToSampleMethod(CString(s))
}

// GetSchedulerName 获取调度器名称
func GetSchedulerName(t Scheduler) string {
	return GoString(sdSchedulerName(t))
}

// StrToScheduler 字符串转 Scheduler
func StrToScheduler(s string) Scheduler {
	return strToScheduler(CString(s))
}

// GetPredictionName 获取预测类型名称
func GetPredictionName(t Prediction) string {
	return GoString(sdPredictionName(t))
}

// StrToPrediction 字符串转 Prediction
func StrToPrediction(s string) Prediction {
	return strToPrediction(CString(s))
}

// GetPreviewName 获取预览类型名称
func GetPreviewName(t Preview) string {
	return GoString(sdPreviewName(t))
}

// StrToPreview 字符串转 Preview
func StrToPreview(s string) Preview {
	return strToPreview(CString(s))
}

// GetLoraApplyModeName 获取 LoRA 应用模式名称
func GetLoraApplyModeName(t LoraApplyMode) string {
	return GoString(sdLoraApplyModeName(t))
}

// StrToLoraApplyMode 字符串转 LoraApplyMode
func StrToLoraApplyMode(s string) LoraApplyMode {
	return strToLoraApplyMode(CString(s))
}

// SdCtxParamsToStr 上下文参数转字符串（调试用）
func SdCtxParamsToStr(params *SdCtxParams) string {
	return GoString(sdCtxParamsToStr(params))
}

// SdSampleParamsToStr 采样参数转字符串（调试用）
func SdSampleParamsToStr(params *SdSampleParams) string {
	return GoString(sdSampleParamsToStr(params))
}

// SdImgGenParamsToStr 图像生成参数转字符串（调试用）
func SdImgGenParamsToStr(params *SdImgGenParams) string {
	return GoString(sdImgGenParamsToStr(params))
}
