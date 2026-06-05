package vk

import (
	"encoding/binary"
	"fmt"
	"math"
	"unsafe"
)

type handle unsafe.Pointer

type Instance handle
type PhysicalDevice handle
type Device handle
type Queue handle
type Surface handle
type Swapchain handle
type Image handle
type ImageView handle
type RenderPass handle
type PipelineLayout handle
type Pipeline handle
type PipelineCache handle
type ShaderModule handle
type Framebuffer handle
type CommandPool handle
type CommandBuffer handle
type Semaphore handle
type Fence handle
type Buffer handle
type DeviceMemory handle
type Sampler handle
type DescriptorSetLayout handle
type DescriptorPool handle
type DescriptorSet handle

var (
	NullInstance            Instance
	NullPhysicalDevice      PhysicalDevice
	NullDevice              Device
	NullQueue               Queue
	NullSurface             Surface
	NullSwapchain           Swapchain
	NullImage               Image
	NullImageView           ImageView
	NullRenderPass          RenderPass
	NullPipelineLayout      PipelineLayout
	NullPipeline            Pipeline
	NullPipelineCache       PipelineCache
	NullShaderModule        ShaderModule
	NullFramebuffer         Framebuffer
	NullCommandPool         CommandPool
	NullCommandBuffer       CommandBuffer
	NullSemaphore           Semaphore
	NullFence               Fence
	NullBuffer              Buffer
	NullDeviceMemory        DeviceMemory
	NullSampler             Sampler
	NullDescriptorSetLayout DescriptorSetLayout
	NullDescriptorPool      DescriptorPool
	NullDescriptorSet       DescriptorSet
)

type Result int32

const (
	Success                   Result = 0
	Suboptimal                Result = 1000001003
	ErrorOutOfDate            Result = -1000001004
	ErrorLayerNotPresent      Result = -6
	ErrorExtensionNotPresent  Result = -7
	ErrorInitializationFailed Result = -3
)

type Error Result

func (e Error) Error() string {
	return fmt.Sprintf("Vulkan result %d", int32(e))
}

type Bool32 uint32

const (
	False Bool32 = 0
	True  Bool32 = 1
)

type DeviceSize uint64

const (
	MaxUint32 = ^uint32(0)
	MaxUint64 = ^uint64(0)
)

const (
	ApiVersion10 uint32 = 1 << 22

	KhrSwapchainExtensionName = "VK_KHR_swapchain"
)

func MakeVersion(major, minor, patch uint32) uint32 {
	return (major << 22) | (minor << 12) | patch
}

type StructureType int32

const (
	StructureTypeApplicationInfo                      StructureType = 0
	StructureTypeInstanceCreateInfo                   StructureType = 1
	StructureTypeDeviceQueueCreateInfo                StructureType = 2
	StructureTypeDeviceCreateInfo                     StructureType = 3
	StructureTypeSubmitInfo                           StructureType = 4
	StructureTypeMemoryAllocateInfo                   StructureType = 5
	StructureTypeFenceCreateInfo                      StructureType = 8
	StructureTypeSemaphoreCreateInfo                  StructureType = 9
	StructureTypeBufferCreateInfo                     StructureType = 12
	StructureTypeImageCreateInfo                      StructureType = 14
	StructureTypeImageViewCreateInfo                  StructureType = 15
	StructureTypeShaderModuleCreateInfo               StructureType = 16
	StructureTypePipelineShaderStageCreateInfo        StructureType = 18
	StructureTypePipelineVertexInputStateCreateInfo   StructureType = 19
	StructureTypePipelineInputAssemblyStateCreateInfo StructureType = 20
	StructureTypePipelineViewportStateCreateInfo      StructureType = 22
	StructureTypePipelineRasterizationStateCreateInfo StructureType = 23
	StructureTypePipelineMultisampleStateCreateInfo   StructureType = 24
	StructureTypePipelineColorBlendStateCreateInfo    StructureType = 26
	StructureTypeGraphicsPipelineCreateInfo           StructureType = 28
	StructureTypePipelineLayoutCreateInfo             StructureType = 30
	StructureTypeSamplerCreateInfo                    StructureType = 31
	StructureTypeDescriptorSetLayoutCreateInfo        StructureType = 32
	StructureTypeDescriptorPoolCreateInfo             StructureType = 33
	StructureTypeDescriptorSetAllocateInfo            StructureType = 34
	StructureTypeWriteDescriptorSet                   StructureType = 35
	StructureTypeFramebufferCreateInfo                StructureType = 37
	StructureTypeRenderPassCreateInfo                 StructureType = 38
	StructureTypeCommandPoolCreateInfo                StructureType = 39
	StructureTypeCommandBufferAllocateInfo            StructureType = 40
	StructureTypeCommandBufferBeginInfo               StructureType = 42
	StructureTypeRenderPassBeginInfo                  StructureType = 43
	StructureTypeImageMemoryBarrier                   StructureType = 45
	StructureTypeSwapchainCreateInfo                  StructureType = 1000001000
	StructureTypePresentInfo                          StructureType = 1000001001
)

type (
	InstanceCreateFlags                 uint32
	DeviceQueueCreateFlags              uint32
	DeviceCreateFlags                   uint32
	BufferCreateFlags                   uint32
	BufferUsageFlags                    uint32
	BufferUsageFlagBits                 uint32
	MemoryPropertyFlags                 uint32
	MemoryPropertyFlagBits              uint32
	ImageCreateFlags                    uint32
	ImageUsageFlags                     uint32
	ImageUsageFlagBits                  uint32
	ImageAspectFlags                    uint32
	ImageAspectFlagBits                 uint32
	QueueFlags                          uint32
	QueueFlagBits                       uint32
	CommandPoolCreateFlags              uint32
	CommandPoolCreateFlagBits           uint32
	CommandBufferUsageFlags             uint32
	CommandBufferUsageFlagBits          uint32
	FenceCreateFlags                    uint32
	FenceCreateFlagBits                 uint32
	PipelineStageFlags                  uint32
	PipelineStageFlagBits               uint32
	AccessFlags                         uint32
	AccessFlagBits                      uint32
	DependencyFlags                     uint32
	ColorComponentFlags                 uint32
	ColorComponentFlagBits              uint32
	CullModeFlags                       uint32
	CullModeFlagBits                    uint32
	CompositeAlphaFlags                 uint32
	CompositeAlphaFlagBits              uint32
	ShaderStageFlags                    uint32
	ShaderStageFlagBits                 uint32
	DescriptorSetLayoutCreateFlags      uint32
	DescriptorPoolCreateFlags           uint32
	PipelineLayoutCreateFlags           uint32
	PipelineVertexInputStateCreateFlags uint32
	PipelineShaderStageCreateFlags      uint32
)

const (
	QueueGraphicsBit QueueFlagBits = 1

	BufferUsageTransferSrcBit  BufferUsageFlagBits = 1
	BufferUsageTransferDstBit  BufferUsageFlagBits = 2
	BufferUsageVertexBufferBit BufferUsageFlagBits = 128
	BufferUsageIndexBufferBit  BufferUsageFlagBits = 64

	MemoryPropertyDeviceLocalBit  MemoryPropertyFlagBits = 1
	MemoryPropertyHostVisibleBit  MemoryPropertyFlagBits = 2
	MemoryPropertyHostCoherentBit MemoryPropertyFlagBits = 4

	ImageUsageTransferDstBit     ImageUsageFlagBits  = 2
	ImageUsageSampledBit         ImageUsageFlagBits  = 4
	ImageUsageColorAttachmentBit ImageUsageFlagBits  = 16
	ImageAspectColorBit          ImageAspectFlagBits = 1

	CommandPoolCreateResetCommandBufferBit CommandPoolCreateFlagBits  = 2
	CommandBufferUsageOneTimeSubmitBit     CommandBufferUsageFlagBits = 1
	FenceCreateSignaledBit                 FenceCreateFlagBits        = 1

	PipelineStageTopOfPipeBit             PipelineStageFlagBits = 1
	PipelineStageTransferBit              PipelineStageFlagBits = 4096
	PipelineStageFragmentShaderBit        PipelineStageFlagBits = 128
	PipelineStageColorAttachmentOutputBit PipelineStageFlagBits = 1024

	AccessShaderReadBit           AccessFlagBits = 32
	AccessColorAttachmentWriteBit AccessFlagBits = 256
	AccessTransferWriteBit        AccessFlagBits = 4096

	ColorComponentRBit ColorComponentFlagBits = 1
	ColorComponentGBit ColorComponentFlagBits = 2
	ColorComponentBBit ColorComponentFlagBits = 4
	ColorComponentABit ColorComponentFlagBits = 8

	CullModeNone CullModeFlagBits = 0

	CompositeAlphaOpaqueBit         CompositeAlphaFlagBits = 1
	CompositeAlphaPreMultipliedBit  CompositeAlphaFlagBits = 2
	CompositeAlphaPostMultipliedBit CompositeAlphaFlagBits = 4
	CompositeAlphaInheritBit        CompositeAlphaFlagBits = 8

	ShaderStageVertexBit   ShaderStageFlagBits = 1
	ShaderStageFragmentBit ShaderStageFlagBits = 16
)

type Format int32

const (
	FormatUndefined          Format = 0
	FormatB8g8r8a8Srgb       Format = 50
	FormatR8g8b8a8Unorm      Format = 37
	FormatR32g32Sfloat       Format = 103
	FormatR32g32b32a32Sfloat Format = 109
)

type ColorSpace int32

const ColorSpaceSrgbNonlinear ColorSpace = 1000104001

type PresentMode int32

const (
	PresentModeImmediate PresentMode = 0
	PresentModeMailbox   PresentMode = 1
	PresentModeFifo      PresentMode = 2
)

type SharingMode int32

const (
	SharingModeExclusive  SharingMode = 0
	SharingModeConcurrent SharingMode = 1
)

type ImageType int32

const ImageType2d ImageType = 1

type ImageTiling int32

const ImageTilingOptimal ImageTiling = 0

type ImageViewType int32

const ImageViewType2d ImageViewType = 1

type ImageLayout int32

const (
	ImageLayoutUndefined              ImageLayout = 0
	ImageLayoutGeneral                ImageLayout = 1
	ImageLayoutColorAttachmentOptimal ImageLayout = 2
	ImageLayoutShaderReadOnlyOptimal  ImageLayout = 5
	ImageLayoutTransferDstOptimal     ImageLayout = 7
	ImageLayoutPresentSrc             ImageLayout = 1000001002
)

type SampleCountFlagBits uint32

const SampleCount1Bit SampleCountFlagBits = 1

type AttachmentLoadOp int32
type AttachmentStoreOp int32

const (
	AttachmentLoadOpLoad     AttachmentLoadOp = 0
	AttachmentLoadOpClear    AttachmentLoadOp = 1
	AttachmentLoadOpDontCare AttachmentLoadOp = 2

	AttachmentStoreOpStore    AttachmentStoreOp = 0
	AttachmentStoreOpDontCare AttachmentStoreOp = 1
)

type PipelineBindPoint int32

const PipelineBindPointGraphics PipelineBindPoint = 0

type PrimitiveTopology int32

const PrimitiveTopologyTriangleList PrimitiveTopology = 3

type PolygonMode int32

const PolygonModeFill PolygonMode = 0

type FrontFace int32

const FrontFaceClockwise FrontFace = 1

type VertexInputRate int32

const VertexInputRateVertex VertexInputRate = 0

type ShaderStageFlagBitsC = ShaderStageFlagBits

type DescriptorType int32

const DescriptorTypeCombinedImageSampler DescriptorType = 1

type Filter int32

const FilterNearest Filter = 0

type SamplerMipmapMode int32

const SamplerMipmapModeNearest SamplerMipmapMode = 0

type SamplerAddressMode int32

const SamplerAddressModeRepeat SamplerAddressMode = 0

type BorderColor int32

const BorderColorIntOpaqueBlack BorderColor = 3

type CompareOp int32

type CommandBufferLevel int32

const CommandBufferLevelPrimary CommandBufferLevel = 0

type SubpassContents int32

const SubpassContentsInline SubpassContents = 0

type IndexType int32

const IndexTypeUint16 IndexType = 0

const (
	SubpassExternal    = ^uint32(0)
	QueueFamilyIgnored = ^uint32(0)
)

type ApplicationInfo struct {
	SType              StructureType
	PApplicationName   string
	ApplicationVersion uint32
	PEngineName        string
	EngineVersion      uint32
	ApiVersion         uint32
}

type InstanceCreateInfo struct {
	SType                   StructureType
	Flags                   InstanceCreateFlags
	PApplicationInfo        *ApplicationInfo
	EnabledExtensionCount   uint32
	PpEnabledExtensionNames []string
	EnabledLayerCount       uint32
	PpEnabledLayerNames     []string
}

type DeviceQueueCreateInfo struct {
	SType            StructureType
	Flags            DeviceQueueCreateFlags
	QueueFamilyIndex uint32
	QueueCount       uint32
	PQueuePriorities []float32
}

type DeviceCreateInfo struct {
	SType                   StructureType
	Flags                   DeviceCreateFlags
	QueueCreateInfoCount    uint32
	PQueueCreateInfos       []DeviceQueueCreateInfo
	EnabledExtensionCount   uint32
	PpEnabledExtensionNames []string
}

type Extent2D struct {
	Width  uint32
	Height uint32
}

type Extent3D struct {
	Width  uint32
	Height uint32
	Depth  uint32
}

type Offset2D struct {
	X int32
	Y int32
}

type Offset3D struct {
	X int32
	Y int32
	Z int32
}

type Rect2D struct {
	Offset Offset2D
	Extent Extent2D
}

type SurfaceCapabilities struct {
	MinImageCount           uint32
	MaxImageCount           uint32
	CurrentExtent           Extent2D
	MinImageExtent          Extent2D
	MaxImageExtent          Extent2D
	CurrentTransform        SurfaceTransformFlagBits
	SupportedCompositeAlpha CompositeAlphaFlags
}

func (s *SurfaceCapabilities) Deref() {}

type SurfaceTransformFlagBits uint32

type SurfaceFormat struct {
	Format     Format
	ColorSpace ColorSpace
}

func (s *SurfaceFormat) Deref() {}

type QueueFamilyProperties struct {
	QueueFlags QueueFlags
	QueueCount uint32
}

func (q *QueueFamilyProperties) Deref() {}

type ExtensionProperties struct {
	ExtensionName [256]byte
	SpecVersion   uint32
}

func (e *ExtensionProperties) Deref() {}

type LayerProperties struct {
	LayerName [256]byte
}

func (l *LayerProperties) Deref() {}

type SwapchainCreateInfo struct {
	SType                 StructureType
	Surface               Surface
	MinImageCount         uint32
	ImageFormat           Format
	ImageColorSpace       ColorSpace
	ImageExtent           Extent2D
	ImageArrayLayers      uint32
	ImageUsage            ImageUsageFlags
	ImageSharingMode      SharingMode
	QueueFamilyIndexCount uint32
	PQueueFamilyIndices   []uint32
	PreTransform          SurfaceTransformFlagBits
	CompositeAlpha        CompositeAlphaFlagBits
	PresentMode           PresentMode
	Clipped               Bool32
	OldSwapchain          Swapchain
}

type ComponentMapping struct{}

type ImageSubresourceRange struct {
	AspectMask     ImageAspectFlags
	BaseMipLevel   uint32
	LevelCount     uint32
	BaseArrayLayer uint32
	LayerCount     uint32
}

type ImageViewCreateInfo struct {
	SType            StructureType
	Image            Image
	ViewType         ImageViewType
	Format           Format
	SubresourceRange ImageSubresourceRange
}

type AttachmentDescription struct {
	Format         Format
	Samples        SampleCountFlagBits
	LoadOp         AttachmentLoadOp
	StoreOp        AttachmentStoreOp
	StencilLoadOp  AttachmentLoadOp
	StencilStoreOp AttachmentStoreOp
	InitialLayout  ImageLayout
	FinalLayout    ImageLayout
}

type AttachmentReference struct {
	Attachment uint32
	Layout     ImageLayout
}

type SubpassDescription struct {
	PipelineBindPoint    PipelineBindPoint
	ColorAttachmentCount uint32
	PColorAttachments    []AttachmentReference
}

type SubpassDependency struct {
	SrcSubpass    uint32
	DstSubpass    uint32
	SrcStageMask  PipelineStageFlags
	DstStageMask  PipelineStageFlags
	SrcAccessMask AccessFlags
	DstAccessMask AccessFlags
}

type RenderPassCreateInfo struct {
	SType           StructureType
	AttachmentCount uint32
	PAttachments    []AttachmentDescription
	SubpassCount    uint32
	PSubpasses      []SubpassDescription
	DependencyCount uint32
	PDependencies   []SubpassDependency
}

type ShaderModuleCreateInfo struct {
	SType    StructureType
	CodeSize uint
	PCode    []uint32
}

type PipelineShaderStageCreateInfo struct {
	SType  StructureType
	Stage  ShaderStageFlagBits
	Module ShaderModule
	PName  string
}

type VertexInputBindingDescription struct {
	Binding   uint32
	Stride    uint32
	InputRate VertexInputRate
}

type VertexInputAttributeDescription struct {
	Location uint32
	Binding  uint32
	Format   Format
	Offset   uint32
}

type PipelineVertexInputStateCreateInfo struct {
	SType                           StructureType
	VertexBindingDescriptionCount   uint32
	PVertexBindingDescriptions      []VertexInputBindingDescription
	VertexAttributeDescriptionCount uint32
	PVertexAttributeDescriptions    []VertexInputAttributeDescription
}

type PipelineInputAssemblyStateCreateInfo struct {
	SType    StructureType
	Topology PrimitiveTopology
}

type Viewport struct {
	X        float32
	Y        float32
	Width    float32
	Height   float32
	MinDepth float32
	MaxDepth float32
}

type PipelineViewportStateCreateInfo struct {
	SType         StructureType
	ViewportCount uint32
	PViewports    []Viewport
	ScissorCount  uint32
	PScissors     []Rect2D
}

type PipelineRasterizationStateCreateInfo struct {
	SType       StructureType
	PolygonMode PolygonMode
	CullMode    CullModeFlags
	FrontFace   FrontFace
	LineWidth   float32
}

type PipelineMultisampleStateCreateInfo struct {
	SType                StructureType
	RasterizationSamples SampleCountFlagBits
}

type PipelineColorBlendAttachmentState struct {
	ColorWriteMask ColorComponentFlags
}

type PipelineColorBlendStateCreateInfo struct {
	SType           StructureType
	AttachmentCount uint32
	PAttachments    []PipelineColorBlendAttachmentState
}

type PipelineLayoutCreateInfo struct {
	SType          StructureType
	SetLayoutCount uint32
	PSetLayouts    []DescriptorSetLayout
}

type GraphicsPipelineCreateInfo struct {
	SType               StructureType
	StageCount          uint32
	PStages             []PipelineShaderStageCreateInfo
	PVertexInputState   *PipelineVertexInputStateCreateInfo
	PInputAssemblyState *PipelineInputAssemblyStateCreateInfo
	PViewportState      *PipelineViewportStateCreateInfo
	PRasterizationState *PipelineRasterizationStateCreateInfo
	PMultisampleState   *PipelineMultisampleStateCreateInfo
	PColorBlendState    *PipelineColorBlendStateCreateInfo
	Layout              PipelineLayout
	RenderPass          RenderPass
	Subpass             uint32
}

type FramebufferCreateInfo struct {
	SType           StructureType
	RenderPass      RenderPass
	AttachmentCount uint32
	PAttachments    []ImageView
	Width           uint32
	Height          uint32
	Layers          uint32
}

type CommandPoolCreateInfo struct {
	SType            StructureType
	Flags            CommandPoolCreateFlags
	QueueFamilyIndex uint32
}

type CommandBufferAllocateInfo struct {
	SType              StructureType
	CommandPool        CommandPool
	Level              CommandBufferLevel
	CommandBufferCount uint32
}

type SemaphoreCreateInfo struct {
	SType StructureType
}

type FenceCreateInfo struct {
	SType StructureType
	Flags FenceCreateFlags
}

type CommandBufferBeginInfo struct {
	SType StructureType
	Flags CommandBufferUsageFlags
}

type ClearValue [16]byte

func NewClearValue(color []float32) ClearValue {
	var value ClearValue
	for i := 0; i < len(color) && i < 4; i++ {
		binary.LittleEndian.PutUint32(value[i*4:], math.Float32bits(color[i]))
	}
	return value
}

func (c ClearValue) color() [4]float32 {
	return [4]float32{
		math.Float32frombits(binary.LittleEndian.Uint32(c[0:])),
		math.Float32frombits(binary.LittleEndian.Uint32(c[4:])),
		math.Float32frombits(binary.LittleEndian.Uint32(c[8:])),
		math.Float32frombits(binary.LittleEndian.Uint32(c[12:])),
	}
}

type RenderPassBeginInfo struct {
	SType           StructureType
	RenderPass      RenderPass
	Framebuffer     Framebuffer
	RenderArea      Rect2D
	ClearValueCount uint32
	PClearValues    []ClearValue
}

type PresentInfo struct {
	SType              StructureType
	WaitSemaphoreCount uint32
	PWaitSemaphores    []Semaphore
	SwapchainCount     uint32
	PSwapchains        []Swapchain
	PImageIndices      []uint32
}

type SubmitInfo struct {
	SType                StructureType
	WaitSemaphoreCount   uint32
	PWaitSemaphores      []Semaphore
	PWaitDstStageMask    []PipelineStageFlags
	CommandBufferCount   uint32
	PCommandBuffers      []CommandBuffer
	SignalSemaphoreCount uint32
	PSignalSemaphores    []Semaphore
}

type MemoryAllocateInfo struct {
	SType           StructureType
	AllocationSize  DeviceSize
	MemoryTypeIndex uint32
}

type MemoryRequirements struct {
	Size           DeviceSize
	Alignment      DeviceSize
	MemoryTypeBits uint32
}

func (m *MemoryRequirements) Deref() {}

type MemoryType struct {
	PropertyFlags MemoryPropertyFlags
	HeapIndex     uint32
}

func (m *MemoryType) Deref() {}

type PhysicalDeviceMemoryProperties struct {
	MemoryTypeCount uint32
	MemoryTypes     [32]MemoryType
}

func (p *PhysicalDeviceMemoryProperties) Deref() {}

type BufferCreateInfo struct {
	SType       StructureType
	Size        DeviceSize
	Usage       BufferUsageFlags
	SharingMode SharingMode
}

type ImageCreateInfo struct {
	SType         StructureType
	ImageType     ImageType
	Format        Format
	Extent        Extent3D
	MipLevels     uint32
	ArrayLayers   uint32
	Samples       SampleCountFlagBits
	Tiling        ImageTiling
	Usage         ImageUsageFlags
	SharingMode   SharingMode
	InitialLayout ImageLayout
}

type SamplerCreateInfo struct {
	SType                   StructureType
	MagFilter               Filter
	MinFilter               Filter
	MipmapMode              SamplerMipmapMode
	AddressModeU            SamplerAddressMode
	AddressModeV            SamplerAddressMode
	AddressModeW            SamplerAddressMode
	MaxLod                  float32
	BorderColor             BorderColor
	UnnormalizedCoordinates Bool32
}

type DescriptorSetLayoutBinding struct {
	Binding         uint32
	DescriptorType  DescriptorType
	DescriptorCount uint32
	StageFlags      ShaderStageFlags
}

type DescriptorSetLayoutCreateInfo struct {
	SType        StructureType
	BindingCount uint32
	PBindings    []DescriptorSetLayoutBinding
}

type DescriptorPoolSize struct {
	Type            DescriptorType
	DescriptorCount uint32
}

type DescriptorPoolCreateInfo struct {
	SType         StructureType
	MaxSets       uint32
	PoolSizeCount uint32
	PPoolSizes    []DescriptorPoolSize
}

type DescriptorSetAllocateInfo struct {
	SType              StructureType
	DescriptorPool     DescriptorPool
	DescriptorSetCount uint32
	PSetLayouts        []DescriptorSetLayout
}

type DescriptorImageInfo struct {
	Sampler     Sampler
	ImageView   ImageView
	ImageLayout ImageLayout
}

type DescriptorBufferInfo struct{}

type WriteDescriptorSet struct {
	SType           StructureType
	DstSet          DescriptorSet
	DstBinding      uint32
	DstArrayElement uint32
	DescriptorCount uint32
	DescriptorType  DescriptorType
	PImageInfo      []DescriptorImageInfo
	PBufferInfo     []DescriptorBufferInfo
}

type CopyDescriptorSet struct{}

type BufferCopy struct {
	SrcOffset DeviceSize
	DstOffset DeviceSize
	Size      DeviceSize
}

type ImageSubresourceLayers struct {
	AspectMask     ImageAspectFlags
	MipLevel       uint32
	BaseArrayLayer uint32
	LayerCount     uint32
}

type BufferImageCopy struct {
	BufferOffset      DeviceSize
	BufferRowLength   uint32
	BufferImageHeight uint32
	ImageSubresource  ImageSubresourceLayers
	ImageOffset       Offset3D
	ImageExtent       Extent3D
}

type MemoryBarrier struct{}
type BufferMemoryBarrier struct{}

type ImageMemoryBarrier struct {
	SType               StructureType
	SrcAccessMask       AccessFlags
	DstAccessMask       AccessFlags
	OldLayout           ImageLayout
	NewLayout           ImageLayout
	SrcQueueFamilyIndex uint32
	DstQueueFamilyIndex uint32
	Image               Image
	SubresourceRange    ImageSubresourceRange
}

type AllocationCallbacks struct{}

func ToString(raw []byte) string {
	for i, b := range raw {
		if b == 0 {
			return string(raw[:i])
		}
	}
	return string(raw)
}

func Memcopy(dst unsafe.Pointer, src []byte) int {
	return copy(unsafe.Slice((*byte)(dst), len(src)), src)
}
