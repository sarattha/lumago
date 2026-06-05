//go:build !darwin

package vulkan

import vk "github.com/vulkan-go/vulkan"

func createNativeInstanceDarwin() (vk.Instance, error) {
	return nil, nil
}

func createNativeDeviceDarwin(physicalDevice vk.PhysicalDevice, graphicsFamily, presentFamily uint32) (vk.Device, error) {
	return nil, nil
}

func createNativeQuadPipelineDarwin(device vk.Device, renderPass vk.RenderPass, extent vk.Extent2D, descriptorSetLayout vk.DescriptorSetLayout, vert, frag vk.ShaderModule) (vk.PipelineLayout, vk.Pipeline, error) {
	return vk.NullPipelineLayout, vk.NullPipeline, nil
}
