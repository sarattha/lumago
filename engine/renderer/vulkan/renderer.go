package vulkan

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/sarattha/lumago/engine/graphics"
	"github.com/sarattha/lumago/engine/platform/desktop"
	erenderer "github.com/sarattha/lumago/engine/renderer"
	vk "github.com/sarattha/lumago/engine/renderer/vulkan/internal/vk"
)

const framesInFlight = 2

type Config struct {
	Window          *desktop.Window
	ShaderDirectory string
	Validation      bool
}

type spriteFrameResources struct {
	vertexBuffer   vk.Buffer
	vertexMemory   vk.DeviceMemory
	indexBuffer    vk.Buffer
	indexMemory    vk.DeviceMemory
	vertexCapacity int
	indexCapacity  int
	vertexUpload   []byte
	indexUpload    []byte
}

type Renderer struct {
	window *desktop.Window

	instance vk.Instance
	surface  vk.Surface

	physicalDevice vk.PhysicalDevice
	device         vk.Device
	graphicsQueue  vk.Queue
	presentQueue   vk.Queue
	graphicsFamily uint32
	presentFamily  uint32

	swapchain       vk.Swapchain
	swapchainFormat vk.Format
	swapchainExtent vk.Extent2D
	swapchainImages []vk.Image
	imageViews      []vk.ImageView
	framebuffers    []vk.Framebuffer

	renderPass     vk.RenderPass
	pipelineLayout vk.PipelineLayout
	pipeline       vk.Pipeline
	commandPool    vk.CommandPool
	commandBuffers []vk.CommandBuffer

	spriteFrames        [framesInFlight]spriteFrameResources
	textureImage        vk.Image
	textureMemory       vk.DeviceMemory
	textureImageView    vk.ImageView
	textureSampler      vk.Sampler
	descriptorSetLayout vk.DescriptorSetLayout
	descriptorPool      vk.DescriptorPool
	descriptorSet       vk.DescriptorSet

	imageAvailable []vk.Semaphore
	renderFinished []vk.Semaphore
	inFlight       []vk.Fence
	frame          int

	shaderDirectory string
	validation      bool
	frameStarted    bool
	imageIndex      uint32
	frameCamera     graphics.Camera2D
	pendingBatch    graphics.SpriteBatch
	pendingLights   []graphics.Light2D
	lightUpload     []byte
	lightingConfig  graphics.LightingConfig2D
	lightingTargets lightingRenderTargets
	lightingBuffers lightingRenderBuffers
	lightingPasses  []lightingPass
	stats           erenderer.FrameStats
}

func NewRenderer(config Config) (*Renderer, error) {
	if config.Window == nil {
		return nil, errors.New("vulkan: window is required")
	}
	if config.ShaderDirectory == "" {
		config.ShaderDirectory = "shaders/bin"
	}

	runtime.LockOSThread()
	if err := vk.SetDefaultGetInstanceProcAddr(); err != nil {
		vk.SetGetInstanceProcAddr(desktop.VulkanProcAddr())
	}
	if err := vk.Init(); err != nil {
		return nil, err
	}

	r := &Renderer{
		window:          config.Window,
		shaderDirectory: config.ShaderDirectory,
		validation:      config.Validation,
	}
	if err := r.init(); err != nil {
		r.Close()
		return nil, err
	}
	return r, nil
}

func (r *Renderer) BeginFrame(camera graphics.Camera2D) error {
	if r.swapchain == vk.NullSwapchain {
		return errors.New("vulkan: swapchain is not initialized")
	}

	fence := r.inFlight[r.frame]
	if err := check(vk.WaitForFences(r.device, 1, []vk.Fence{fence}, vk.True, vk.MaxUint64), "wait for frame fence"); err != nil {
		return err
	}

	result := vk.AcquireNextImage(r.device, r.swapchain, vk.MaxUint64, r.imageAvailable[r.frame], vk.NullFence, &r.imageIndex)
	if result == vk.ErrorOutOfDate {
		return r.recreateSwapchain()
	}
	if result != vk.Success && result != vk.Suboptimal {
		return check(result, "acquire swapchain image")
	}

	if err := check(vk.ResetFences(r.device, 1, []vk.Fence{fence}), "reset frame fence"); err != nil {
		return err
	}
	if err := check(vk.ResetCommandBuffer(r.commandBuffers[r.frame], 0), "reset command buffer"); err != nil {
		return err
	}
	r.frameStarted = true
	r.frameCamera = camera
	r.pendingBatch = graphics.SpriteBatch{}
	r.pendingLights = r.pendingLights[:0]
	r.lightUpload = r.lightUpload[:0]
	r.lightingConfig = graphics.DefaultLightingConfig2D()
	r.lightingPasses = defaultLightingPasses(r.lightingConfig.DebugView)
	r.stats = erenderer.FrameStats{}
	return nil
}

func (r *Renderer) SubmitSpriteBatch(batch graphics.SpriteBatch) error {
	r.pendingBatch = batch
	r.stats = erenderer.FrameStats{
		Sprites:   batch.Stats.SpriteCount,
		DrawCalls: batch.Stats.DrawCalls,
		Vertices:  batch.Stats.VertexCount,
		Indices:   batch.Stats.IndexCount,
	}
	return r.uploadSpriteBatch(batch)
}

func (r *Renderer) ConfigureLighting(config graphics.LightingConfig2D) error {
	r.lightingConfig = config.WithDefaults()
	r.lightingPasses = defaultLightingPasses(r.lightingConfig.DebugView)
	return nil
}

func (r *Renderer) SubmitLights(lights []graphics.Light2D) error {
	r.pendingLights = prepareLightsForFrame(r.pendingLights[:0], lights, r.frameCamera)
	r.lightUpload = packLights(r.lightUpload, r.pendingLights)
	r.stats.Lights = len(r.pendingLights)
	return nil
}

func (r *Renderer) SubmitOccluders(occluders []graphics.Occluder2D) error {
	return nil
}

func (r *Renderer) Stats() erenderer.FrameStats {
	return r.stats
}

func (r *Renderer) EndFrame() error {
	if !r.frameStarted {
		return nil
	}
	r.frameStarted = false
	if err := r.recordCommandBuffer(r.commandBuffers[r.frame], r.imageIndex); err != nil {
		return err
	}

	waitStages := []vk.PipelineStageFlags{vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit)}
	submit := vk.SubmitInfo{
		SType:                vk.StructureTypeSubmitInfo,
		WaitSemaphoreCount:   1,
		PWaitSemaphores:      []vk.Semaphore{r.imageAvailable[r.frame]},
		PWaitDstStageMask:    waitStages,
		CommandBufferCount:   1,
		PCommandBuffers:      []vk.CommandBuffer{r.commandBuffers[r.frame]},
		SignalSemaphoreCount: 1,
		PSignalSemaphores:    []vk.Semaphore{r.renderFinished[r.frame]},
	}
	if err := check(vk.QueueSubmit(r.graphicsQueue, 1, []vk.SubmitInfo{submit}, r.inFlight[r.frame]), "submit draw commands"); err != nil {
		return err
	}

	present := vk.PresentInfo{
		SType:              vk.StructureTypePresentInfo,
		WaitSemaphoreCount: 1,
		PWaitSemaphores:    []vk.Semaphore{r.renderFinished[r.frame]},
		SwapchainCount:     1,
		PSwapchains:        []vk.Swapchain{r.swapchain},
		PImageIndices:      []uint32{r.imageIndex},
	}
	result := vk.QueuePresent(r.presentQueue, &present)
	if result == vk.ErrorOutOfDate || result == vk.Suboptimal {
		if err := r.recreateSwapchain(); err != nil {
			return err
		}
	} else if result != vk.Success {
		return check(result, "present swapchain image")
	}

	r.frame = (r.frame + 1) % framesInFlight
	return nil
}

func (r *Renderer) Resize(width, height int) error {
	if width <= 0 || height <= 0 || r.swapchain == vk.NullSwapchain {
		return nil
	}
	return r.recreateSwapchain()
}

func (r *Renderer) Close() error {
	if r.device != vk.NullDevice {
		vk.DeviceWaitIdle(r.device)
	}
	r.cleanupSwapchain()
	r.cleanupQuadResources()
	for i := 0; i < len(r.imageAvailable); i++ {
		if r.imageAvailable[i] != vk.NullSemaphore {
			vk.DestroySemaphore(r.device, r.imageAvailable[i], nil)
		}
		if r.renderFinished[i] != vk.NullSemaphore {
			vk.DestroySemaphore(r.device, r.renderFinished[i], nil)
		}
		if r.inFlight[i] != vk.NullFence {
			vk.DestroyFence(r.device, r.inFlight[i], nil)
		}
	}
	if r.commandPool != vk.NullCommandPool {
		vk.DestroyCommandPool(r.device, r.commandPool, nil)
	}
	if r.device != vk.NullDevice {
		vk.DestroyDevice(r.device, nil)
	}
	if r.surface != vk.NullSurface {
		vk.DestroySurface(r.instance, r.surface, nil)
	}
	if r.instance != vk.NullInstance {
		vk.DestroyInstance(r.instance, nil)
	}
	return nil
}

func (r *Renderer) init() error {
	if err := r.createInstance(); err != nil {
		return err
	}
	surface, err := r.window.CreateSurface(vk.InstanceHandle(r.instance))
	if err != nil {
		return err
	}
	r.surface = vk.SurfaceFromPointer(surface)
	if err := r.pickPhysicalDevice(); err != nil {
		return err
	}
	if err := r.createLogicalDevice(); err != nil {
		return err
	}
	if err := r.createSwapchain(); err != nil {
		return err
	}
	r.lightingTargets = defaultLightingRenderTargets(r.swapchainExtent, r.swapchainFormat)
	r.lightingConfig = graphics.DefaultLightingConfig2D()
	r.lightingPasses = defaultLightingPasses(r.lightingConfig.DebugView)
	if err := r.createLightingRenderBuffers(); err != nil {
		return err
	}
	if err := r.createRenderPass(); err != nil {
		return err
	}
	if err := r.createCommandPool(); err != nil {
		return err
	}
	if err := r.createQuadResources(); err != nil {
		return err
	}
	if err := r.createPipeline(); err != nil {
		return err
	}
	if err := r.createFramebuffers(); err != nil {
		return err
	}
	if err := r.createCommandBuffers(); err != nil {
		return err
	}
	return r.createSyncObjects()
}

func (r *Renderer) createInstance() error {
	if runtime.GOOS == "darwin" {
		instance, err := createNativeInstanceDarwin()
		if err != nil {
			return err
		}
		r.instance = instance
		return vk.InitInstance(r.instance)
	}

	extensions := append([]string{}, r.window.RequiredInstanceExtensions()...)
	if runtime.GOOS == "darwin" {
		extensions = appendUnique(extensions, "VK_KHR_portability_enumeration")
		extensions = appendUnique(extensions, "VK_KHR_get_physical_device_properties2")
	}

	layers := []string{}
	if r.validation && hasValidationLayer() {
		layers = append(layers, "VK_LAYER_KHRONOS_validation")
	}

	appInfo := vk.ApplicationInfo{
		SType:              vk.StructureTypeApplicationInfo,
		PApplicationName:   "LumaGo",
		ApplicationVersion: vk.MakeVersion(0, 0, 1),
		PEngineName:        "LumaGo",
		EngineVersion:      vk.MakeVersion(0, 0, 1),
		ApiVersion:         vk.ApiVersion10,
	}
	createInfo := vk.InstanceCreateInfo{
		SType:                   vk.StructureTypeInstanceCreateInfo,
		Flags:                   portabilityEnumerationFlag(),
		PApplicationInfo:        &appInfo,
		EnabledExtensionCount:   uint32(len(extensions)),
		PpEnabledExtensionNames: extensions,
		EnabledLayerCount:       uint32(len(layers)),
		PpEnabledLayerNames:     layers,
	}
	if os.Getenv("LUMAGO_VULKAN_DEBUG") == "1" {
		fmt.Printf("vulkan extensions=%v layers=%v flags=0x%x\n", extensions, layers, createInfo.Flags)
	}
	result := vk.CreateInstance(&createInfo, nil, &r.instance)
	if result == vk.ErrorLayerNotPresent && len(layers) > 0 {
		createInfo.EnabledLayerCount = 0
		createInfo.PpEnabledLayerNames = nil
		result = vk.CreateInstance(&createInfo, nil, &r.instance)
	}
	return check(result, "create Vulkan instance")
}

func (r *Renderer) pickPhysicalDevice() error {
	var count uint32
	if err := check(vk.EnumeratePhysicalDevices(r.instance, &count, nil), "count physical devices"); err != nil {
		return err
	}
	if count == 0 {
		return errors.New("vulkan: no physical devices found")
	}
	devices := make([]vk.PhysicalDevice, count)
	if err := check(vk.EnumeratePhysicalDevices(r.instance, &count, devices), "enumerate physical devices"); err != nil {
		return err
	}
	for _, device := range devices {
		graphicsFamily, presentFamily, ok := r.queueFamilies(device)
		if !ok || !supportsDeviceExtensions(device) {
			continue
		}
		r.physicalDevice = device
		r.graphicsFamily = graphicsFamily
		r.presentFamily = presentFamily
		return nil
	}
	return errors.New("vulkan: no device supports graphics and present queues")
}

func (r *Renderer) queueFamilies(device vk.PhysicalDevice) (uint32, uint32, bool) {
	var count uint32
	vk.GetPhysicalDeviceQueueFamilyProperties(device, &count, nil)
	properties := make([]vk.QueueFamilyProperties, count)
	vk.GetPhysicalDeviceQueueFamilyProperties(device, &count, properties)

	var graphicsFamily, presentFamily uint32
	hasGraphics, hasPresent := false, false
	for i := range properties {
		properties[i].Deref()
		if properties[i].QueueFlags&vk.QueueFlags(vk.QueueGraphicsBit) != 0 {
			graphicsFamily = uint32(i)
			hasGraphics = true
		}
		var present vk.Bool32
		if vk.GetPhysicalDeviceSurfaceSupport(device, uint32(i), r.surface, &present) == vk.Success && present == vk.True {
			presentFamily = uint32(i)
			hasPresent = true
		}
	}
	return graphicsFamily, presentFamily, hasGraphics && hasPresent
}

func (r *Renderer) createLogicalDevice() error {
	if runtime.GOOS == "darwin" {
		device, err := createNativeDeviceDarwin(r.physicalDevice, r.graphicsFamily, r.presentFamily)
		if err != nil {
			return err
		}
		r.device = device
		vk.GetDeviceQueue(r.device, r.graphicsFamily, 0, &r.graphicsQueue)
		vk.GetDeviceQueue(r.device, r.presentFamily, 0, &r.presentQueue)
		return nil
	}

	priority := []float32{1}
	families := []uint32{r.graphicsFamily}
	if r.presentFamily != r.graphicsFamily {
		families = append(families, r.presentFamily)
	}
	queueInfos := make([]vk.DeviceQueueCreateInfo, 0, len(families))
	for _, family := range families {
		queueInfos = append(queueInfos, vk.DeviceQueueCreateInfo{
			SType:            vk.StructureTypeDeviceQueueCreateInfo,
			QueueFamilyIndex: family,
			QueueCount:       1,
			PQueuePriorities: priority,
		})
	}

	extensions := []string{vk.KhrSwapchainExtensionName}
	if runtime.GOOS == "darwin" {
		extensions = append(extensions, "VK_KHR_portability_subset")
	}
	if os.Getenv("LUMAGO_VULKAN_DEBUG") == "1" {
		fmt.Printf("vulkan device extensions requested=%v available=%v\n", extensions, deviceExtensionNames(r.physicalDevice))
	}

	createInfo := vk.DeviceCreateInfo{
		SType:                   vk.StructureTypeDeviceCreateInfo,
		QueueCreateInfoCount:    uint32(len(queueInfos)),
		PQueueCreateInfos:       queueInfos,
		EnabledExtensionCount:   uint32(len(extensions)),
		PpEnabledExtensionNames: extensions,
	}
	if err := check(vk.CreateDevice(r.physicalDevice, &createInfo, nil, &r.device), "create logical device"); err != nil {
		return err
	}
	vk.GetDeviceQueue(r.device, r.graphicsFamily, 0, &r.graphicsQueue)
	vk.GetDeviceQueue(r.device, r.presentFamily, 0, &r.presentQueue)
	return nil
}

func (r *Renderer) createSwapchain() error {
	support, err := r.swapchainSupport()
	if err != nil {
		return err
	}
	if len(support.formats) == 0 || len(support.presentModes) == 0 {
		return errors.New("vulkan: swapchain has no formats or present modes")
	}

	format := chooseSurfaceFormat(support.formats)
	presentMode := choosePresentMode(support.presentModes)
	extent := chooseExtent(support.capabilities, r.window)
	imageCount := support.capabilities.MinImageCount + 1
	if support.capabilities.MaxImageCount > 0 && imageCount > support.capabilities.MaxImageCount {
		imageCount = support.capabilities.MaxImageCount
	}

	createInfo := vk.SwapchainCreateInfo{
		SType:            vk.StructureTypeSwapchainCreateInfo,
		Surface:          r.surface,
		MinImageCount:    imageCount,
		ImageFormat:      format.Format,
		ImageColorSpace:  format.ColorSpace,
		ImageExtent:      extent,
		ImageArrayLayers: 1,
		ImageUsage:       vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit),
		PreTransform:     support.capabilities.CurrentTransform,
		CompositeAlpha:   chooseCompositeAlpha(support.capabilities),
		PresentMode:      presentMode,
		Clipped:          vk.True,
		OldSwapchain:     vk.NullSwapchain,
	}
	if r.graphicsFamily != r.presentFamily {
		createInfo.ImageSharingMode = vk.SharingModeConcurrent
		createInfo.QueueFamilyIndexCount = 2
		createInfo.PQueueFamilyIndices = []uint32{r.graphicsFamily, r.presentFamily}
	} else {
		createInfo.ImageSharingMode = vk.SharingModeExclusive
	}

	if err := check(vk.CreateSwapchain(r.device, &createInfo, nil, &r.swapchain), "create swapchain"); err != nil {
		return err
	}
	r.swapchainFormat = format.Format
	r.swapchainExtent = extent

	var count uint32
	if err := check(vk.GetSwapchainImages(r.device, r.swapchain, &count, nil), "count swapchain images"); err != nil {
		return err
	}
	r.swapchainImages = make([]vk.Image, count)
	if err := check(vk.GetSwapchainImages(r.device, r.swapchain, &count, r.swapchainImages), "get swapchain images"); err != nil {
		return err
	}

	r.imageViews = make([]vk.ImageView, len(r.swapchainImages))
	for i, image := range r.swapchainImages {
		viewInfo := vk.ImageViewCreateInfo{
			SType:    vk.StructureTypeImageViewCreateInfo,
			Image:    image,
			ViewType: vk.ImageViewType2d,
			Format:   r.swapchainFormat,
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				BaseMipLevel:   0,
				LevelCount:     1,
				BaseArrayLayer: 0,
				LayerCount:     1,
			},
		}
		if err := check(vk.CreateImageView(r.device, &viewInfo, nil, &r.imageViews[i]), "create swapchain image view"); err != nil {
			return err
		}
	}
	return nil
}

func (r *Renderer) createRenderPass() error {
	attachment := vk.AttachmentDescription{
		Format:         r.swapchainFormat,
		Samples:        vk.SampleCount1Bit,
		LoadOp:         vk.AttachmentLoadOpClear,
		StoreOp:        vk.AttachmentStoreOpStore,
		StencilLoadOp:  vk.AttachmentLoadOpDontCare,
		StencilStoreOp: vk.AttachmentStoreOpDontCare,
		InitialLayout:  vk.ImageLayoutUndefined,
		FinalLayout:    vk.ImageLayoutPresentSrc,
	}
	colorRef := vk.AttachmentReference{
		Attachment: 0,
		Layout:     vk.ImageLayoutColorAttachmentOptimal,
	}
	subpass := vk.SubpassDescription{
		PipelineBindPoint:    vk.PipelineBindPointGraphics,
		ColorAttachmentCount: 1,
		PColorAttachments:    []vk.AttachmentReference{colorRef},
	}
	dependency := vk.SubpassDependency{
		SrcSubpass:    vk.SubpassExternal,
		DstSubpass:    0,
		SrcStageMask:  vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit),
		DstStageMask:  vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit),
		DstAccessMask: vk.AccessFlags(vk.AccessColorAttachmentWriteBit),
	}
	createInfo := vk.RenderPassCreateInfo{
		SType:           vk.StructureTypeRenderPassCreateInfo,
		AttachmentCount: 1,
		PAttachments:    []vk.AttachmentDescription{attachment},
		SubpassCount:    1,
		PSubpasses:      []vk.SubpassDescription{subpass},
		DependencyCount: 1,
		PDependencies:   []vk.SubpassDependency{dependency},
	}
	return check(vk.CreateRenderPass(r.device, &createInfo, nil, &r.renderPass), "create render pass")
}

func (r *Renderer) createPipeline() error {
	vert, err := r.createShaderModule(filepath.Join(r.shaderDirectory, "quad.vert.spv"))
	if err != nil {
		return err
	}
	defer vk.DestroyShaderModule(r.device, vert, nil)
	frag, err := r.createShaderModule(filepath.Join(r.shaderDirectory, "quad.frag.spv"))
	if err != nil {
		return err
	}
	defer vk.DestroyShaderModule(r.device, frag, nil)

	if runtime.GOOS == "darwin" {
		layout, pipeline, err := createNativeQuadPipelineDarwin(r.device, r.renderPass, r.swapchainExtent, r.descriptorSetLayout, vert, frag)
		if err != nil {
			return err
		}
		r.pipelineLayout = layout
		r.pipeline = pipeline
		return nil
	}

	stages := []vk.PipelineShaderStageCreateInfo{
		{SType: vk.StructureTypePipelineShaderStageCreateInfo, Stage: vk.ShaderStageVertexBit, Module: vert, PName: "main"},
		{SType: vk.StructureTypePipelineShaderStageCreateInfo, Stage: vk.ShaderStageFragmentBit, Module: frag, PName: "main"},
	}
	binding := vk.VertexInputBindingDescription{
		Binding:   0,
		Stride:    quadVertexStride,
		InputRate: vk.VertexInputRateVertex,
	}
	attributes := []vk.VertexInputAttributeDescription{
		{Location: 0, Binding: 0, Format: vk.FormatR32g32Sfloat, Offset: 0},
		{Location: 1, Binding: 0, Format: vk.FormatR32g32Sfloat, Offset: 8},
		{Location: 2, Binding: 0, Format: vk.FormatR32g32b32a32Sfloat, Offset: 16},
	}
	vertexInput := vk.PipelineVertexInputStateCreateInfo{
		SType:                           vk.StructureTypePipelineVertexInputStateCreateInfo,
		VertexBindingDescriptionCount:   1,
		PVertexBindingDescriptions:      []vk.VertexInputBindingDescription{binding},
		VertexAttributeDescriptionCount: uint32(len(attributes)),
		PVertexAttributeDescriptions:    attributes,
	}
	inputAssembly := vk.PipelineInputAssemblyStateCreateInfo{
		SType:    vk.StructureTypePipelineInputAssemblyStateCreateInfo,
		Topology: vk.PrimitiveTopologyTriangleList,
	}
	viewport := vk.Viewport{
		X:        0,
		Y:        0,
		Width:    float32(r.swapchainExtent.Width),
		Height:   float32(r.swapchainExtent.Height),
		MinDepth: 0,
		MaxDepth: 1,
	}
	scissor := vk.Rect2D{Offset: vk.Offset2D{X: 0, Y: 0}, Extent: r.swapchainExtent}
	viewportState := vk.PipelineViewportStateCreateInfo{
		SType:         vk.StructureTypePipelineViewportStateCreateInfo,
		ViewportCount: 1,
		PViewports:    []vk.Viewport{viewport},
		ScissorCount:  1,
		PScissors:     []vk.Rect2D{scissor},
	}
	rasterizer := vk.PipelineRasterizationStateCreateInfo{
		SType:       vk.StructureTypePipelineRasterizationStateCreateInfo,
		PolygonMode: vk.PolygonModeFill,
		CullMode:    vk.CullModeFlags(vk.CullModeNone),
		FrontFace:   vk.FrontFaceClockwise,
		LineWidth:   1,
	}
	multisample := vk.PipelineMultisampleStateCreateInfo{
		SType:                vk.StructureTypePipelineMultisampleStateCreateInfo,
		RasterizationSamples: vk.SampleCount1Bit,
	}
	colorBlendAttachment := vk.PipelineColorBlendAttachmentState{
		ColorWriteMask: vk.ColorComponentFlags(vk.ColorComponentRBit | vk.ColorComponentGBit | vk.ColorComponentBBit | vk.ColorComponentABit),
	}
	colorBlend := vk.PipelineColorBlendStateCreateInfo{
		SType:           vk.StructureTypePipelineColorBlendStateCreateInfo,
		AttachmentCount: 1,
		PAttachments:    []vk.PipelineColorBlendAttachmentState{colorBlendAttachment},
	}
	layoutInfo := vk.PipelineLayoutCreateInfo{
		SType:          vk.StructureTypePipelineLayoutCreateInfo,
		SetLayoutCount: 1,
		PSetLayouts:    []vk.DescriptorSetLayout{r.descriptorSetLayout},
	}
	if err := check(vk.CreatePipelineLayout(r.device, &layoutInfo, nil, &r.pipelineLayout), "create pipeline layout"); err != nil {
		return err
	}

	pipelineInfo := vk.GraphicsPipelineCreateInfo{
		SType:               vk.StructureTypeGraphicsPipelineCreateInfo,
		StageCount:          uint32(len(stages)),
		PStages:             stages,
		PVertexInputState:   &vertexInput,
		PInputAssemblyState: &inputAssembly,
		PViewportState:      &viewportState,
		PRasterizationState: &rasterizer,
		PMultisampleState:   &multisample,
		PColorBlendState:    &colorBlend,
		Layout:              r.pipelineLayout,
		RenderPass:          r.renderPass,
		Subpass:             0,
	}
	pipelines := []vk.Pipeline{vk.NullPipeline}
	if err := check(vk.CreateGraphicsPipelines(r.device, vk.NullPipelineCache, 1, []vk.GraphicsPipelineCreateInfo{pipelineInfo}, nil, pipelines), "create graphics pipeline"); err != nil {
		return err
	}
	r.pipeline = pipelines[0]
	return nil
}

func (r *Renderer) createFramebuffers() error {
	r.framebuffers = make([]vk.Framebuffer, len(r.imageViews))
	for i, view := range r.imageViews {
		createInfo := vk.FramebufferCreateInfo{
			SType:           vk.StructureTypeFramebufferCreateInfo,
			RenderPass:      r.renderPass,
			AttachmentCount: 1,
			PAttachments:    []vk.ImageView{view},
			Width:           r.swapchainExtent.Width,
			Height:          r.swapchainExtent.Height,
			Layers:          1,
		}
		if err := check(vk.CreateFramebuffer(r.device, &createInfo, nil, &r.framebuffers[i]), "create framebuffer"); err != nil {
			return err
		}
	}
	return nil
}

func (r *Renderer) createCommandPool() error {
	createInfo := vk.CommandPoolCreateInfo{
		SType:            vk.StructureTypeCommandPoolCreateInfo,
		Flags:            vk.CommandPoolCreateFlags(vk.CommandPoolCreateResetCommandBufferBit),
		QueueFamilyIndex: r.graphicsFamily,
	}
	return check(vk.CreateCommandPool(r.device, &createInfo, nil, &r.commandPool), "create command pool")
}

func (r *Renderer) createCommandBuffers() error {
	r.commandBuffers = make([]vk.CommandBuffer, framesInFlight)
	allocInfo := vk.CommandBufferAllocateInfo{
		SType:              vk.StructureTypeCommandBufferAllocateInfo,
		CommandPool:        r.commandPool,
		Level:              vk.CommandBufferLevelPrimary,
		CommandBufferCount: uint32(len(r.commandBuffers)),
	}
	return check(vk.AllocateCommandBuffers(r.device, &allocInfo, r.commandBuffers), "allocate command buffers")
}

func (r *Renderer) createSyncObjects() error {
	r.imageAvailable = make([]vk.Semaphore, framesInFlight)
	r.renderFinished = make([]vk.Semaphore, framesInFlight)
	r.inFlight = make([]vk.Fence, framesInFlight)
	semaphoreInfo := vk.SemaphoreCreateInfo{SType: vk.StructureTypeSemaphoreCreateInfo}
	fenceInfo := vk.FenceCreateInfo{SType: vk.StructureTypeFenceCreateInfo, Flags: vk.FenceCreateFlags(vk.FenceCreateSignaledBit)}
	for i := 0; i < framesInFlight; i++ {
		if err := check(vk.CreateSemaphore(r.device, &semaphoreInfo, nil, &r.imageAvailable[i]), "create image semaphore"); err != nil {
			return err
		}
		if err := check(vk.CreateSemaphore(r.device, &semaphoreInfo, nil, &r.renderFinished[i]), "create render semaphore"); err != nil {
			return err
		}
		if err := check(vk.CreateFence(r.device, &fenceInfo, nil, &r.inFlight[i]), "create frame fence"); err != nil {
			return err
		}
	}
	return nil
}

func (r *Renderer) recordCommandBuffer(commandBuffer vk.CommandBuffer, imageIndex uint32) error {
	beginInfo := vk.CommandBufferBeginInfo{SType: vk.StructureTypeCommandBufferBeginInfo}
	if err := check(vk.BeginCommandBuffer(commandBuffer, &beginInfo), "begin command buffer"); err != nil {
		return err
	}
	clear := vk.NewClearValue([]float32{0.03, 0.04, 0.05, 1})
	renderPassInfo := vk.RenderPassBeginInfo{
		SType:           vk.StructureTypeRenderPassBeginInfo,
		RenderPass:      r.renderPass,
		Framebuffer:     r.framebuffers[imageIndex],
		RenderArea:      vk.Rect2D{Offset: vk.Offset2D{X: 0, Y: 0}, Extent: r.swapchainExtent},
		ClearValueCount: 1,
		PClearValues:    []vk.ClearValue{clear},
	}
	vk.CmdBeginRenderPass(commandBuffer, &renderPassInfo, vk.SubpassContentsInline)
	vk.CmdBindPipeline(commandBuffer, vk.PipelineBindPointGraphics, r.pipeline)
	vk.CmdBindDescriptorSets(commandBuffer, vk.PipelineBindPointGraphics, r.pipelineLayout, 0, 1, []vk.DescriptorSet{r.descriptorSet}, 0, nil)
	frame := &r.spriteFrames[r.frame]
	vk.CmdBindVertexBuffers(commandBuffer, 0, 1, []vk.Buffer{frame.vertexBuffer}, []vk.DeviceSize{0})
	vk.CmdBindIndexBuffer(commandBuffer, frame.indexBuffer, 0, vk.IndexTypeUint16)
	if r.pendingBatch.Stats.IndexCount > 0 {
		vk.CmdDrawIndexed(commandBuffer, uint32(r.pendingBatch.Stats.IndexCount), 1, 0, 0, 0)
	}
	vk.CmdEndRenderPass(commandBuffer)
	return check(vk.EndCommandBuffer(commandBuffer), "end command buffer")
}

func (r *Renderer) uploadSpriteBatch(batch graphics.SpriteBatch) error {
	if len(batch.Vertices) == 0 || len(batch.Indices) == 0 {
		return nil
	}

	frame := &r.spriteFrames[r.frame]
	frame.vertexUpload = packSpriteVertices(frame.vertexUpload, batch.Vertices)
	frame.indexUpload = packSpriteIndices(frame.indexUpload, batch.Indices)

	if err := r.ensureHostVertexBuffer(frame, len(frame.vertexUpload)); err != nil {
		return err
	}
	if err := r.ensureHostIndexBuffer(frame, len(frame.indexUpload)); err != nil {
		return err
	}
	if err := r.copyToMemory(frame.vertexMemory, frame.vertexUpload); err != nil {
		return err
	}
	return r.copyToMemory(frame.indexMemory, frame.indexUpload)
}

func (r *Renderer) ensureHostVertexBuffer(frame *spriteFrameResources, size int) error {
	if size <= frame.vertexCapacity {
		return nil
	}
	r.destroyBuffer(&frame.vertexBuffer, &frame.vertexMemory)
	if err := r.createHostBuffer(size, vk.BufferUsageFlags(vk.BufferUsageVertexBufferBit), &frame.vertexBuffer, &frame.vertexMemory); err != nil {
		return err
	}
	frame.vertexCapacity = size
	return nil
}

func (r *Renderer) ensureHostIndexBuffer(frame *spriteFrameResources, size int) error {
	if size <= frame.indexCapacity {
		return nil
	}
	r.destroyBuffer(&frame.indexBuffer, &frame.indexMemory)
	if err := r.createHostBuffer(size, vk.BufferUsageFlags(vk.BufferUsageIndexBufferBit), &frame.indexBuffer, &frame.indexMemory); err != nil {
		return err
	}
	frame.indexCapacity = size
	return nil
}

func (r *Renderer) recreateSwapchain() error {
	width, height := r.window.FramebufferSize()
	if width == 0 || height == 0 {
		r.window.WaitForFramebuffer()
		width, height = r.window.FramebufferSize()
		if width == 0 || height == 0 || r.window.ShouldClose() {
			return nil
		}
	}
	if err := check(vk.DeviceWaitIdle(r.device), "wait for resize idle"); err != nil {
		return err
	}
	r.cleanupSwapchain()
	if err := r.createSwapchain(); err != nil {
		return err
	}
	r.lightingTargets = defaultLightingRenderTargets(r.swapchainExtent, r.swapchainFormat)
	if err := r.createLightingRenderBuffers(); err != nil {
		return err
	}
	if err := r.createRenderPass(); err != nil {
		return err
	}
	if err := r.createPipeline(); err != nil {
		return err
	}
	return r.createFramebuffers()
}

func (r *Renderer) cleanupSwapchain() {
	if r.device == vk.NullDevice {
		return
	}
	for _, framebuffer := range r.framebuffers {
		if framebuffer != vk.NullFramebuffer {
			vk.DestroyFramebuffer(r.device, framebuffer, nil)
		}
	}
	r.framebuffers = nil
	if r.pipeline != vk.NullPipeline {
		vk.DestroyPipeline(r.device, r.pipeline, nil)
		r.pipeline = vk.NullPipeline
	}
	if r.pipelineLayout != vk.NullPipelineLayout {
		vk.DestroyPipelineLayout(r.device, r.pipelineLayout, nil)
		r.pipelineLayout = vk.NullPipelineLayout
	}
	if r.renderPass != vk.NullRenderPass {
		vk.DestroyRenderPass(r.device, r.renderPass, nil)
		r.renderPass = vk.NullRenderPass
	}
	for _, view := range r.imageViews {
		if view != vk.NullImageView {
			vk.DestroyImageView(r.device, view, nil)
		}
	}
	r.imageViews = nil
	r.cleanupLightingRenderBuffers()
	if r.swapchain != vk.NullSwapchain {
		vk.DestroySwapchain(r.device, r.swapchain, nil)
		r.swapchain = vk.NullSwapchain
	}
	r.lightingTargets = lightingRenderTargets{}
}

func (r *Renderer) createShaderModule(path string) (vk.ShaderModule, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return vk.NullShaderModule, err
	}
	if len(data)%4 != 0 {
		return vk.NullShaderModule, fmt.Errorf("vulkan: shader %s is not 32-bit aligned", path)
	}
	code := make([]uint32, len(data)/4)
	for i := range code {
		code[i] = binary.LittleEndian.Uint32(data[i*4:])
	}
	createInfo := vk.ShaderModuleCreateInfo{
		SType:    vk.StructureTypeShaderModuleCreateInfo,
		CodeSize: uint(len(data)),
		PCode:    code,
	}
	var module vk.ShaderModule
	return module, check(vk.CreateShaderModule(r.device, &createInfo, nil, &module), "create shader module")
}

type swapchainSupport struct {
	capabilities vk.SurfaceCapabilities
	formats      []vk.SurfaceFormat
	presentModes []vk.PresentMode
}

func (r *Renderer) swapchainSupport() (swapchainSupport, error) {
	var support swapchainSupport
	if err := check(vk.GetPhysicalDeviceSurfaceCapabilities(r.physicalDevice, r.surface, &support.capabilities), "get surface capabilities"); err != nil {
		return support, err
	}
	support.capabilities.Deref()

	var formatCount uint32
	if err := check(vk.GetPhysicalDeviceSurfaceFormats(r.physicalDevice, r.surface, &formatCount, nil), "count surface formats"); err != nil {
		return support, err
	}
	if formatCount > 0 {
		support.formats = make([]vk.SurfaceFormat, formatCount)
		if err := check(vk.GetPhysicalDeviceSurfaceFormats(r.physicalDevice, r.surface, &formatCount, support.formats), "get surface formats"); err != nil {
			return support, err
		}
		for i := range support.formats {
			support.formats[i].Deref()
		}
	}

	var modeCount uint32
	if err := check(vk.GetPhysicalDeviceSurfacePresentModes(r.physicalDevice, r.surface, &modeCount, nil), "count present modes"); err != nil {
		return support, err
	}
	if modeCount > 0 {
		support.presentModes = make([]vk.PresentMode, modeCount)
		if err := check(vk.GetPhysicalDeviceSurfacePresentModes(r.physicalDevice, r.surface, &modeCount, support.presentModes), "get present modes"); err != nil {
			return support, err
		}
	}
	return support, nil
}

func chooseSurfaceFormat(formats []vk.SurfaceFormat) vk.SurfaceFormat {
	for _, format := range formats {
		if format.Format == vk.FormatB8g8r8a8Srgb && format.ColorSpace == vk.ColorSpaceSrgbNonlinear {
			return format
		}
	}
	return formats[0]
}

func choosePresentMode(modes []vk.PresentMode) vk.PresentMode {
	for _, mode := range modes {
		if mode == vk.PresentModeMailbox {
			return mode
		}
	}
	return vk.PresentModeFifo
}

func chooseExtent(capabilities vk.SurfaceCapabilities, window *desktop.Window) vk.Extent2D {
	if capabilities.CurrentExtent.Width != vk.MaxUint32 {
		return capabilities.CurrentExtent
	}
	width, height := window.FramebufferSize()
	extent := vk.Extent2D{Width: uint32(width), Height: uint32(height)}
	if extent.Width < capabilities.MinImageExtent.Width {
		extent.Width = capabilities.MinImageExtent.Width
	}
	if extent.Width > capabilities.MaxImageExtent.Width {
		extent.Width = capabilities.MaxImageExtent.Width
	}
	if extent.Height < capabilities.MinImageExtent.Height {
		extent.Height = capabilities.MinImageExtent.Height
	}
	if extent.Height > capabilities.MaxImageExtent.Height {
		extent.Height = capabilities.MaxImageExtent.Height
	}
	return extent
}

func chooseCompositeAlpha(capabilities vk.SurfaceCapabilities) vk.CompositeAlphaFlagBits {
	if capabilities.SupportedCompositeAlpha&vk.CompositeAlphaFlags(vk.CompositeAlphaOpaqueBit) != 0 {
		return vk.CompositeAlphaOpaqueBit
	}
	if capabilities.SupportedCompositeAlpha&vk.CompositeAlphaFlags(vk.CompositeAlphaPreMultipliedBit) != 0 {
		return vk.CompositeAlphaPreMultipliedBit
	}
	if capabilities.SupportedCompositeAlpha&vk.CompositeAlphaFlags(vk.CompositeAlphaPostMultipliedBit) != 0 {
		return vk.CompositeAlphaPostMultipliedBit
	}
	return vk.CompositeAlphaInheritBit
}

func supportsDeviceExtensions(device vk.PhysicalDevice) bool {
	return supportsExtension(device, vk.KhrSwapchainExtensionName)
}

func supportsExtension(device vk.PhysicalDevice, name string) bool {
	var count uint32
	if vk.EnumerateDeviceExtensionProperties(device, "", &count, nil) != vk.Success {
		return false
	}
	properties := make([]vk.ExtensionProperties, count)
	if vk.EnumerateDeviceExtensionProperties(device, "", &count, properties) != vk.Success {
		return false
	}
	for i := range properties {
		properties[i].Deref()
		if vk.ToString(properties[i].ExtensionName[:]) == name {
			return true
		}
	}
	return false
}

func deviceExtensionNames(device vk.PhysicalDevice) []string {
	var count uint32
	if vk.EnumerateDeviceExtensionProperties(device, "", &count, nil) != vk.Success {
		return nil
	}
	properties := make([]vk.ExtensionProperties, count)
	if vk.EnumerateDeviceExtensionProperties(device, "", &count, properties) != vk.Success {
		return nil
	}
	names := make([]string, 0, len(properties))
	for i := range properties {
		properties[i].Deref()
		names = append(names, vk.ToString(properties[i].ExtensionName[:]))
	}
	return names
}

func hasValidationLayer() bool {
	var count uint32
	if vk.EnumerateInstanceLayerProperties(&count, nil) != vk.Success {
		return false
	}
	properties := make([]vk.LayerProperties, count)
	if vk.EnumerateInstanceLayerProperties(&count, properties) != vk.Success {
		return false
	}
	for i := range properties {
		properties[i].Deref()
		if vk.ToString(properties[i].LayerName[:]) == "VK_LAYER_KHRONOS_validation" {
			return true
		}
	}
	return false
}

func appendUnique(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func portabilityEnumerationFlag() vk.InstanceCreateFlags {
	if runtime.GOOS == "darwin" {
		return vk.InstanceCreateFlags(0x00000001)
	}
	return 0
}

func check(result vk.Result, action string) error {
	if result == vk.Success {
		return nil
	}
	return fmt.Errorf("%s: %w", action, vk.Error(result))
}
