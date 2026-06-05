//go:build !darwin

package vulkan

import vk "github.com/sarattha/lumago/engine/renderer/vulkan/internal/vk"

func createNativeInstanceDarwin() (vk.Instance, error) {
	return vk.NullInstance, nil
}

func createNativeDeviceDarwin(physicalDevice vk.PhysicalDevice, graphicsFamily, presentFamily uint32) (vk.Device, error) {
	return vk.NullDevice, nil
}

func createNativeQuadPipelineDarwin(device vk.Device, renderPass vk.RenderPass, extent vk.Extent2D, descriptorSetLayout vk.DescriptorSetLayout, vert, frag vk.ShaderModule) (vk.PipelineLayout, vk.Pipeline, error) {
	return vk.NullPipelineLayout, vk.NullPipeline, nil
}
