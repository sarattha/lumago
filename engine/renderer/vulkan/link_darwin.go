package vulkan

/*
#cgo darwin pkg-config: vulkan
#cgo darwin LDFLAGS: -L/opt/homebrew/lib
#include <stdlib.h>
#include <vulkan/vulkan.h>

static inline VkResult lumagoCreateInstanceDarwin(VkInstance* instance) {
	const char* extensions[] = {
		"VK_KHR_surface",
		"VK_EXT_metal_surface",
		"VK_KHR_portability_enumeration",
		"VK_KHR_get_physical_device_properties2"
	};

	VkApplicationInfo appInfo = {0};
	appInfo.sType = VK_STRUCTURE_TYPE_APPLICATION_INFO;
	appInfo.pApplicationName = "LumaGo";
	appInfo.applicationVersion = VK_MAKE_VERSION(0, 0, 1);
	appInfo.pEngineName = "LumaGo";
	appInfo.engineVersion = VK_MAKE_VERSION(0, 0, 1);
	appInfo.apiVersion = VK_API_VERSION_1_0;

	VkInstanceCreateInfo createInfo = {0};
	createInfo.sType = VK_STRUCTURE_TYPE_INSTANCE_CREATE_INFO;
	createInfo.flags = VK_INSTANCE_CREATE_ENUMERATE_PORTABILITY_BIT_KHR;
	createInfo.pApplicationInfo = &appInfo;
	createInfo.enabledExtensionCount = 4;
	createInfo.ppEnabledExtensionNames = extensions;

	return vkCreateInstance(&createInfo, NULL, instance);
}

static inline VkResult lumagoCreateDeviceDarwin(VkPhysicalDevice physicalDevice, uint32_t graphicsFamily, uint32_t presentFamily, VkDevice* device) {
	float priority = 1.0f;
	VkDeviceQueueCreateInfo queues[2] = {0};
	uint32_t queueCount = graphicsFamily == presentFamily ? 1 : 2;

	queues[0].sType = VK_STRUCTURE_TYPE_DEVICE_QUEUE_CREATE_INFO;
	queues[0].queueFamilyIndex = graphicsFamily;
	queues[0].queueCount = 1;
	queues[0].pQueuePriorities = &priority;

	if (queueCount == 2) {
		queues[1].sType = VK_STRUCTURE_TYPE_DEVICE_QUEUE_CREATE_INFO;
		queues[1].queueFamilyIndex = presentFamily;
		queues[1].queueCount = 1;
		queues[1].pQueuePriorities = &priority;
	}

	const char* extensions[] = {
		"VK_KHR_swapchain",
		"VK_KHR_portability_subset"
	};

	VkDeviceCreateInfo createInfo = {0};
	createInfo.sType = VK_STRUCTURE_TYPE_DEVICE_CREATE_INFO;
	createInfo.queueCreateInfoCount = queueCount;
	createInfo.pQueueCreateInfos = queues;
	createInfo.enabledExtensionCount = 2;
	createInfo.ppEnabledExtensionNames = extensions;

	return vkCreateDevice(physicalDevice, &createInfo, NULL, device);
}

static inline VkResult lumagoCreateQuadPipelineDarwin(
	VkDevice device,
	VkRenderPass renderPass,
	VkExtent2D extent,
	VkDescriptorSetLayout descriptorSetLayout,
	VkShaderModule vert,
	VkShaderModule frag,
	VkPipelineLayout* layout,
	VkPipeline* pipeline
) {
	VkPipelineShaderStageCreateInfo stages[2] = {0};
	stages[0].sType = VK_STRUCTURE_TYPE_PIPELINE_SHADER_STAGE_CREATE_INFO;
	stages[0].stage = VK_SHADER_STAGE_VERTEX_BIT;
	stages[0].module = vert;
	stages[0].pName = "main";
	stages[1].sType = VK_STRUCTURE_TYPE_PIPELINE_SHADER_STAGE_CREATE_INFO;
	stages[1].stage = VK_SHADER_STAGE_FRAGMENT_BIT;
	stages[1].module = frag;
	stages[1].pName = "main";

	VkPipelineVertexInputStateCreateInfo vertexInput = {0};
	vertexInput.sType = VK_STRUCTURE_TYPE_PIPELINE_VERTEX_INPUT_STATE_CREATE_INFO;
	VkVertexInputBindingDescription binding = {0};
	binding.binding = 0;
	binding.stride = 16;
	binding.inputRate = VK_VERTEX_INPUT_RATE_VERTEX;
	VkVertexInputAttributeDescription attributes[2] = {0};
	attributes[0].location = 0;
	attributes[0].binding = 0;
	attributes[0].format = VK_FORMAT_R32G32_SFLOAT;
	attributes[0].offset = 0;
	attributes[1].location = 1;
	attributes[1].binding = 0;
	attributes[1].format = VK_FORMAT_R32G32_SFLOAT;
	attributes[1].offset = 8;
	vertexInput.vertexBindingDescriptionCount = 1;
	vertexInput.pVertexBindingDescriptions = &binding;
	vertexInput.vertexAttributeDescriptionCount = 2;
	vertexInput.pVertexAttributeDescriptions = attributes;

	VkPipelineInputAssemblyStateCreateInfo inputAssembly = {0};
	inputAssembly.sType = VK_STRUCTURE_TYPE_PIPELINE_INPUT_ASSEMBLY_STATE_CREATE_INFO;
	inputAssembly.topology = VK_PRIMITIVE_TOPOLOGY_TRIANGLE_LIST;

	VkViewport viewport = {0};
	viewport.width = (float) extent.width;
	viewport.height = (float) extent.height;
	viewport.maxDepth = 1.0f;

	VkRect2D scissor = {0};
	scissor.extent = extent;

	VkPipelineViewportStateCreateInfo viewportState = {0};
	viewportState.sType = VK_STRUCTURE_TYPE_PIPELINE_VIEWPORT_STATE_CREATE_INFO;
	viewportState.viewportCount = 1;
	viewportState.pViewports = &viewport;
	viewportState.scissorCount = 1;
	viewportState.pScissors = &scissor;

	VkPipelineRasterizationStateCreateInfo rasterizer = {0};
	rasterizer.sType = VK_STRUCTURE_TYPE_PIPELINE_RASTERIZATION_STATE_CREATE_INFO;
	rasterizer.polygonMode = VK_POLYGON_MODE_FILL;
	rasterizer.cullMode = VK_CULL_MODE_NONE;
	rasterizer.frontFace = VK_FRONT_FACE_CLOCKWISE;
	rasterizer.lineWidth = 1.0f;

	VkPipelineMultisampleStateCreateInfo multisample = {0};
	multisample.sType = VK_STRUCTURE_TYPE_PIPELINE_MULTISAMPLE_STATE_CREATE_INFO;
	multisample.rasterizationSamples = VK_SAMPLE_COUNT_1_BIT;

	VkPipelineColorBlendAttachmentState colorAttachment = {0};
	colorAttachment.colorWriteMask =
		VK_COLOR_COMPONENT_R_BIT |
		VK_COLOR_COMPONENT_G_BIT |
		VK_COLOR_COMPONENT_B_BIT |
		VK_COLOR_COMPONENT_A_BIT;

	VkPipelineColorBlendStateCreateInfo colorBlend = {0};
	colorBlend.sType = VK_STRUCTURE_TYPE_PIPELINE_COLOR_BLEND_STATE_CREATE_INFO;
	colorBlend.attachmentCount = 1;
	colorBlend.pAttachments = &colorAttachment;

	VkPipelineLayoutCreateInfo layoutInfo = {0};
	layoutInfo.sType = VK_STRUCTURE_TYPE_PIPELINE_LAYOUT_CREATE_INFO;
	layoutInfo.setLayoutCount = 1;
	layoutInfo.pSetLayouts = &descriptorSetLayout;
	VkResult result = vkCreatePipelineLayout(device, &layoutInfo, NULL, layout);
	if (result != VK_SUCCESS) {
		return result;
	}

	VkGraphicsPipelineCreateInfo pipelineInfo = {0};
	pipelineInfo.sType = VK_STRUCTURE_TYPE_GRAPHICS_PIPELINE_CREATE_INFO;
	pipelineInfo.stageCount = 2;
	pipelineInfo.pStages = stages;
	pipelineInfo.pVertexInputState = &vertexInput;
	pipelineInfo.pInputAssemblyState = &inputAssembly;
	pipelineInfo.pViewportState = &viewportState;
	pipelineInfo.pRasterizationState = &rasterizer;
	pipelineInfo.pMultisampleState = &multisample;
	pipelineInfo.pColorBlendState = &colorBlend;
	pipelineInfo.layout = *layout;
	pipelineInfo.renderPass = renderPass;

	return vkCreateGraphicsPipelines(device, VK_NULL_HANDLE, 1, &pipelineInfo, NULL, pipeline);
}
*/
import "C"

import (
	"fmt"
	"unsafe"

	vk "github.com/sarattha/lumago/engine/renderer/vulkan/internal/vk"
)

func createNativeInstanceDarwin() (vk.Instance, error) {
	var instance C.VkInstance
	result := C.lumagoCreateInstanceDarwin(&instance)
	if result != C.VK_SUCCESS {
		return vk.NullInstance, fmt.Errorf("create Vulkan instance: native result %d", result)
	}
	return vk.InstanceFromPointer(unsafe.Pointer(instance)), nil
}

func createNativeDeviceDarwin(physicalDevice vk.PhysicalDevice, graphicsFamily, presentFamily uint32) (vk.Device, error) {
	var device C.VkDevice
	result := C.lumagoCreateDeviceDarwin(
		C.VkPhysicalDevice(vk.PhysicalDeviceHandle(physicalDevice)),
		C.uint32_t(graphicsFamily),
		C.uint32_t(presentFamily),
		&device,
	)
	if result != C.VK_SUCCESS {
		return vk.NullDevice, fmt.Errorf("create logical device: native result %d", result)
	}
	return vk.DeviceFromPointer(unsafe.Pointer(device)), nil
}

func createNativeQuadPipelineDarwin(device vk.Device, renderPass vk.RenderPass, extent vk.Extent2D, descriptorSetLayout vk.DescriptorSetLayout, vert, frag vk.ShaderModule) (vk.PipelineLayout, vk.Pipeline, error) {
	var layout C.VkPipelineLayout
	var pipeline C.VkPipeline
	cExtent := C.VkExtent2D{width: C.uint32_t(extent.Width), height: C.uint32_t(extent.Height)}
	result := C.lumagoCreateQuadPipelineDarwin(
		C.VkDevice(vk.DeviceHandle(device)),
		C.VkRenderPass(vk.RenderPassHandle(renderPass)),
		cExtent,
		C.VkDescriptorSetLayout(vk.DescriptorSetLayoutHandle(descriptorSetLayout)),
		C.VkShaderModule(vk.ShaderModuleHandle(vert)),
		C.VkShaderModule(vk.ShaderModuleHandle(frag)),
		&layout,
		&pipeline,
	)
	if result != C.VK_SUCCESS {
		return vk.NullPipelineLayout, vk.NullPipeline, fmt.Errorf("create graphics pipeline: native result %d", result)
	}
	return vk.PipelineLayoutFromPointer(unsafe.Pointer(layout)), vk.PipelineFromPointer(unsafe.Pointer(pipeline)), nil
}
