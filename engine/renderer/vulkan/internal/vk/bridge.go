package vk

/*
#cgo pkg-config: vulkan
#cgo darwin LDFLAGS: -L/opt/homebrew/lib
#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <vulkan/vulkan.h>

static inline void lumagoSetClearColor(VkClearValue* value, float r, float g, float b, float a) {
	value->color.float32[0] = r;
	value->color.float32[1] = g;
	value->color.float32[2] = b;
	value->color.float32[3] = a;
}
*/
import "C"

import (
	"errors"
	"unsafe"
)

func cHandle[T ~unsafe.Pointer](h T) unsafe.Pointer {
	return unsafe.Pointer(h)
}

func goHandle[T ~unsafe.Pointer](h unsafe.Pointer) T {
	return T(h)
}

func result(r C.VkResult) Result {
	return Result(r)
}

func SetDefaultGetInstanceProcAddr() error { return nil }

func SetGetInstanceProcAddr(_ unsafe.Pointer) {}

func Init() error { return nil }

func InitInstance(_ Instance) error { return nil }

func cStringArray(values []string) (**C.char, func()) {
	if len(values) == 0 {
		return nil, func() {}
	}
	ptrs := C.malloc(C.size_t(len(values)) * C.size_t(unsafe.Sizeof(uintptr(0))))
	array := unsafe.Slice((**C.char)(ptrs), len(values))
	for i, value := range values {
		array[i] = C.CString(value)
	}
	return (**C.char)(ptrs), func() {
		for _, value := range array {
			C.free(unsafe.Pointer(value))
		}
		C.free(ptrs)
	}
}

func CreateInstance(info *InstanceCreateInfo, _ *AllocationCallbacks, out *Instance) Result {
	appName := C.CString(info.PApplicationInfo.PApplicationName)
	engineName := C.CString(info.PApplicationInfo.PEngineName)
	defer C.free(unsafe.Pointer(appName))
	defer C.free(unsafe.Pointer(engineName))
	extensions, freeExtensions := cStringArray(info.PpEnabledExtensionNames)
	defer freeExtensions()
	layers, freeLayers := cStringArray(info.PpEnabledLayerNames)
	defer freeLayers()

	app := C.VkApplicationInfo{
		sType:              C.VK_STRUCTURE_TYPE_APPLICATION_INFO,
		pApplicationName:   appName,
		applicationVersion: C.uint32_t(info.PApplicationInfo.ApplicationVersion),
		pEngineName:        engineName,
		engineVersion:      C.uint32_t(info.PApplicationInfo.EngineVersion),
		apiVersion:         C.uint32_t(info.PApplicationInfo.ApiVersion),
	}
	create := C.VkInstanceCreateInfo{
		sType:                   C.VK_STRUCTURE_TYPE_INSTANCE_CREATE_INFO,
		flags:                   C.VkInstanceCreateFlags(info.Flags),
		pApplicationInfo:        &app,
		enabledExtensionCount:   C.uint32_t(len(info.PpEnabledExtensionNames)),
		ppEnabledExtensionNames: extensions,
		enabledLayerCount:       C.uint32_t(len(info.PpEnabledLayerNames)),
		ppEnabledLayerNames:     layers,
	}
	var instance C.VkInstance
	r := C.vkCreateInstance(&create, nil, &instance)
	*out = goHandle[Instance](unsafe.Pointer(instance))
	return result(r)
}

func DestroyInstance(instance Instance, _ *AllocationCallbacks) {
	C.vkDestroyInstance(C.VkInstance(cHandle(instance)), nil)
}

func DestroySurface(instance Instance, surface Surface, _ *AllocationCallbacks) {
	C.vkDestroySurfaceKHR(C.VkInstance(cHandle(instance)), C.VkSurfaceKHR(cHandle(surface)), nil)
}

func EnumerateInstanceLayerProperties(count *uint32, properties []LayerProperties) Result {
	cCount := C.uint32_t(*count)
	if properties == nil {
		r := C.vkEnumerateInstanceLayerProperties(&cCount, nil)
		*count = uint32(cCount)
		return result(r)
	}
	cProperties := make([]C.VkLayerProperties, len(properties))
	r := C.vkEnumerateInstanceLayerProperties(&cCount, &cProperties[0])
	*count = uint32(cCount)
	for i := 0; i < int(cCount) && i < len(properties); i++ {
		copy(properties[i].LayerName[:], C.GoBytes(unsafe.Pointer(&cProperties[i].layerName[0]), 256))
	}
	return result(r)
}

func EnumeratePhysicalDevices(instance Instance, count *uint32, devices []PhysicalDevice) Result {
	cCount := C.uint32_t(*count)
	if devices == nil {
		r := C.vkEnumeratePhysicalDevices(C.VkInstance(cHandle(instance)), &cCount, nil)
		*count = uint32(cCount)
		return result(r)
	}
	cDevices := make([]C.VkPhysicalDevice, len(devices))
	r := C.vkEnumeratePhysicalDevices(C.VkInstance(cHandle(instance)), &cCount, &cDevices[0])
	*count = uint32(cCount)
	for i := 0; i < int(cCount) && i < len(devices); i++ {
		devices[i] = goHandle[PhysicalDevice](unsafe.Pointer(cDevices[i]))
	}
	return result(r)
}

func GetPhysicalDeviceQueueFamilyProperties(device PhysicalDevice, count *uint32, properties []QueueFamilyProperties) {
	cCount := C.uint32_t(*count)
	if properties == nil {
		C.vkGetPhysicalDeviceQueueFamilyProperties(C.VkPhysicalDevice(cHandle(device)), &cCount, nil)
		*count = uint32(cCount)
		return
	}
	cProperties := make([]C.VkQueueFamilyProperties, len(properties))
	C.vkGetPhysicalDeviceQueueFamilyProperties(C.VkPhysicalDevice(cHandle(device)), &cCount, &cProperties[0])
	*count = uint32(cCount)
	for i := 0; i < int(cCount) && i < len(properties); i++ {
		properties[i].QueueFlags = QueueFlags(cProperties[i].queueFlags)
		properties[i].QueueCount = uint32(cProperties[i].queueCount)
	}
}

func GetPhysicalDeviceSurfaceSupport(device PhysicalDevice, family uint32, surface Surface, supported *Bool32) Result {
	var cSupported C.VkBool32
	r := C.vkGetPhysicalDeviceSurfaceSupportKHR(C.VkPhysicalDevice(cHandle(device)), C.uint32_t(family), C.VkSurfaceKHR(cHandle(surface)), &cSupported)
	*supported = Bool32(cSupported)
	return result(r)
}

func EnumerateDeviceExtensionProperties(device PhysicalDevice, layerName string, count *uint32, properties []ExtensionProperties) Result {
	var cLayer *C.char
	if layerName != "" {
		cLayer = C.CString(layerName)
		defer C.free(unsafe.Pointer(cLayer))
	}
	cCount := C.uint32_t(*count)
	if properties == nil {
		r := C.vkEnumerateDeviceExtensionProperties(C.VkPhysicalDevice(cHandle(device)), cLayer, &cCount, nil)
		*count = uint32(cCount)
		return result(r)
	}
	cProperties := make([]C.VkExtensionProperties, len(properties))
	r := C.vkEnumerateDeviceExtensionProperties(C.VkPhysicalDevice(cHandle(device)), cLayer, &cCount, &cProperties[0])
	*count = uint32(cCount)
	for i := 0; i < int(cCount) && i < len(properties); i++ {
		copy(properties[i].ExtensionName[:], C.GoBytes(unsafe.Pointer(&cProperties[i].extensionName[0]), 256))
		properties[i].SpecVersion = uint32(cProperties[i].specVersion)
	}
	return result(r)
}

func CreateDevice(physicalDevice PhysicalDevice, info *DeviceCreateInfo, _ *AllocationCallbacks, out *Device) Result {
	extensions, freeExtensions := cStringArray(info.PpEnabledExtensionNames)
	defer freeExtensions()

	queueInfos := make([]C.VkDeviceQueueCreateInfo, len(info.PQueueCreateInfos))
	priorities := make([][]C.float, len(info.PQueueCreateInfos))
	for i, queue := range info.PQueueCreateInfos {
		priorities[i] = make([]C.float, len(queue.PQueuePriorities))
		for j, priority := range queue.PQueuePriorities {
			priorities[i][j] = C.float(priority)
		}
		queueInfos[i] = C.VkDeviceQueueCreateInfo{
			sType:            C.VK_STRUCTURE_TYPE_DEVICE_QUEUE_CREATE_INFO,
			flags:            C.VkDeviceQueueCreateFlags(queue.Flags),
			queueFamilyIndex: C.uint32_t(queue.QueueFamilyIndex),
			queueCount:       C.uint32_t(queue.QueueCount),
			pQueuePriorities: &priorities[i][0],
		}
	}
	create := C.VkDeviceCreateInfo{
		sType:                   C.VK_STRUCTURE_TYPE_DEVICE_CREATE_INFO,
		flags:                   C.VkDeviceCreateFlags(info.Flags),
		queueCreateInfoCount:    C.uint32_t(len(queueInfos)),
		pQueueCreateInfos:       &queueInfos[0],
		enabledExtensionCount:   C.uint32_t(len(info.PpEnabledExtensionNames)),
		ppEnabledExtensionNames: extensions,
	}
	var device C.VkDevice
	r := C.vkCreateDevice(C.VkPhysicalDevice(cHandle(physicalDevice)), &create, nil, &device)
	*out = goHandle[Device](unsafe.Pointer(device))
	return result(r)
}

func DestroyDevice(device Device, _ *AllocationCallbacks) {
	C.vkDestroyDevice(C.VkDevice(cHandle(device)), nil)
}

func GetDeviceQueue(device Device, family, index uint32, out *Queue) {
	var queue C.VkQueue
	C.vkGetDeviceQueue(C.VkDevice(cHandle(device)), C.uint32_t(family), C.uint32_t(index), &queue)
	*out = goHandle[Queue](unsafe.Pointer(queue))
}

func DeviceWaitIdle(device Device) Result {
	return result(C.vkDeviceWaitIdle(C.VkDevice(cHandle(device))))
}

func GetPhysicalDeviceSurfaceCapabilities(device PhysicalDevice, surface Surface, out *SurfaceCapabilities) Result {
	var caps C.VkSurfaceCapabilitiesKHR
	r := C.vkGetPhysicalDeviceSurfaceCapabilitiesKHR(C.VkPhysicalDevice(cHandle(device)), C.VkSurfaceKHR(cHandle(surface)), &caps)
	out.MinImageCount = uint32(caps.minImageCount)
	out.MaxImageCount = uint32(caps.maxImageCount)
	out.CurrentExtent = Extent2D{Width: uint32(caps.currentExtent.width), Height: uint32(caps.currentExtent.height)}
	out.MinImageExtent = Extent2D{Width: uint32(caps.minImageExtent.width), Height: uint32(caps.minImageExtent.height)}
	out.MaxImageExtent = Extent2D{Width: uint32(caps.maxImageExtent.width), Height: uint32(caps.maxImageExtent.height)}
	out.CurrentTransform = SurfaceTransformFlagBits(caps.currentTransform)
	out.SupportedCompositeAlpha = CompositeAlphaFlags(caps.supportedCompositeAlpha)
	return result(r)
}

func GetPhysicalDeviceSurfaceFormats(device PhysicalDevice, surface Surface, count *uint32, formats []SurfaceFormat) Result {
	cCount := C.uint32_t(*count)
	if formats == nil {
		r := C.vkGetPhysicalDeviceSurfaceFormatsKHR(C.VkPhysicalDevice(cHandle(device)), C.VkSurfaceKHR(cHandle(surface)), &cCount, nil)
		*count = uint32(cCount)
		return result(r)
	}
	cFormats := make([]C.VkSurfaceFormatKHR, len(formats))
	r := C.vkGetPhysicalDeviceSurfaceFormatsKHR(C.VkPhysicalDevice(cHandle(device)), C.VkSurfaceKHR(cHandle(surface)), &cCount, &cFormats[0])
	*count = uint32(cCount)
	for i := 0; i < int(cCount) && i < len(formats); i++ {
		formats[i] = SurfaceFormat{Format: Format(cFormats[i].format), ColorSpace: ColorSpace(cFormats[i].colorSpace)}
	}
	return result(r)
}

func GetPhysicalDeviceSurfacePresentModes(device PhysicalDevice, surface Surface, count *uint32, modes []PresentMode) Result {
	cCount := C.uint32_t(*count)
	if modes == nil {
		r := C.vkGetPhysicalDeviceSurfacePresentModesKHR(C.VkPhysicalDevice(cHandle(device)), C.VkSurfaceKHR(cHandle(surface)), &cCount, nil)
		*count = uint32(cCount)
		return result(r)
	}
	cModes := make([]C.VkPresentModeKHR, len(modes))
	r := C.vkGetPhysicalDeviceSurfacePresentModesKHR(C.VkPhysicalDevice(cHandle(device)), C.VkSurfaceKHR(cHandle(surface)), &cCount, &cModes[0])
	*count = uint32(cCount)
	for i := 0; i < int(cCount) && i < len(modes); i++ {
		modes[i] = PresentMode(cModes[i])
	}
	return result(r)
}

func CreateSwapchain(device Device, info *SwapchainCreateInfo, _ *AllocationCallbacks, out *Swapchain) Result {
	var families *C.uint32_t
	if len(info.PQueueFamilyIndices) > 0 {
		families = (*C.uint32_t)(unsafe.Pointer(&info.PQueueFamilyIndices[0]))
	}
	create := C.VkSwapchainCreateInfoKHR{
		sType:                 C.VK_STRUCTURE_TYPE_SWAPCHAIN_CREATE_INFO_KHR,
		surface:               C.VkSurfaceKHR(cHandle(info.Surface)),
		minImageCount:         C.uint32_t(info.MinImageCount),
		imageFormat:           C.VkFormat(info.ImageFormat),
		imageColorSpace:       C.VkColorSpaceKHR(info.ImageColorSpace),
		imageExtent:           C.VkExtent2D{width: C.uint32_t(info.ImageExtent.Width), height: C.uint32_t(info.ImageExtent.Height)},
		imageArrayLayers:      C.uint32_t(info.ImageArrayLayers),
		imageUsage:            C.VkImageUsageFlags(info.ImageUsage),
		imageSharingMode:      C.VkSharingMode(info.ImageSharingMode),
		queueFamilyIndexCount: C.uint32_t(len(info.PQueueFamilyIndices)),
		pQueueFamilyIndices:   families,
		preTransform:          C.VkSurfaceTransformFlagBitsKHR(info.PreTransform),
		compositeAlpha:        C.VkCompositeAlphaFlagBitsKHR(info.CompositeAlpha),
		presentMode:           C.VkPresentModeKHR(info.PresentMode),
		clipped:               C.VkBool32(info.Clipped),
		oldSwapchain:          C.VkSwapchainKHR(cHandle(info.OldSwapchain)),
	}
	var swapchain C.VkSwapchainKHR
	r := C.vkCreateSwapchainKHR(C.VkDevice(cHandle(device)), &create, nil, &swapchain)
	*out = goHandle[Swapchain](unsafe.Pointer(swapchain))
	return result(r)
}

func DestroySwapchain(device Device, swapchain Swapchain, _ *AllocationCallbacks) {
	C.vkDestroySwapchainKHR(C.VkDevice(cHandle(device)), C.VkSwapchainKHR(cHandle(swapchain)), nil)
}

func GetSwapchainImages(device Device, swapchain Swapchain, count *uint32, images []Image) Result {
	cCount := C.uint32_t(*count)
	if images == nil {
		r := C.vkGetSwapchainImagesKHR(C.VkDevice(cHandle(device)), C.VkSwapchainKHR(cHandle(swapchain)), &cCount, nil)
		*count = uint32(cCount)
		return result(r)
	}
	cImages := make([]C.VkImage, len(images))
	r := C.vkGetSwapchainImagesKHR(C.VkDevice(cHandle(device)), C.VkSwapchainKHR(cHandle(swapchain)), &cCount, &cImages[0])
	*count = uint32(cCount)
	for i := 0; i < int(cCount) && i < len(images); i++ {
		images[i] = goHandle[Image](unsafe.Pointer(cImages[i]))
	}
	return result(r)
}

func Present(queue Queue, info *PresentInfo) Result {
	if len(info.PSwapchains) == 0 || len(info.PImageIndices) == 0 {
		return ErrorInitializationFailed
	}
	swapchains := make([]C.VkSwapchainKHR, len(info.PSwapchains))
	for i, swapchain := range info.PSwapchains {
		swapchains[i] = C.VkSwapchainKHR(cHandle(swapchain))
	}
	var waitSemaphores []C.VkSemaphore
	var waitPtr *C.VkSemaphore
	if len(info.PWaitSemaphores) > 0 {
		waitSemaphores = make([]C.VkSemaphore, len(info.PWaitSemaphores))
		for i, semaphore := range info.PWaitSemaphores {
			waitSemaphores[i] = C.VkSemaphore(cHandle(semaphore))
		}
		waitPtr = &waitSemaphores[0]
	}
	present := C.VkPresentInfoKHR{
		sType:              C.VK_STRUCTURE_TYPE_PRESENT_INFO_KHR,
		waitSemaphoreCount: C.uint32_t(len(waitSemaphores)),
		pWaitSemaphores:    waitPtr,
		swapchainCount:     C.uint32_t(len(swapchains)),
		pSwapchains:        &swapchains[0],
		pImageIndices:      (*C.uint32_t)(unsafe.Pointer(&info.PImageIndices[0])),
	}
	return result(C.vkQueuePresentKHR(C.VkQueue(cHandle(queue)), &present))
}

func QueuePresent(queue Queue, info *PresentInfo) Result {
	return Present(queue, info)
}

func CreateImageView(device Device, info *ImageViewCreateInfo, _ *AllocationCallbacks, out *ImageView) Result {
	create := C.VkImageViewCreateInfo{
		sType:    C.VK_STRUCTURE_TYPE_IMAGE_VIEW_CREATE_INFO,
		image:    C.VkImage(cHandle(info.Image)),
		viewType: C.VkImageViewType(info.ViewType),
		format:   C.VkFormat(info.Format),
		subresourceRange: C.VkImageSubresourceRange{
			aspectMask:     C.VkImageAspectFlags(info.SubresourceRange.AspectMask),
			baseMipLevel:   C.uint32_t(info.SubresourceRange.BaseMipLevel),
			levelCount:     C.uint32_t(info.SubresourceRange.LevelCount),
			baseArrayLayer: C.uint32_t(info.SubresourceRange.BaseArrayLayer),
			layerCount:     C.uint32_t(info.SubresourceRange.LayerCount),
		},
	}
	var view C.VkImageView
	r := C.vkCreateImageView(C.VkDevice(cHandle(device)), &create, nil, &view)
	*out = goHandle[ImageView](unsafe.Pointer(view))
	return result(r)
}

func DestroyImageView(device Device, view ImageView, _ *AllocationCallbacks) {
	C.vkDestroyImageView(C.VkDevice(cHandle(device)), C.VkImageView(cHandle(view)), nil)
}

func CreateRenderPass(device Device, info *RenderPassCreateInfo, _ *AllocationCallbacks, out *RenderPass) Result {
	attachments := make([]C.VkAttachmentDescription, len(info.PAttachments))
	for i, attachment := range info.PAttachments {
		attachments[i] = C.VkAttachmentDescription{
			format:         C.VkFormat(attachment.Format),
			samples:        C.VkSampleCountFlagBits(attachment.Samples),
			loadOp:         C.VkAttachmentLoadOp(attachment.LoadOp),
			storeOp:        C.VkAttachmentStoreOp(attachment.StoreOp),
			stencilLoadOp:  C.VkAttachmentLoadOp(attachment.StencilLoadOp),
			stencilStoreOp: C.VkAttachmentStoreOp(attachment.StencilStoreOp),
			initialLayout:  C.VkImageLayout(attachment.InitialLayout),
			finalLayout:    C.VkImageLayout(attachment.FinalLayout),
		}
	}
	colorRefs := make([][]C.VkAttachmentReference, len(info.PSubpasses))
	subpasses := make([]C.VkSubpassDescription, len(info.PSubpasses))
	for i, subpass := range info.PSubpasses {
		colorRefs[i] = make([]C.VkAttachmentReference, len(subpass.PColorAttachments))
		for j, ref := range subpass.PColorAttachments {
			colorRefs[i][j] = C.VkAttachmentReference{attachment: C.uint32_t(ref.Attachment), layout: C.VkImageLayout(ref.Layout)}
		}
		subpasses[i] = C.VkSubpassDescription{
			pipelineBindPoint:    C.VkPipelineBindPoint(subpass.PipelineBindPoint),
			colorAttachmentCount: C.uint32_t(len(colorRefs[i])),
			pColorAttachments:    &colorRefs[i][0],
		}
	}
	dependencies := make([]C.VkSubpassDependency, len(info.PDependencies))
	for i, dependency := range info.PDependencies {
		dependencies[i] = C.VkSubpassDependency{
			srcSubpass:    C.uint32_t(dependency.SrcSubpass),
			dstSubpass:    C.uint32_t(dependency.DstSubpass),
			srcStageMask:  C.VkPipelineStageFlags(dependency.SrcStageMask),
			dstStageMask:  C.VkPipelineStageFlags(dependency.DstStageMask),
			srcAccessMask: C.VkAccessFlags(dependency.SrcAccessMask),
			dstAccessMask: C.VkAccessFlags(dependency.DstAccessMask),
		}
	}
	create := C.VkRenderPassCreateInfo{
		sType:           C.VK_STRUCTURE_TYPE_RENDER_PASS_CREATE_INFO,
		attachmentCount: C.uint32_t(len(attachments)),
		pAttachments:    &attachments[0],
		subpassCount:    C.uint32_t(len(subpasses)),
		pSubpasses:      &subpasses[0],
		dependencyCount: C.uint32_t(len(dependencies)),
		pDependencies:   &dependencies[0],
	}
	var renderPass C.VkRenderPass
	r := C.vkCreateRenderPass(C.VkDevice(cHandle(device)), &create, nil, &renderPass)
	*out = goHandle[RenderPass](unsafe.Pointer(renderPass))
	return result(r)
}

func DestroyRenderPass(device Device, renderPass RenderPass, _ *AllocationCallbacks) {
	C.vkDestroyRenderPass(C.VkDevice(cHandle(device)), C.VkRenderPass(cHandle(renderPass)), nil)
}

func CreateShaderModule(device Device, info *ShaderModuleCreateInfo, _ *AllocationCallbacks, out *ShaderModule) Result {
	if len(info.PCode) == 0 {
		return ErrorInitializationFailed
	}
	create := C.VkShaderModuleCreateInfo{
		sType:    C.VK_STRUCTURE_TYPE_SHADER_MODULE_CREATE_INFO,
		codeSize: C.size_t(info.CodeSize),
		pCode:    (*C.uint32_t)(unsafe.Pointer(&info.PCode[0])),
	}
	var shader C.VkShaderModule
	r := C.vkCreateShaderModule(C.VkDevice(cHandle(device)), &create, nil, &shader)
	*out = goHandle[ShaderModule](unsafe.Pointer(shader))
	return result(r)
}

func DestroyShaderModule(device Device, shader ShaderModule, _ *AllocationCallbacks) {
	C.vkDestroyShaderModule(C.VkDevice(cHandle(device)), C.VkShaderModule(cHandle(shader)), nil)
}

func CreatePipelineLayout(device Device, info *PipelineLayoutCreateInfo, _ *AllocationCallbacks, out *PipelineLayout) Result {
	var layouts []C.VkDescriptorSetLayout
	var layoutPtr *C.VkDescriptorSetLayout
	if len(info.PSetLayouts) > 0 {
		layouts = make([]C.VkDescriptorSetLayout, len(info.PSetLayouts))
		for i, layout := range info.PSetLayouts {
			layouts[i] = C.VkDescriptorSetLayout(cHandle(layout))
		}
		layoutPtr = &layouts[0]
	}
	create := C.VkPipelineLayoutCreateInfo{
		sType:          C.VK_STRUCTURE_TYPE_PIPELINE_LAYOUT_CREATE_INFO,
		setLayoutCount: C.uint32_t(len(layouts)),
		pSetLayouts:    layoutPtr,
	}
	var layout C.VkPipelineLayout
	r := C.vkCreatePipelineLayout(C.VkDevice(cHandle(device)), &create, nil, &layout)
	*out = goHandle[PipelineLayout](unsafe.Pointer(layout))
	return result(r)
}

func DestroyPipelineLayout(device Device, layout PipelineLayout, _ *AllocationCallbacks) {
	C.vkDestroyPipelineLayout(C.VkDevice(cHandle(device)), C.VkPipelineLayout(cHandle(layout)), nil)
}

func CreateGraphicsPipelines(device Device, _ PipelineCache, _ uint32, infos []GraphicsPipelineCreateInfo, _ *AllocationCallbacks, out []Pipeline) Result {
	if len(infos) != 1 || len(out) == 0 {
		return ErrorInitializationFailed
	}
	info := infos[0]
	stageNames := make([]*C.char, len(info.PStages))
	defer func() {
		for _, name := range stageNames {
			C.free(unsafe.Pointer(name))
		}
	}()
	stages := make([]C.VkPipelineShaderStageCreateInfo, len(info.PStages))
	for i, stage := range info.PStages {
		stageNames[i] = C.CString(stage.PName)
		stages[i] = C.VkPipelineShaderStageCreateInfo{
			sType:  C.VK_STRUCTURE_TYPE_PIPELINE_SHADER_STAGE_CREATE_INFO,
			stage:  C.VkShaderStageFlagBits(stage.Stage),
			module: C.VkShaderModule(cHandle(stage.Module)),
			pName:  stageNames[i],
		}
	}
	bindings := make([]C.VkVertexInputBindingDescription, len(info.PVertexInputState.PVertexBindingDescriptions))
	for i, binding := range info.PVertexInputState.PVertexBindingDescriptions {
		bindings[i] = C.VkVertexInputBindingDescription{
			binding:   C.uint32_t(binding.Binding),
			stride:    C.uint32_t(binding.Stride),
			inputRate: C.VkVertexInputRate(binding.InputRate),
		}
	}
	attributes := make([]C.VkVertexInputAttributeDescription, len(info.PVertexInputState.PVertexAttributeDescriptions))
	for i, attribute := range info.PVertexInputState.PVertexAttributeDescriptions {
		attributes[i] = C.VkVertexInputAttributeDescription{
			location: C.uint32_t(attribute.Location),
			binding:  C.uint32_t(attribute.Binding),
			format:   C.VkFormat(attribute.Format),
			offset:   C.uint32_t(attribute.Offset),
		}
	}
	vertexInput := C.VkPipelineVertexInputStateCreateInfo{
		sType:                           C.VK_STRUCTURE_TYPE_PIPELINE_VERTEX_INPUT_STATE_CREATE_INFO,
		vertexBindingDescriptionCount:   C.uint32_t(len(bindings)),
		pVertexBindingDescriptions:      &bindings[0],
		vertexAttributeDescriptionCount: C.uint32_t(len(attributes)),
		pVertexAttributeDescriptions:    &attributes[0],
	}
	inputAssembly := C.VkPipelineInputAssemblyStateCreateInfo{
		sType:    C.VK_STRUCTURE_TYPE_PIPELINE_INPUT_ASSEMBLY_STATE_CREATE_INFO,
		topology: C.VkPrimitiveTopology(info.PInputAssemblyState.Topology),
	}
	viewports := make([]C.VkViewport, len(info.PViewportState.PViewports))
	for i, viewport := range info.PViewportState.PViewports {
		viewports[i] = C.VkViewport{
			x:        C.float(viewport.X),
			y:        C.float(viewport.Y),
			width:    C.float(viewport.Width),
			height:   C.float(viewport.Height),
			minDepth: C.float(viewport.MinDepth),
			maxDepth: C.float(viewport.MaxDepth),
		}
	}
	scissors := make([]C.VkRect2D, len(info.PViewportState.PScissors))
	for i, scissor := range info.PViewportState.PScissors {
		scissors[i] = C.VkRect2D{
			offset: C.VkOffset2D{x: C.int32_t(scissor.Offset.X), y: C.int32_t(scissor.Offset.Y)},
			extent: C.VkExtent2D{width: C.uint32_t(scissor.Extent.Width), height: C.uint32_t(scissor.Extent.Height)},
		}
	}
	viewportState := C.VkPipelineViewportStateCreateInfo{
		sType:         C.VK_STRUCTURE_TYPE_PIPELINE_VIEWPORT_STATE_CREATE_INFO,
		viewportCount: C.uint32_t(len(viewports)),
		pViewports:    &viewports[0],
		scissorCount:  C.uint32_t(len(scissors)),
		pScissors:     &scissors[0],
	}
	rasterizer := C.VkPipelineRasterizationStateCreateInfo{
		sType:       C.VK_STRUCTURE_TYPE_PIPELINE_RASTERIZATION_STATE_CREATE_INFO,
		polygonMode: C.VkPolygonMode(info.PRasterizationState.PolygonMode),
		cullMode:    C.VkCullModeFlags(info.PRasterizationState.CullMode),
		frontFace:   C.VkFrontFace(info.PRasterizationState.FrontFace),
		lineWidth:   C.float(info.PRasterizationState.LineWidth),
	}
	multisample := C.VkPipelineMultisampleStateCreateInfo{
		sType:                C.VK_STRUCTURE_TYPE_PIPELINE_MULTISAMPLE_STATE_CREATE_INFO,
		rasterizationSamples: C.VkSampleCountFlagBits(info.PMultisampleState.RasterizationSamples),
	}
	attachments := make([]C.VkPipelineColorBlendAttachmentState, len(info.PColorBlendState.PAttachments))
	for i, attachment := range info.PColorBlendState.PAttachments {
		attachments[i] = C.VkPipelineColorBlendAttachmentState{colorWriteMask: C.VkColorComponentFlags(attachment.ColorWriteMask)}
	}
	colorBlend := C.VkPipelineColorBlendStateCreateInfo{
		sType:           C.VK_STRUCTURE_TYPE_PIPELINE_COLOR_BLEND_STATE_CREATE_INFO,
		attachmentCount: C.uint32_t(len(attachments)),
		pAttachments:    &attachments[0],
	}
	create := C.VkGraphicsPipelineCreateInfo{
		sType:               C.VK_STRUCTURE_TYPE_GRAPHICS_PIPELINE_CREATE_INFO,
		stageCount:          C.uint32_t(len(stages)),
		pStages:             &stages[0],
		pVertexInputState:   &vertexInput,
		pInputAssemblyState: &inputAssembly,
		pViewportState:      &viewportState,
		pRasterizationState: &rasterizer,
		pMultisampleState:   &multisample,
		pColorBlendState:    &colorBlend,
		layout:              C.VkPipelineLayout(cHandle(info.Layout)),
		renderPass:          C.VkRenderPass(cHandle(info.RenderPass)),
		subpass:             C.uint32_t(info.Subpass),
	}
	var pipeline C.VkPipeline
	r := C.vkCreateGraphicsPipelines(C.VkDevice(cHandle(device)), nil, 1, &create, nil, &pipeline)
	out[0] = goHandle[Pipeline](unsafe.Pointer(pipeline))
	return result(r)
}

func DestroyPipeline(device Device, pipeline Pipeline, _ *AllocationCallbacks) {
	C.vkDestroyPipeline(C.VkDevice(cHandle(device)), C.VkPipeline(cHandle(pipeline)), nil)
}

func CreateFramebuffer(device Device, info *FramebufferCreateInfo, _ *AllocationCallbacks, out *Framebuffer) Result {
	attachments := make([]C.VkImageView, len(info.PAttachments))
	for i, attachment := range info.PAttachments {
		attachments[i] = C.VkImageView(cHandle(attachment))
	}
	create := C.VkFramebufferCreateInfo{
		sType:           C.VK_STRUCTURE_TYPE_FRAMEBUFFER_CREATE_INFO,
		renderPass:      C.VkRenderPass(cHandle(info.RenderPass)),
		attachmentCount: C.uint32_t(len(attachments)),
		pAttachments:    &attachments[0],
		width:           C.uint32_t(info.Width),
		height:          C.uint32_t(info.Height),
		layers:          C.uint32_t(info.Layers),
	}
	var framebuffer C.VkFramebuffer
	r := C.vkCreateFramebuffer(C.VkDevice(cHandle(device)), &create, nil, &framebuffer)
	*out = goHandle[Framebuffer](unsafe.Pointer(framebuffer))
	return result(r)
}

func DestroyFramebuffer(device Device, framebuffer Framebuffer, _ *AllocationCallbacks) {
	C.vkDestroyFramebuffer(C.VkDevice(cHandle(device)), C.VkFramebuffer(cHandle(framebuffer)), nil)
}

func CreateCommandPool(device Device, info *CommandPoolCreateInfo, _ *AllocationCallbacks, out *CommandPool) Result {
	create := C.VkCommandPoolCreateInfo{
		sType:            C.VK_STRUCTURE_TYPE_COMMAND_POOL_CREATE_INFO,
		flags:            C.VkCommandPoolCreateFlags(info.Flags),
		queueFamilyIndex: C.uint32_t(info.QueueFamilyIndex),
	}
	var pool C.VkCommandPool
	r := C.vkCreateCommandPool(C.VkDevice(cHandle(device)), &create, nil, &pool)
	*out = goHandle[CommandPool](unsafe.Pointer(pool))
	return result(r)
}

func DestroyCommandPool(device Device, pool CommandPool, _ *AllocationCallbacks) {
	C.vkDestroyCommandPool(C.VkDevice(cHandle(device)), C.VkCommandPool(cHandle(pool)), nil)
}

func AllocateCommandBuffers(device Device, info *CommandBufferAllocateInfo, out []CommandBuffer) Result {
	cBuffers := make([]C.VkCommandBuffer, len(out))
	alloc := C.VkCommandBufferAllocateInfo{
		sType:              C.VK_STRUCTURE_TYPE_COMMAND_BUFFER_ALLOCATE_INFO,
		commandPool:        C.VkCommandPool(cHandle(info.CommandPool)),
		level:              C.VkCommandBufferLevel(info.Level),
		commandBufferCount: C.uint32_t(len(out)),
	}
	r := C.vkAllocateCommandBuffers(C.VkDevice(cHandle(device)), &alloc, &cBuffers[0])
	for i, buffer := range cBuffers {
		out[i] = goHandle[CommandBuffer](unsafe.Pointer(buffer))
	}
	return result(r)
}

func FreeCommandBuffers(device Device, pool CommandPool, count uint32, buffers []CommandBuffer) {
	if len(buffers) == 0 {
		return
	}
	cBuffers := make([]C.VkCommandBuffer, len(buffers))
	for i, buffer := range buffers {
		cBuffers[i] = C.VkCommandBuffer(cHandle(buffer))
	}
	C.vkFreeCommandBuffers(C.VkDevice(cHandle(device)), C.VkCommandPool(cHandle(pool)), C.uint32_t(count), &cBuffers[0])
}

func CreateSemaphore(device Device, info *SemaphoreCreateInfo, _ *AllocationCallbacks, out *Semaphore) Result {
	create := C.VkSemaphoreCreateInfo{sType: C.VK_STRUCTURE_TYPE_SEMAPHORE_CREATE_INFO}
	var semaphore C.VkSemaphore
	r := C.vkCreateSemaphore(C.VkDevice(cHandle(device)), &create, nil, &semaphore)
	*out = goHandle[Semaphore](unsafe.Pointer(semaphore))
	return result(r)
}

func DestroySemaphore(device Device, semaphore Semaphore, _ *AllocationCallbacks) {
	C.vkDestroySemaphore(C.VkDevice(cHandle(device)), C.VkSemaphore(cHandle(semaphore)), nil)
}

func CreateFence(device Device, info *FenceCreateInfo, _ *AllocationCallbacks, out *Fence) Result {
	create := C.VkFenceCreateInfo{
		sType: C.VK_STRUCTURE_TYPE_FENCE_CREATE_INFO,
		flags: C.VkFenceCreateFlags(info.Flags),
	}
	var fence C.VkFence
	r := C.vkCreateFence(C.VkDevice(cHandle(device)), &create, nil, &fence)
	*out = goHandle[Fence](unsafe.Pointer(fence))
	return result(r)
}

func DestroyFence(device Device, fence Fence, _ *AllocationCallbacks) {
	C.vkDestroyFence(C.VkDevice(cHandle(device)), C.VkFence(cHandle(fence)), nil)
}

func WaitForFences(device Device, count uint32, fences []Fence, waitAll Bool32, timeout uint64) Result {
	cFences := make([]C.VkFence, len(fences))
	for i, fence := range fences {
		cFences[i] = C.VkFence(cHandle(fence))
	}
	return result(C.vkWaitForFences(C.VkDevice(cHandle(device)), C.uint32_t(count), &cFences[0], C.VkBool32(waitAll), C.uint64_t(timeout)))
}

func ResetFences(device Device, count uint32, fences []Fence) Result {
	cFences := make([]C.VkFence, len(fences))
	for i, fence := range fences {
		cFences[i] = C.VkFence(cHandle(fence))
	}
	return result(C.vkResetFences(C.VkDevice(cHandle(device)), C.uint32_t(count), &cFences[0]))
}

func AcquireNextImage(device Device, swapchain Swapchain, timeout uint64, semaphore Semaphore, fence Fence, imageIndex *uint32) Result {
	var cIndex C.uint32_t
	r := C.vkAcquireNextImageKHR(C.VkDevice(cHandle(device)), C.VkSwapchainKHR(cHandle(swapchain)), C.uint64_t(timeout), C.VkSemaphore(cHandle(semaphore)), C.VkFence(cHandle(fence)), &cIndex)
	*imageIndex = uint32(cIndex)
	return result(r)
}

func ResetCommandBuffer(buffer CommandBuffer, flags uint32) Result {
	return result(C.vkResetCommandBuffer(C.VkCommandBuffer(cHandle(buffer)), C.VkCommandBufferResetFlags(flags)))
}

func BeginCommandBuffer(buffer CommandBuffer, info *CommandBufferBeginInfo) Result {
	begin := C.VkCommandBufferBeginInfo{
		sType: C.VK_STRUCTURE_TYPE_COMMAND_BUFFER_BEGIN_INFO,
		flags: C.VkCommandBufferUsageFlags(info.Flags),
	}
	return result(C.vkBeginCommandBuffer(C.VkCommandBuffer(cHandle(buffer)), &begin))
}

func EndCommandBuffer(buffer CommandBuffer) Result {
	return result(C.vkEndCommandBuffer(C.VkCommandBuffer(cHandle(buffer))))
}

func CmdBeginRenderPass(buffer CommandBuffer, info *RenderPassBeginInfo, contents SubpassContents) {
	var clearValues []C.VkClearValue
	var clearPtr *C.VkClearValue
	if len(info.PClearValues) > 0 {
		clearValues = make([]C.VkClearValue, len(info.PClearValues))
		for i, value := range info.PClearValues {
			color := value.color()
			C.lumagoSetClearColor(&clearValues[i], C.float(color[0]), C.float(color[1]), C.float(color[2]), C.float(color[3]))
		}
		clearPtr = &clearValues[0]
	}
	begin := C.VkRenderPassBeginInfo{
		sType:       C.VK_STRUCTURE_TYPE_RENDER_PASS_BEGIN_INFO,
		renderPass:  C.VkRenderPass(cHandle(info.RenderPass)),
		framebuffer: C.VkFramebuffer(cHandle(info.Framebuffer)),
		renderArea: C.VkRect2D{
			offset: C.VkOffset2D{x: C.int32_t(info.RenderArea.Offset.X), y: C.int32_t(info.RenderArea.Offset.Y)},
			extent: C.VkExtent2D{width: C.uint32_t(info.RenderArea.Extent.Width), height: C.uint32_t(info.RenderArea.Extent.Height)},
		},
		clearValueCount: C.uint32_t(len(clearValues)),
		pClearValues:    clearPtr,
	}
	C.vkCmdBeginRenderPass(C.VkCommandBuffer(cHandle(buffer)), &begin, C.VkSubpassContents(contents))
}

func CmdEndRenderPass(buffer CommandBuffer) {
	C.vkCmdEndRenderPass(C.VkCommandBuffer(cHandle(buffer)))
}

func CmdBindPipeline(buffer CommandBuffer, bindPoint PipelineBindPoint, pipeline Pipeline) {
	C.vkCmdBindPipeline(C.VkCommandBuffer(cHandle(buffer)), C.VkPipelineBindPoint(bindPoint), C.VkPipeline(cHandle(pipeline)))
}

func CmdBindDescriptorSets(buffer CommandBuffer, bindPoint PipelineBindPoint, layout PipelineLayout, firstSet, setCount uint32, sets []DescriptorSet, dynamicOffsetCount uint32, dynamicOffsets []uint32) {
	cSets := make([]C.VkDescriptorSet, len(sets))
	for i, set := range sets {
		cSets[i] = C.VkDescriptorSet(cHandle(set))
	}
	C.vkCmdBindDescriptorSets(C.VkCommandBuffer(cHandle(buffer)), C.VkPipelineBindPoint(bindPoint), C.VkPipelineLayout(cHandle(layout)), C.uint32_t(firstSet), C.uint32_t(setCount), &cSets[0], C.uint32_t(dynamicOffsetCount), nil)
}

func CmdBindVertexBuffers(buffer CommandBuffer, firstBinding, bindingCount uint32, buffers []Buffer, offsets []DeviceSize) {
	cBuffers := make([]C.VkBuffer, len(buffers))
	for i, buffer := range buffers {
		cBuffers[i] = C.VkBuffer(cHandle(buffer))
	}
	C.vkCmdBindVertexBuffers(C.VkCommandBuffer(cHandle(buffer)), C.uint32_t(firstBinding), C.uint32_t(bindingCount), &cBuffers[0], (*C.VkDeviceSize)(unsafe.Pointer(&offsets[0])))
}

func CmdBindIndexBuffer(commandBuffer CommandBuffer, buffer Buffer, offset DeviceSize, indexType IndexType) {
	C.vkCmdBindIndexBuffer(C.VkCommandBuffer(cHandle(commandBuffer)), C.VkBuffer(cHandle(buffer)), C.VkDeviceSize(offset), C.VkIndexType(indexType))
}

func CmdDrawIndexed(buffer CommandBuffer, indexCount, instanceCount, firstIndex uint32, vertexOffset int32, firstInstance uint32) {
	C.vkCmdDrawIndexed(C.VkCommandBuffer(cHandle(buffer)), C.uint32_t(indexCount), C.uint32_t(instanceCount), C.uint32_t(firstIndex), C.int32_t(vertexOffset), C.uint32_t(firstInstance))
}

func QueueSubmit(queue Queue, count uint32, submits []SubmitInfo, fence Fence) Result {
	if len(submits) != 1 {
		return ErrorInitializationFailed
	}
	submit := submits[0]
	var waitSemaphores []C.VkSemaphore
	var waitPtr *C.VkSemaphore
	if len(submit.PWaitSemaphores) > 0 {
		waitSemaphores = make([]C.VkSemaphore, len(submit.PWaitSemaphores))
		for i, semaphore := range submit.PWaitSemaphores {
			waitSemaphores[i] = C.VkSemaphore(cHandle(semaphore))
		}
		waitPtr = &waitSemaphores[0]
	}
	var signalSemaphores []C.VkSemaphore
	var signalPtr *C.VkSemaphore
	if len(submit.PSignalSemaphores) > 0 {
		signalSemaphores = make([]C.VkSemaphore, len(submit.PSignalSemaphores))
		for i, semaphore := range submit.PSignalSemaphores {
			signalSemaphores[i] = C.VkSemaphore(cHandle(semaphore))
		}
		signalPtr = &signalSemaphores[0]
	}
	commandBuffers := make([]C.VkCommandBuffer, len(submit.PCommandBuffers))
	for i, buffer := range submit.PCommandBuffers {
		commandBuffers[i] = C.VkCommandBuffer(cHandle(buffer))
	}
	var waitStages *C.VkPipelineStageFlags
	if len(submit.PWaitDstStageMask) > 0 {
		waitStages = (*C.VkPipelineStageFlags)(unsafe.Pointer(&submit.PWaitDstStageMask[0]))
	}
	cSubmit := C.VkSubmitInfo{
		sType:                C.VK_STRUCTURE_TYPE_SUBMIT_INFO,
		waitSemaphoreCount:   C.uint32_t(len(waitSemaphores)),
		pWaitSemaphores:      waitPtr,
		pWaitDstStageMask:    waitStages,
		commandBufferCount:   C.uint32_t(len(commandBuffers)),
		pCommandBuffers:      &commandBuffers[0],
		signalSemaphoreCount: C.uint32_t(len(signalSemaphores)),
		pSignalSemaphores:    signalPtr,
	}
	return result(C.vkQueueSubmit(C.VkQueue(cHandle(queue)), C.uint32_t(count), &cSubmit, C.VkFence(cHandle(fence))))
}

func QueueWaitIdle(queue Queue) Result {
	return result(C.vkQueueWaitIdle(C.VkQueue(cHandle(queue))))
}

func CreateBuffer(device Device, info *BufferCreateInfo, _ *AllocationCallbacks, out *Buffer) Result {
	create := C.VkBufferCreateInfo{
		sType:       C.VK_STRUCTURE_TYPE_BUFFER_CREATE_INFO,
		size:        C.VkDeviceSize(info.Size),
		usage:       C.VkBufferUsageFlags(info.Usage),
		sharingMode: C.VkSharingMode(info.SharingMode),
	}
	var buffer C.VkBuffer
	r := C.vkCreateBuffer(C.VkDevice(cHandle(device)), &create, nil, &buffer)
	*out = goHandle[Buffer](unsafe.Pointer(buffer))
	return result(r)
}

func DestroyBuffer(device Device, buffer Buffer, _ *AllocationCallbacks) {
	C.vkDestroyBuffer(C.VkDevice(cHandle(device)), C.VkBuffer(cHandle(buffer)), nil)
}

func GetBufferMemoryRequirements(device Device, buffer Buffer, out *MemoryRequirements) {
	var req C.VkMemoryRequirements
	C.vkGetBufferMemoryRequirements(C.VkDevice(cHandle(device)), C.VkBuffer(cHandle(buffer)), &req)
	out.Size = DeviceSize(req.size)
	out.Alignment = DeviceSize(req.alignment)
	out.MemoryTypeBits = uint32(req.memoryTypeBits)
}

func AllocateMemory(device Device, info *MemoryAllocateInfo, _ *AllocationCallbacks, out *DeviceMemory) Result {
	alloc := C.VkMemoryAllocateInfo{
		sType:           C.VK_STRUCTURE_TYPE_MEMORY_ALLOCATE_INFO,
		allocationSize:  C.VkDeviceSize(info.AllocationSize),
		memoryTypeIndex: C.uint32_t(info.MemoryTypeIndex),
	}
	var memory C.VkDeviceMemory
	r := C.vkAllocateMemory(C.VkDevice(cHandle(device)), &alloc, nil, &memory)
	*out = goHandle[DeviceMemory](unsafe.Pointer(memory))
	return result(r)
}

func FreeMemory(device Device, memory DeviceMemory, _ *AllocationCallbacks) {
	C.vkFreeMemory(C.VkDevice(cHandle(device)), C.VkDeviceMemory(cHandle(memory)), nil)
}

func BindBufferMemory(device Device, buffer Buffer, memory DeviceMemory, offset DeviceSize) Result {
	return result(C.vkBindBufferMemory(C.VkDevice(cHandle(device)), C.VkBuffer(cHandle(buffer)), C.VkDeviceMemory(cHandle(memory)), C.VkDeviceSize(offset)))
}

func MapMemory(device Device, memory DeviceMemory, offset, size DeviceSize, flags uint32, data *unsafe.Pointer) Result {
	return result(C.vkMapMemory(C.VkDevice(cHandle(device)), C.VkDeviceMemory(cHandle(memory)), C.VkDeviceSize(offset), C.VkDeviceSize(size), C.VkMemoryMapFlags(flags), data))
}

func UnmapMemory(device Device, memory DeviceMemory) {
	C.vkUnmapMemory(C.VkDevice(cHandle(device)), C.VkDeviceMemory(cHandle(memory)))
}

func GetPhysicalDeviceMemoryProperties(device PhysicalDevice, out *PhysicalDeviceMemoryProperties) {
	var props C.VkPhysicalDeviceMemoryProperties
	C.vkGetPhysicalDeviceMemoryProperties(C.VkPhysicalDevice(cHandle(device)), &props)
	out.MemoryTypeCount = uint32(props.memoryTypeCount)
	for i := 0; i < int(props.memoryTypeCount) && i < len(out.MemoryTypes); i++ {
		out.MemoryTypes[i] = MemoryType{
			PropertyFlags: MemoryPropertyFlags(props.memoryTypes[i].propertyFlags),
			HeapIndex:     uint32(props.memoryTypes[i].heapIndex),
		}
	}
}

func CreateImage(device Device, info *ImageCreateInfo, _ *AllocationCallbacks, out *Image) Result {
	create := C.VkImageCreateInfo{
		sType:         C.VK_STRUCTURE_TYPE_IMAGE_CREATE_INFO,
		imageType:     C.VkImageType(info.ImageType),
		format:        C.VkFormat(info.Format),
		extent:        C.VkExtent3D{width: C.uint32_t(info.Extent.Width), height: C.uint32_t(info.Extent.Height), depth: C.uint32_t(info.Extent.Depth)},
		mipLevels:     C.uint32_t(info.MipLevels),
		arrayLayers:   C.uint32_t(info.ArrayLayers),
		samples:       C.VkSampleCountFlagBits(info.Samples),
		tiling:        C.VkImageTiling(info.Tiling),
		usage:         C.VkImageUsageFlags(info.Usage),
		sharingMode:   C.VkSharingMode(info.SharingMode),
		initialLayout: C.VkImageLayout(info.InitialLayout),
	}
	var image C.VkImage
	r := C.vkCreateImage(C.VkDevice(cHandle(device)), &create, nil, &image)
	*out = goHandle[Image](unsafe.Pointer(image))
	return result(r)
}

func DestroyImage(device Device, image Image, _ *AllocationCallbacks) {
	C.vkDestroyImage(C.VkDevice(cHandle(device)), C.VkImage(cHandle(image)), nil)
}

func GetImageMemoryRequirements(device Device, image Image, out *MemoryRequirements) {
	var req C.VkMemoryRequirements
	C.vkGetImageMemoryRequirements(C.VkDevice(cHandle(device)), C.VkImage(cHandle(image)), &req)
	out.Size = DeviceSize(req.size)
	out.Alignment = DeviceSize(req.alignment)
	out.MemoryTypeBits = uint32(req.memoryTypeBits)
}

func BindImageMemory(device Device, image Image, memory DeviceMemory, offset DeviceSize) Result {
	return result(C.vkBindImageMemory(C.VkDevice(cHandle(device)), C.VkImage(cHandle(image)), C.VkDeviceMemory(cHandle(memory)), C.VkDeviceSize(offset)))
}

func CreateSampler(device Device, info *SamplerCreateInfo, _ *AllocationCallbacks, out *Sampler) Result {
	create := C.VkSamplerCreateInfo{
		sType:                   C.VK_STRUCTURE_TYPE_SAMPLER_CREATE_INFO,
		magFilter:               C.VkFilter(info.MagFilter),
		minFilter:               C.VkFilter(info.MinFilter),
		mipmapMode:              C.VkSamplerMipmapMode(info.MipmapMode),
		addressModeU:            C.VkSamplerAddressMode(info.AddressModeU),
		addressModeV:            C.VkSamplerAddressMode(info.AddressModeV),
		addressModeW:            C.VkSamplerAddressMode(info.AddressModeW),
		maxLod:                  C.float(info.MaxLod),
		borderColor:             C.VkBorderColor(info.BorderColor),
		unnormalizedCoordinates: C.VkBool32(info.UnnormalizedCoordinates),
	}
	var sampler C.VkSampler
	r := C.vkCreateSampler(C.VkDevice(cHandle(device)), &create, nil, &sampler)
	*out = goHandle[Sampler](unsafe.Pointer(sampler))
	return result(r)
}

func DestroySampler(device Device, sampler Sampler, _ *AllocationCallbacks) {
	C.vkDestroySampler(C.VkDevice(cHandle(device)), C.VkSampler(cHandle(sampler)), nil)
}

func CreateDescriptorSetLayout(device Device, info *DescriptorSetLayoutCreateInfo, _ *AllocationCallbacks, out *DescriptorSetLayout) Result {
	bindings := make([]C.VkDescriptorSetLayoutBinding, len(info.PBindings))
	for i, binding := range info.PBindings {
		bindings[i] = C.VkDescriptorSetLayoutBinding{
			binding:         C.uint32_t(binding.Binding),
			descriptorType:  C.VkDescriptorType(binding.DescriptorType),
			descriptorCount: C.uint32_t(binding.DescriptorCount),
			stageFlags:      C.VkShaderStageFlags(binding.StageFlags),
		}
	}
	create := C.VkDescriptorSetLayoutCreateInfo{
		sType:        C.VK_STRUCTURE_TYPE_DESCRIPTOR_SET_LAYOUT_CREATE_INFO,
		bindingCount: C.uint32_t(len(bindings)),
		pBindings:    &bindings[0],
	}
	var layout C.VkDescriptorSetLayout
	r := C.vkCreateDescriptorSetLayout(C.VkDevice(cHandle(device)), &create, nil, &layout)
	*out = goHandle[DescriptorSetLayout](unsafe.Pointer(layout))
	return result(r)
}

func DestroyDescriptorSetLayout(device Device, layout DescriptorSetLayout, _ *AllocationCallbacks) {
	C.vkDestroyDescriptorSetLayout(C.VkDevice(cHandle(device)), C.VkDescriptorSetLayout(cHandle(layout)), nil)
}

func CreateDescriptorPool(device Device, info *DescriptorPoolCreateInfo, _ *AllocationCallbacks, out *DescriptorPool) Result {
	sizes := make([]C.VkDescriptorPoolSize, len(info.PPoolSizes))
	for i, size := range info.PPoolSizes {
		sizes[i] = C.VkDescriptorPoolSize{_type: C.VkDescriptorType(size.Type), descriptorCount: C.uint32_t(size.DescriptorCount)}
	}
	create := C.VkDescriptorPoolCreateInfo{
		sType:         C.VK_STRUCTURE_TYPE_DESCRIPTOR_POOL_CREATE_INFO,
		maxSets:       C.uint32_t(info.MaxSets),
		poolSizeCount: C.uint32_t(len(sizes)),
		pPoolSizes:    &sizes[0],
	}
	var pool C.VkDescriptorPool
	r := C.vkCreateDescriptorPool(C.VkDevice(cHandle(device)), &create, nil, &pool)
	*out = goHandle[DescriptorPool](unsafe.Pointer(pool))
	return result(r)
}

func DestroyDescriptorPool(device Device, pool DescriptorPool, _ *AllocationCallbacks) {
	C.vkDestroyDescriptorPool(C.VkDevice(cHandle(device)), C.VkDescriptorPool(cHandle(pool)), nil)
}

func AllocateDescriptorSets(device Device, info *DescriptorSetAllocateInfo, out *DescriptorSet) Result {
	layouts := make([]C.VkDescriptorSetLayout, len(info.PSetLayouts))
	for i, layout := range info.PSetLayouts {
		layouts[i] = C.VkDescriptorSetLayout(cHandle(layout))
	}
	alloc := C.VkDescriptorSetAllocateInfo{
		sType:              C.VK_STRUCTURE_TYPE_DESCRIPTOR_SET_ALLOCATE_INFO,
		descriptorPool:     C.VkDescriptorPool(cHandle(info.DescriptorPool)),
		descriptorSetCount: C.uint32_t(info.DescriptorSetCount),
		pSetLayouts:        &layouts[0],
	}
	var set C.VkDescriptorSet
	r := C.vkAllocateDescriptorSets(C.VkDevice(cHandle(device)), &alloc, &set)
	*out = goHandle[DescriptorSet](unsafe.Pointer(set))
	return result(r)
}

func UpdateDescriptorSets(device Device, writeCount uint32, writes []WriteDescriptorSet, _ uint32, _ []CopyDescriptorSet) {
	if len(writes) == 0 {
		return
	}
	imageInfos := make([]C.VkDescriptorImageInfo, len(writes))
	cWrites := make([]C.VkWriteDescriptorSet, len(writes))
	for i, write := range writes {
		if len(write.PImageInfo) > 0 {
			imageInfo := write.PImageInfo[0]
			imageInfos[i] = C.VkDescriptorImageInfo{
				sampler:     C.VkSampler(cHandle(imageInfo.Sampler)),
				imageView:   C.VkImageView(cHandle(imageInfo.ImageView)),
				imageLayout: C.VkImageLayout(imageInfo.ImageLayout),
			}
			cWrites[i].pImageInfo = &imageInfos[i]
		}
		cWrites[i].sType = C.VK_STRUCTURE_TYPE_WRITE_DESCRIPTOR_SET
		cWrites[i].dstSet = C.VkDescriptorSet(cHandle(write.DstSet))
		cWrites[i].dstBinding = C.uint32_t(write.DstBinding)
		cWrites[i].dstArrayElement = C.uint32_t(write.DstArrayElement)
		cWrites[i].descriptorCount = C.uint32_t(write.DescriptorCount)
		cWrites[i].descriptorType = C.VkDescriptorType(write.DescriptorType)
	}
	C.vkUpdateDescriptorSets(C.VkDevice(cHandle(device)), C.uint32_t(writeCount), &cWrites[0], 0, nil)
}

func CmdCopyBuffer(buffer CommandBuffer, src Buffer, dst Buffer, count uint32, regions []BufferCopy) {
	cRegions := make([]C.VkBufferCopy, len(regions))
	for i, region := range regions {
		cRegions[i] = C.VkBufferCopy{srcOffset: C.VkDeviceSize(region.SrcOffset), dstOffset: C.VkDeviceSize(region.DstOffset), size: C.VkDeviceSize(region.Size)}
	}
	C.vkCmdCopyBuffer(C.VkCommandBuffer(cHandle(buffer)), C.VkBuffer(cHandle(src)), C.VkBuffer(cHandle(dst)), C.uint32_t(count), &cRegions[0])
}

func CmdCopyBufferToImage(buffer CommandBuffer, src Buffer, dst Image, layout ImageLayout, count uint32, regions []BufferImageCopy) {
	cRegions := make([]C.VkBufferImageCopy, len(regions))
	for i, region := range regions {
		cRegions[i] = C.VkBufferImageCopy{
			bufferOffset:      C.VkDeviceSize(region.BufferOffset),
			bufferRowLength:   C.uint32_t(region.BufferRowLength),
			bufferImageHeight: C.uint32_t(region.BufferImageHeight),
			imageSubresource: C.VkImageSubresourceLayers{
				aspectMask:     C.VkImageAspectFlags(region.ImageSubresource.AspectMask),
				mipLevel:       C.uint32_t(region.ImageSubresource.MipLevel),
				baseArrayLayer: C.uint32_t(region.ImageSubresource.BaseArrayLayer),
				layerCount:     C.uint32_t(region.ImageSubresource.LayerCount),
			},
			imageOffset: C.VkOffset3D{x: C.int32_t(region.ImageOffset.X), y: C.int32_t(region.ImageOffset.Y), z: C.int32_t(region.ImageOffset.Z)},
			imageExtent: C.VkExtent3D{width: C.uint32_t(region.ImageExtent.Width), height: C.uint32_t(region.ImageExtent.Height), depth: C.uint32_t(region.ImageExtent.Depth)},
		}
	}
	C.vkCmdCopyBufferToImage(C.VkCommandBuffer(cHandle(buffer)), C.VkBuffer(cHandle(src)), C.VkImage(cHandle(dst)), C.VkImageLayout(layout), C.uint32_t(count), &cRegions[0])
}

func CmdPipelineBarrier(buffer CommandBuffer, srcStage, dstStage PipelineStageFlags, dependencyFlags DependencyFlags, _ uint32, _ []MemoryBarrier, _ uint32, _ []BufferMemoryBarrier, imageBarrierCount uint32, imageBarriers []ImageMemoryBarrier) {
	var cImageBarriers []C.VkImageMemoryBarrier
	var cImagePtr *C.VkImageMemoryBarrier
	if len(imageBarriers) > 0 {
		cImageBarriers = make([]C.VkImageMemoryBarrier, len(imageBarriers))
		for i, barrier := range imageBarriers {
			cImageBarriers[i] = C.VkImageMemoryBarrier{
				sType:               C.VK_STRUCTURE_TYPE_IMAGE_MEMORY_BARRIER,
				srcAccessMask:       C.VkAccessFlags(barrier.SrcAccessMask),
				dstAccessMask:       C.VkAccessFlags(barrier.DstAccessMask),
				oldLayout:           C.VkImageLayout(barrier.OldLayout),
				newLayout:           C.VkImageLayout(barrier.NewLayout),
				srcQueueFamilyIndex: C.uint32_t(barrier.SrcQueueFamilyIndex),
				dstQueueFamilyIndex: C.uint32_t(barrier.DstQueueFamilyIndex),
				image:               C.VkImage(cHandle(barrier.Image)),
				subresourceRange: C.VkImageSubresourceRange{
					aspectMask:     C.VkImageAspectFlags(barrier.SubresourceRange.AspectMask),
					baseMipLevel:   C.uint32_t(barrier.SubresourceRange.BaseMipLevel),
					levelCount:     C.uint32_t(barrier.SubresourceRange.LevelCount),
					baseArrayLayer: C.uint32_t(barrier.SubresourceRange.BaseArrayLayer),
					layerCount:     C.uint32_t(barrier.SubresourceRange.LayerCount),
				},
			}
		}
		cImagePtr = &cImageBarriers[0]
	}
	C.vkCmdPipelineBarrier(C.VkCommandBuffer(cHandle(buffer)), C.VkPipelineStageFlags(srcStage), C.VkPipelineStageFlags(dstStage), C.VkDependencyFlags(dependencyFlags), 0, nil, 0, nil, C.uint32_t(imageBarrierCount), cImagePtr)
}

func SurfaceHandle(surface Surface) unsafe.Pointer {
	return cHandle(surface)
}

func SurfaceFromPointer(surface unsafe.Pointer) Surface {
	return goHandle[Surface](surface)
}

func InstanceHandle(instance Instance) unsafe.Pointer {
	return cHandle(instance)
}

func InstanceFromPointer(instance unsafe.Pointer) Instance {
	return goHandle[Instance](instance)
}

func DeviceHandle(device Device) unsafe.Pointer {
	return cHandle(device)
}

func DeviceFromPointer(device unsafe.Pointer) Device {
	return goHandle[Device](device)
}

func PhysicalDeviceHandle(device PhysicalDevice) unsafe.Pointer {
	return cHandle(device)
}

func RenderPassHandle(renderPass RenderPass) unsafe.Pointer {
	return cHandle(renderPass)
}

func PipelineLayoutFromPointer(layout unsafe.Pointer) PipelineLayout {
	return goHandle[PipelineLayout](layout)
}

func PipelineFromPointer(pipeline unsafe.Pointer) Pipeline {
	return goHandle[Pipeline](pipeline)
}

func ShaderModuleHandle(shader ShaderModule) unsafe.Pointer {
	return cHandle(shader)
}

func DescriptorSetLayoutHandle(layout DescriptorSetLayout) unsafe.Pointer {
	return cHandle(layout)
}

func ErrMissingPointer() error {
	return errors.New("vulkan: missing native pointer")
}
