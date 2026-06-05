package vulkan

import (
	"fmt"

	vk "github.com/sarattha/lumago/engine/renderer/vulkan/internal/vk"
)

type lightingRenderBuffer struct {
	Target      lightingTarget
	Image       vk.Image
	Memory      vk.DeviceMemory
	View        vk.ImageView
	RenderPass  vk.RenderPass
	Framebuffer vk.Framebuffer
}

type lightingRenderBuffers struct {
	SceneColor    lightingRenderBuffer
	SceneNormal   lightingRenderBuffer
	LightBuffer   lightingRenderBuffer
	SceneEmissive lightingRenderBuffer
}

func (r *Renderer) createLightingRenderBuffers() error {
	r.cleanupLightingRenderBuffers()

	var err error
	if r.lightingBuffers.SceneColor, err = r.createLightingRenderBuffer(r.lightingTargets.SceneColor); err != nil {
		r.cleanupLightingRenderBuffers()
		return fmt.Errorf("create scene color render buffer: %w", err)
	}
	if r.lightingBuffers.SceneNormal, err = r.createLightingRenderBuffer(r.lightingTargets.SceneNormal); err != nil {
		r.cleanupLightingRenderBuffers()
		return fmt.Errorf("create scene normal render buffer: %w", err)
	}
	if r.lightingBuffers.LightBuffer, err = r.createLightingRenderBuffer(r.lightingTargets.LightBuffer); err != nil {
		r.cleanupLightingRenderBuffers()
		return fmt.Errorf("create light buffer render buffer: %w", err)
	}
	if r.lightingBuffers.SceneEmissive, err = r.createLightingRenderBuffer(r.lightingTargets.SceneEmissive); err != nil {
		r.cleanupLightingRenderBuffers()
		return fmt.Errorf("create scene emissive render buffer: %w", err)
	}
	return nil
}

func (r *Renderer) createLightingRenderBuffer(target lightingTarget) (lightingRenderBuffer, error) {
	buffer := lightingRenderBuffer{Target: target}
	if target.Width == 0 || target.Height == 0 {
		return buffer, fmt.Errorf("vulkan: invalid %s extent %dx%d", target.Name, target.Width, target.Height)
	}

	if err := r.createImage(
		target.Width,
		target.Height,
		target.Format,
		vk.ImageTilingOptimal,
		vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit|vk.ImageUsageSampledBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyDeviceLocalBit),
		&buffer.Image,
		&buffer.Memory,
	); err != nil {
		return buffer, err
	}

	viewInfo := vk.ImageViewCreateInfo{
		SType:    vk.StructureTypeImageViewCreateInfo,
		Image:    buffer.Image,
		ViewType: vk.ImageViewType2d,
		Format:   target.Format,
		SubresourceRange: vk.ImageSubresourceRange{
			AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
			BaseMipLevel:   0,
			LevelCount:     1,
			BaseArrayLayer: 0,
			LayerCount:     1,
		},
	}
	if err := check(vk.CreateImageView(r.device, &viewInfo, nil, &buffer.View), "create lighting image view"); err != nil {
		r.destroyLightingRenderBuffer(&buffer)
		return lightingRenderBuffer{}, err
	}

	renderPass, err := r.createLightingRenderPass(target.Format)
	if err != nil {
		r.destroyLightingRenderBuffer(&buffer)
		return lightingRenderBuffer{}, err
	}
	buffer.RenderPass = renderPass

	createInfo := vk.FramebufferCreateInfo{
		SType:           vk.StructureTypeFramebufferCreateInfo,
		RenderPass:      buffer.RenderPass,
		AttachmentCount: 1,
		PAttachments:    []vk.ImageView{buffer.View},
		Width:           target.Width,
		Height:          target.Height,
		Layers:          1,
	}
	if err := check(vk.CreateFramebuffer(r.device, &createInfo, nil, &buffer.Framebuffer), "create lighting framebuffer"); err != nil {
		r.destroyLightingRenderBuffer(&buffer)
		return lightingRenderBuffer{}, err
	}
	return buffer, nil
}

func (r *Renderer) createLightingRenderPass(format vk.Format) (vk.RenderPass, error) {
	attachment := vk.AttachmentDescription{
		Format:         format,
		Samples:        vk.SampleCount1Bit,
		LoadOp:         vk.AttachmentLoadOpClear,
		StoreOp:        vk.AttachmentStoreOpStore,
		StencilLoadOp:  vk.AttachmentLoadOpDontCare,
		StencilStoreOp: vk.AttachmentStoreOpDontCare,
		InitialLayout:  vk.ImageLayoutUndefined,
		FinalLayout:    vk.ImageLayoutShaderReadOnlyOptimal,
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
	dependencies := []vk.SubpassDependency{
		{
			SrcSubpass:    vk.SubpassExternal,
			DstSubpass:    0,
			SrcStageMask:  vk.PipelineStageFlags(vk.PipelineStageFragmentShaderBit),
			DstStageMask:  vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit),
			DstAccessMask: vk.AccessFlags(vk.AccessColorAttachmentWriteBit),
		},
		{
			SrcSubpass:    0,
			DstSubpass:    vk.SubpassExternal,
			SrcStageMask:  vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit),
			DstStageMask:  vk.PipelineStageFlags(vk.PipelineStageFragmentShaderBit),
			SrcAccessMask: vk.AccessFlags(vk.AccessColorAttachmentWriteBit),
			DstAccessMask: vk.AccessFlags(vk.AccessShaderReadBit),
		},
	}
	createInfo := vk.RenderPassCreateInfo{
		SType:           vk.StructureTypeRenderPassCreateInfo,
		AttachmentCount: 1,
		PAttachments:    []vk.AttachmentDescription{attachment},
		SubpassCount:    1,
		PSubpasses:      []vk.SubpassDescription{subpass},
		DependencyCount: uint32(len(dependencies)),
		PDependencies:   dependencies,
	}
	var renderPass vk.RenderPass
	return renderPass, check(vk.CreateRenderPass(r.device, &createInfo, nil, &renderPass), "create lighting render pass")
}

func (r *Renderer) cleanupLightingRenderBuffers() {
	r.destroyLightingRenderBuffer(&r.lightingBuffers.SceneColor)
	r.destroyLightingRenderBuffer(&r.lightingBuffers.SceneNormal)
	r.destroyLightingRenderBuffer(&r.lightingBuffers.LightBuffer)
	r.destroyLightingRenderBuffer(&r.lightingBuffers.SceneEmissive)
	r.lightingBuffers = lightingRenderBuffers{}
}

func (r *Renderer) destroyLightingRenderBuffer(buffer *lightingRenderBuffer) {
	if r.device == vk.NullDevice {
		return
	}
	if buffer.Framebuffer != vk.NullFramebuffer {
		vk.DestroyFramebuffer(r.device, buffer.Framebuffer, nil)
		buffer.Framebuffer = vk.NullFramebuffer
	}
	if buffer.RenderPass != vk.NullRenderPass {
		vk.DestroyRenderPass(r.device, buffer.RenderPass, nil)
		buffer.RenderPass = vk.NullRenderPass
	}
	if buffer.View != vk.NullImageView {
		vk.DestroyImageView(r.device, buffer.View, nil)
		buffer.View = vk.NullImageView
	}
	if buffer.Image != vk.NullImage {
		vk.DestroyImage(r.device, buffer.Image, nil)
		buffer.Image = vk.NullImage
	}
	if buffer.Memory != vk.NullDeviceMemory {
		vk.FreeMemory(r.device, buffer.Memory, nil)
		buffer.Memory = vk.NullDeviceMemory
	}
}
