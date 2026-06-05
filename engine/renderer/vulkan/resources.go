package vulkan

import (
	"encoding/binary"
	"fmt"
	"math"
	"unsafe"

	vk "github.com/vulkan-go/vulkan"
)

const (
	quadVertexStride = 16
	quadIndexCount   = 6
)

func (r *Renderer) createQuadResources() error {
	vertexBytes := quadVertexBytes()
	if err := r.createDeviceBuffer(
		vertexBytes,
		vk.BufferUsageFlags(vk.BufferUsageTransferDstBit|vk.BufferUsageVertexBufferBit),
		&r.vertexBuffer,
		&r.vertexMemory,
	); err != nil {
		return fmt.Errorf("create quad vertex buffer: %w", err)
	}

	indexBytes := quadIndexBytes()
	if err := r.createDeviceBuffer(
		indexBytes,
		vk.BufferUsageFlags(vk.BufferUsageTransferDstBit|vk.BufferUsageIndexBufferBit),
		&r.indexBuffer,
		&r.indexMemory,
	); err != nil {
		return fmt.Errorf("create quad index buffer: %w", err)
	}

	if err := r.createQuadTexture(); err != nil {
		return err
	}
	return r.createQuadDescriptors()
}

func (r *Renderer) cleanupQuadResources() {
	if r.device == nil {
		return
	}
	if r.descriptorPool != vk.NullDescriptorPool {
		vk.DestroyDescriptorPool(r.device, r.descriptorPool, nil)
		r.descriptorPool = vk.NullDescriptorPool
		r.descriptorSet = vk.NullDescriptorSet
	}
	if r.descriptorSetLayout != vk.NullDescriptorSetLayout {
		vk.DestroyDescriptorSetLayout(r.device, r.descriptorSetLayout, nil)
		r.descriptorSetLayout = vk.NullDescriptorSetLayout
	}
	if r.textureSampler != vk.NullSampler {
		vk.DestroySampler(r.device, r.textureSampler, nil)
		r.textureSampler = vk.NullSampler
	}
	if r.textureImageView != vk.NullImageView {
		vk.DestroyImageView(r.device, r.textureImageView, nil)
		r.textureImageView = vk.NullImageView
	}
	if r.textureImage != vk.NullImage {
		vk.DestroyImage(r.device, r.textureImage, nil)
		r.textureImage = vk.NullImage
	}
	if r.textureMemory != vk.NullDeviceMemory {
		vk.FreeMemory(r.device, r.textureMemory, nil)
		r.textureMemory = vk.NullDeviceMemory
	}
	if r.indexBuffer != vk.NullBuffer {
		vk.DestroyBuffer(r.device, r.indexBuffer, nil)
		r.indexBuffer = vk.NullBuffer
	}
	if r.indexMemory != vk.NullDeviceMemory {
		vk.FreeMemory(r.device, r.indexMemory, nil)
		r.indexMemory = vk.NullDeviceMemory
	}
	if r.vertexBuffer != vk.NullBuffer {
		vk.DestroyBuffer(r.device, r.vertexBuffer, nil)
		r.vertexBuffer = vk.NullBuffer
	}
	if r.vertexMemory != vk.NullDeviceMemory {
		vk.FreeMemory(r.device, r.vertexMemory, nil)
		r.vertexMemory = vk.NullDeviceMemory
	}
}

func (r *Renderer) createDeviceBuffer(data []byte, usage vk.BufferUsageFlags, buffer *vk.Buffer, memory *vk.DeviceMemory) error {
	var stagingBuffer vk.Buffer
	var stagingMemory vk.DeviceMemory
	if err := r.createBuffer(
		vk.DeviceSize(len(data)),
		vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit),
		&stagingBuffer,
		&stagingMemory,
	); err != nil {
		return err
	}
	defer vk.DestroyBuffer(r.device, stagingBuffer, nil)
	defer vk.FreeMemory(r.device, stagingMemory, nil)

	if err := r.copyToMemory(stagingMemory, data); err != nil {
		return err
	}
	if err := r.createBuffer(
		vk.DeviceSize(len(data)),
		usage,
		vk.MemoryPropertyFlags(vk.MemoryPropertyDeviceLocalBit),
		buffer,
		memory,
	); err != nil {
		return err
	}
	return r.copyBuffer(stagingBuffer, *buffer, vk.DeviceSize(len(data)))
}

func (r *Renderer) createBuffer(size vk.DeviceSize, usage vk.BufferUsageFlags, properties vk.MemoryPropertyFlags, buffer *vk.Buffer, memory *vk.DeviceMemory) error {
	createInfo := vk.BufferCreateInfo{
		SType:       vk.StructureTypeBufferCreateInfo,
		Size:        size,
		Usage:       usage,
		SharingMode: vk.SharingModeExclusive,
	}
	if err := check(vk.CreateBuffer(r.device, &createInfo, nil, buffer), "create buffer"); err != nil {
		return err
	}

	var requirements vk.MemoryRequirements
	vk.GetBufferMemoryRequirements(r.device, *buffer, &requirements)
	requirements.Deref()

	memoryType, err := r.findMemoryType(requirements.MemoryTypeBits, properties)
	if err != nil {
		vk.DestroyBuffer(r.device, *buffer, nil)
		*buffer = vk.NullBuffer
		return err
	}
	allocInfo := vk.MemoryAllocateInfo{
		SType:           vk.StructureTypeMemoryAllocateInfo,
		AllocationSize:  requirements.Size,
		MemoryTypeIndex: memoryType,
	}
	if err := check(vk.AllocateMemory(r.device, &allocInfo, nil, memory), "allocate buffer memory"); err != nil {
		vk.DestroyBuffer(r.device, *buffer, nil)
		*buffer = vk.NullBuffer
		return err
	}
	if err := check(vk.BindBufferMemory(r.device, *buffer, *memory, 0), "bind buffer memory"); err != nil {
		vk.FreeMemory(r.device, *memory, nil)
		vk.DestroyBuffer(r.device, *buffer, nil)
		*memory = vk.NullDeviceMemory
		*buffer = vk.NullBuffer
		return err
	}
	return nil
}

func (r *Renderer) createQuadTexture() error {
	const width = 2
	const height = 2
	pixels := []byte{
		0xf2, 0xd1, 0x52, 0xff,
		0x29, 0x8f, 0xb3, 0xff,
		0x29, 0x8f, 0xb3, 0xff,
		0xf2, 0xd1, 0x52, 0xff,
	}

	var stagingBuffer vk.Buffer
	var stagingMemory vk.DeviceMemory
	if err := r.createBuffer(
		vk.DeviceSize(len(pixels)),
		vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit),
		&stagingBuffer,
		&stagingMemory,
	); err != nil {
		return fmt.Errorf("create texture staging buffer: %w", err)
	}
	defer vk.DestroyBuffer(r.device, stagingBuffer, nil)
	defer vk.FreeMemory(r.device, stagingMemory, nil)

	if err := r.copyToMemory(stagingMemory, pixels); err != nil {
		return err
	}
	if err := r.createImage(
		width,
		height,
		vk.FormatR8g8b8a8Unorm,
		vk.ImageTilingOptimal,
		vk.ImageUsageFlags(vk.ImageUsageTransferDstBit|vk.ImageUsageSampledBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyDeviceLocalBit),
		&r.textureImage,
		&r.textureMemory,
	); err != nil {
		return fmt.Errorf("create quad texture image: %w", err)
	}
	if err := r.transitionImageLayout(r.textureImage, vk.ImageLayoutUndefined, vk.ImageLayoutTransferDstOptimal); err != nil {
		return err
	}
	if err := r.copyBufferToImage(stagingBuffer, r.textureImage, width, height); err != nil {
		return err
	}
	if err := r.transitionImageLayout(r.textureImage, vk.ImageLayoutTransferDstOptimal, vk.ImageLayoutShaderReadOnlyOptimal); err != nil {
		return err
	}

	viewInfo := vk.ImageViewCreateInfo{
		SType:    vk.StructureTypeImageViewCreateInfo,
		Image:    r.textureImage,
		ViewType: vk.ImageViewType2d,
		Format:   vk.FormatR8g8b8a8Unorm,
		SubresourceRange: vk.ImageSubresourceRange{
			AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
			BaseMipLevel:   0,
			LevelCount:     1,
			BaseArrayLayer: 0,
			LayerCount:     1,
		},
	}
	if err := check(vk.CreateImageView(r.device, &viewInfo, nil, &r.textureImageView), "create texture image view"); err != nil {
		return err
	}

	samplerInfo := vk.SamplerCreateInfo{
		SType:                   vk.StructureTypeSamplerCreateInfo,
		MagFilter:               vk.FilterNearest,
		MinFilter:               vk.FilterNearest,
		MipmapMode:              vk.SamplerMipmapModeNearest,
		AddressModeU:            vk.SamplerAddressModeRepeat,
		AddressModeV:            vk.SamplerAddressModeRepeat,
		AddressModeW:            vk.SamplerAddressModeRepeat,
		MaxLod:                  1,
		BorderColor:             vk.BorderColorIntOpaqueBlack,
		UnnormalizedCoordinates: vk.False,
	}
	return check(vk.CreateSampler(r.device, &samplerInfo, nil, &r.textureSampler), "create texture sampler")
}

func (r *Renderer) createImage(width, height uint32, format vk.Format, tiling vk.ImageTiling, usage vk.ImageUsageFlags, properties vk.MemoryPropertyFlags, image *vk.Image, memory *vk.DeviceMemory) error {
	createInfo := vk.ImageCreateInfo{
		SType:         vk.StructureTypeImageCreateInfo,
		ImageType:     vk.ImageType2d,
		Format:        format,
		Extent:        vk.Extent3D{Width: width, Height: height, Depth: 1},
		MipLevels:     1,
		ArrayLayers:   1,
		Samples:       vk.SampleCount1Bit,
		Tiling:        tiling,
		Usage:         usage,
		SharingMode:   vk.SharingModeExclusive,
		InitialLayout: vk.ImageLayoutUndefined,
	}
	if err := check(vk.CreateImage(r.device, &createInfo, nil, image), "create image"); err != nil {
		return err
	}

	var requirements vk.MemoryRequirements
	vk.GetImageMemoryRequirements(r.device, *image, &requirements)
	requirements.Deref()

	memoryType, err := r.findMemoryType(requirements.MemoryTypeBits, properties)
	if err != nil {
		vk.DestroyImage(r.device, *image, nil)
		*image = vk.NullImage
		return err
	}
	allocInfo := vk.MemoryAllocateInfo{
		SType:           vk.StructureTypeMemoryAllocateInfo,
		AllocationSize:  requirements.Size,
		MemoryTypeIndex: memoryType,
	}
	if err := check(vk.AllocateMemory(r.device, &allocInfo, nil, memory), "allocate image memory"); err != nil {
		vk.DestroyImage(r.device, *image, nil)
		*image = vk.NullImage
		return err
	}
	if err := check(vk.BindImageMemory(r.device, *image, *memory, 0), "bind image memory"); err != nil {
		vk.FreeMemory(r.device, *memory, nil)
		vk.DestroyImage(r.device, *image, nil)
		*memory = vk.NullDeviceMemory
		*image = vk.NullImage
		return err
	}
	return nil
}

func (r *Renderer) createQuadDescriptors() error {
	binding := vk.DescriptorSetLayoutBinding{
		Binding:         0,
		DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
		DescriptorCount: 1,
		StageFlags:      vk.ShaderStageFlags(vk.ShaderStageFragmentBit),
	}
	layoutInfo := vk.DescriptorSetLayoutCreateInfo{
		SType:        vk.StructureTypeDescriptorSetLayoutCreateInfo,
		BindingCount: 1,
		PBindings:    []vk.DescriptorSetLayoutBinding{binding},
	}
	if err := check(vk.CreateDescriptorSetLayout(r.device, &layoutInfo, nil, &r.descriptorSetLayout), "create descriptor set layout"); err != nil {
		return err
	}

	poolSize := vk.DescriptorPoolSize{
		Type:            vk.DescriptorTypeCombinedImageSampler,
		DescriptorCount: 1,
	}
	poolInfo := vk.DescriptorPoolCreateInfo{
		SType:         vk.StructureTypeDescriptorPoolCreateInfo,
		MaxSets:       1,
		PoolSizeCount: 1,
		PPoolSizes:    []vk.DescriptorPoolSize{poolSize},
	}
	if err := check(vk.CreateDescriptorPool(r.device, &poolInfo, nil, &r.descriptorPool), "create descriptor pool"); err != nil {
		return err
	}

	allocInfo := vk.DescriptorSetAllocateInfo{
		SType:              vk.StructureTypeDescriptorSetAllocateInfo,
		DescriptorPool:     r.descriptorPool,
		DescriptorSetCount: 1,
		PSetLayouts:        []vk.DescriptorSetLayout{r.descriptorSetLayout},
	}
	if err := check(vk.AllocateDescriptorSets(r.device, &allocInfo, &r.descriptorSet), "allocate descriptor set"); err != nil {
		return err
	}

	imageInfo := vk.DescriptorImageInfo{
		Sampler:     r.textureSampler,
		ImageView:   r.textureImageView,
		ImageLayout: vk.ImageLayoutShaderReadOnlyOptimal,
	}
	write := vk.WriteDescriptorSet{
		SType:           vk.StructureTypeWriteDescriptorSet,
		DstSet:          r.descriptorSet,
		DstBinding:      0,
		DescriptorCount: 1,
		DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
		PImageInfo:      []vk.DescriptorImageInfo{imageInfo},
	}
	vk.UpdateDescriptorSets(r.device, 1, []vk.WriteDescriptorSet{write}, 0, nil)
	return nil
}

func (r *Renderer) copyToMemory(memory vk.DeviceMemory, data []byte) error {
	var mapped unsafe.Pointer
	if err := check(vk.MapMemory(r.device, memory, 0, vk.DeviceSize(len(data)), 0, &mapped), "map device memory"); err != nil {
		return err
	}
	vk.Memcopy(mapped, data)
	vk.UnmapMemory(r.device, memory)
	return nil
}

func (r *Renderer) beginSingleTimeCommands() (vk.CommandBuffer, error) {
	allocInfo := vk.CommandBufferAllocateInfo{
		SType:              vk.StructureTypeCommandBufferAllocateInfo,
		CommandPool:        r.commandPool,
		Level:              vk.CommandBufferLevelPrimary,
		CommandBufferCount: 1,
	}
	commandBuffers := []vk.CommandBuffer{nil}
	if err := check(vk.AllocateCommandBuffers(r.device, &allocInfo, commandBuffers), "allocate upload command buffer"); err != nil {
		return nil, err
	}
	commandBuffer := commandBuffers[0]
	beginInfo := vk.CommandBufferBeginInfo{
		SType: vk.StructureTypeCommandBufferBeginInfo,
		Flags: vk.CommandBufferUsageFlags(vk.CommandBufferUsageOneTimeSubmitBit),
	}
	if err := check(vk.BeginCommandBuffer(commandBuffer, &beginInfo), "begin upload command buffer"); err != nil {
		vk.FreeCommandBuffers(r.device, r.commandPool, 1, []vk.CommandBuffer{commandBuffer})
		return nil, err
	}
	return commandBuffer, nil
}

func (r *Renderer) endSingleTimeCommands(commandBuffer vk.CommandBuffer) error {
	defer vk.FreeCommandBuffers(r.device, r.commandPool, 1, []vk.CommandBuffer{commandBuffer})
	if err := check(vk.EndCommandBuffer(commandBuffer), "end upload command buffer"); err != nil {
		return err
	}
	submit := vk.SubmitInfo{
		SType:              vk.StructureTypeSubmitInfo,
		CommandBufferCount: 1,
		PCommandBuffers:    []vk.CommandBuffer{commandBuffer},
	}
	err := check(vk.QueueSubmit(r.graphicsQueue, 1, []vk.SubmitInfo{submit}, vk.NullFence), "submit upload command buffer")
	if err == nil {
		err = check(vk.QueueWaitIdle(r.graphicsQueue), "wait for upload command buffer")
	}
	return err
}

func (r *Renderer) copyBuffer(src, dst vk.Buffer, size vk.DeviceSize) error {
	commandBuffer, err := r.beginSingleTimeCommands()
	if err != nil {
		return err
	}
	region := vk.BufferCopy{Size: size}
	vk.CmdCopyBuffer(commandBuffer, src, dst, 1, []vk.BufferCopy{region})
	return r.endSingleTimeCommands(commandBuffer)
}

func (r *Renderer) copyBufferToImage(buffer vk.Buffer, image vk.Image, width, height uint32) error {
	commandBuffer, err := r.beginSingleTimeCommands()
	if err != nil {
		return err
	}
	region := vk.BufferImageCopy{
		ImageSubresource: vk.ImageSubresourceLayers{
			AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
			MipLevel:       0,
			BaseArrayLayer: 0,
			LayerCount:     1,
		},
		ImageExtent: vk.Extent3D{Width: width, Height: height, Depth: 1},
	}
	vk.CmdCopyBufferToImage(commandBuffer, buffer, image, vk.ImageLayoutTransferDstOptimal, 1, []vk.BufferImageCopy{region})
	return r.endSingleTimeCommands(commandBuffer)
}

func (r *Renderer) transitionImageLayout(image vk.Image, oldLayout, newLayout vk.ImageLayout) error {
	commandBuffer, err := r.beginSingleTimeCommands()
	if err != nil {
		return err
	}

	barrier := vk.ImageMemoryBarrier{
		SType:               vk.StructureTypeImageMemoryBarrier,
		OldLayout:           oldLayout,
		NewLayout:           newLayout,
		SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
		DstQueueFamilyIndex: vk.QueueFamilyIgnored,
		Image:               image,
		SubresourceRange: vk.ImageSubresourceRange{
			AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
			BaseMipLevel:   0,
			LevelCount:     1,
			BaseArrayLayer: 0,
			LayerCount:     1,
		},
	}

	var srcStage, dstStage vk.PipelineStageFlags
	switch {
	case oldLayout == vk.ImageLayoutUndefined && newLayout == vk.ImageLayoutTransferDstOptimal:
		barrier.SrcAccessMask = 0
		barrier.DstAccessMask = vk.AccessFlags(vk.AccessTransferWriteBit)
		srcStage = vk.PipelineStageFlags(vk.PipelineStageTopOfPipeBit)
		dstStage = vk.PipelineStageFlags(vk.PipelineStageTransferBit)
	case oldLayout == vk.ImageLayoutTransferDstOptimal && newLayout == vk.ImageLayoutShaderReadOnlyOptimal:
		barrier.SrcAccessMask = vk.AccessFlags(vk.AccessTransferWriteBit)
		barrier.DstAccessMask = vk.AccessFlags(vk.AccessShaderReadBit)
		srcStage = vk.PipelineStageFlags(vk.PipelineStageTransferBit)
		dstStage = vk.PipelineStageFlags(vk.PipelineStageFragmentShaderBit)
	default:
		vk.FreeCommandBuffers(r.device, r.commandPool, 1, []vk.CommandBuffer{commandBuffer})
		return fmt.Errorf("vulkan: unsupported image layout transition %d -> %d", oldLayout, newLayout)
	}

	vk.CmdPipelineBarrier(commandBuffer, srcStage, dstStage, 0, 0, nil, 0, nil, 1, []vk.ImageMemoryBarrier{barrier})
	return r.endSingleTimeCommands(commandBuffer)
}

func (r *Renderer) findMemoryType(typeFilter uint32, properties vk.MemoryPropertyFlags) (uint32, error) {
	var memoryProperties vk.PhysicalDeviceMemoryProperties
	vk.GetPhysicalDeviceMemoryProperties(r.physicalDevice, &memoryProperties)
	memoryProperties.Deref()
	for i := uint32(0); i < memoryProperties.MemoryTypeCount; i++ {
		memoryType := memoryProperties.MemoryTypes[i]
		memoryType.Deref()
		if typeFilter&(1<<i) != 0 && memoryType.PropertyFlags&properties == properties {
			return i, nil
		}
	}
	return 0, fmt.Errorf("vulkan: no memory type supports 0x%x with properties 0x%x", typeFilter, properties)
}

func quadVertexBytes() []byte {
	vertices := [][4]float32{
		{-0.65, -0.65, 0, 1},
		{0.65, -0.65, 1, 1},
		{0.65, 0.65, 1, 0},
		{-0.65, 0.65, 0, 0},
	}
	data := make([]byte, len(vertices)*quadVertexStride)
	for i, vertex := range vertices {
		offset := i * quadVertexStride
		for j, value := range vertex {
			binary.LittleEndian.PutUint32(data[offset+j*4:], math.Float32bits(value))
		}
	}
	return data
}

func quadIndexBytes() []byte {
	indices := []uint16{0, 1, 2, 2, 3, 0}
	data := make([]byte, len(indices)*2)
	for i, index := range indices {
		binary.LittleEndian.PutUint16(data[i*2:], index)
	}
	return data
}
