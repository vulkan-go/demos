package vulkancube

import (
	"bytes"
	"errors"
	"image"
	"image/draw"
	"image/png"
	"log"
	"unsafe"

	as "github.com/vulkan-go/asche"
	vk "github.com/vulkan-go/vulkan"
	lin "github.com/xlab/linmath"
)

func NewSpinningCube(spinAngle float32) *SpinningCube {
	a := &SpinningCube{
		spinAngle: spinAngle,
		eyeVec:    &lin.Vec3{0.0, 3.0, 5.0},
		originVec: &lin.Vec3{0.0, 0.0, 0.0},
		upVec:     &lin.Vec3{0.0, 1.0, 0.0},
	}

	a.projectionMatrix.Perspective(lin.DegreesToRadians(45.0), 1.0, 0.1, 100.0)
	a.viewMatrix.LookAt(a.eyeVec, a.originVec, a.upVec)
	a.modelMatrix.Identity()
	a.projectionMatrix[1][1] *= -1 // Flip projection matrix from GL to Vulkan orientation.
	return a
}

type SpinningCube struct {
	as.BaseVulkanApp

	width      uint32
	height     uint32
	format     vk.Format
	colorSpace vk.ColorSpace

	textures          []*Texture
	depth             *Depth
	useStagingBuffers bool

	descPool vk.DescriptorPool

	pipelineLayout vk.PipelineLayout
	descLayout     vk.DescriptorSetLayout
	pipelineCache  vk.PipelineCache
	renderPass     vk.RenderPass
	pipeline       vk.Pipeline

	frameIndex int

	projectionMatrix lin.Mat4x4
	viewMatrix       lin.Mat4x4
	modelMatrix      lin.Mat4x4

	eyeVec    *lin.Vec3
	originVec *lin.Vec3
	upVec     *lin.Vec3

	spinAngle float32
}

func (s *SpinningCube) prepareDepth() {
	dev := s.Context().Device()
	depthFormat := vk.FormatD16Unorm
	s.depth = &Depth{
		format: depthFormat,
	}
	ret := vk.CreateImage(dev, &vk.ImageCreateInfo{
		SType:     vk.StructureTypeImageCreateInfo,
		ImageType: vk.ImageType2d,
		Format:    depthFormat,
		Extent: vk.Extent3D{
			Width:  s.width,
			Height: s.height,
			Depth:  1,
		},
		MipLevels:   1,
		ArrayLayers: 1,
		Samples:     vk.SampleCount1Bit,
		Tiling:      vk.ImageTilingOptimal,
		Usage:       vk.ImageUsageFlags(vk.ImageUsageDepthStencilAttachmentBit),
	}, nil, &s.depth.image)
	orPanic(as.NewError(ret))

	var memReqs vk.MemoryRequirements
	vk.GetImageMemoryRequirements(dev, s.depth.image, &memReqs)
	memReqs.Deref()

	memProps := s.Context().Platform().MemoryProperties()
	memTypeIndex, _ := as.FindRequiredMemoryTypeFallback(memProps,
		vk.MemoryPropertyFlagBits(memReqs.MemoryTypeBits), vk.MemoryPropertyDeviceLocalBit)
	s.depth.memAlloc = &vk.MemoryAllocateInfo{
		SType:           vk.StructureTypeMemoryAllocateInfo,
		AllocationSize:  memReqs.Size,
		MemoryTypeIndex: memTypeIndex,
	}

	var mem vk.DeviceMemory
	ret = vk.AllocateMemory(dev, s.depth.memAlloc, nil, &mem)
	orPanic(as.NewError(ret))
	s.depth.mem = mem

	ret = vk.BindImageMemory(dev, s.depth.image, s.depth.mem, 0)
	orPanic(as.NewError(ret))

	var view vk.ImageView
	ret = vk.CreateImageView(dev, &vk.ImageViewCreateInfo{
		SType:  vk.StructureTypeImageViewCreateInfo,
		Format: depthFormat,
		SubresourceRange: vk.ImageSubresourceRange{
			AspectMask: vk.ImageAspectFlags(vk.ImageAspectDepthBit),
			LevelCount: 1,
			LayerCount: 1,
		},
		ViewType: vk.ImageViewType2d,
		Image:    s.depth.image,
	}, nil, &view)
	orPanic(as.NewError(ret))
	s.depth.view = view
}

var texEnabled = []string{
	"textures/gopher.png",
}

func (s *SpinningCube) prepareTextureImage(path string, tiling vk.ImageTiling,
	usage vk.ImageUsageFlagBits, memoryProps vk.MemoryPropertyFlagBits) *Texture {

	dev := s.Context().Device()
	texFormat := vk.FormatR8g8b8a8Unorm
	_, width, height, err := loadTextureData(path, 0)
	if err != nil {
		orPanic(err)
	}
	tex := &Texture{
		texWidth:    int32(width),
		texHeight:   int32(height),
		imageLayout: vk.ImageLayoutShaderReadOnlyOptimal,
	}

	var image vk.Image
	ret := vk.CreateImage(dev, &vk.ImageCreateInfo{
		SType:     vk.StructureTypeImageCreateInfo,
		ImageType: vk.ImageType2d,
		Format:    texFormat,
		Extent: vk.Extent3D{
			Width:  uint32(width),
			Height: uint32(height),
			Depth:  1,
		},
		MipLevels:     1,
		ArrayLayers:   1,
		Samples:       vk.SampleCount1Bit,
		Tiling:        tiling,
		Usage:         vk.ImageUsageFlags(usage),
		InitialLayout: vk.ImageLayoutPreinitialized,
	}, nil, &image)
	orPanic(as.NewError(ret))
	tex.image = image

	var memReqs vk.MemoryRequirements
	vk.GetImageMemoryRequirements(dev, tex.image, &memReqs)
	memReqs.Deref()

	memProps := s.Context().Platform().MemoryProperties()
	memTypeIndex, _ := as.FindRequiredMemoryTypeFallback(memProps,
		vk.MemoryPropertyFlagBits(memReqs.MemoryTypeBits), memoryProps)
	tex.memAlloc = &vk.MemoryAllocateInfo{
		SType:           vk.StructureTypeMemoryAllocateInfo,
		AllocationSize:  memReqs.Size,
		MemoryTypeIndex: memTypeIndex,
	}
	var mem vk.DeviceMemory
	ret = vk.AllocateMemory(dev, tex.memAlloc, nil, &mem)
	orPanic(as.NewError(ret))
	tex.mem = mem
	ret = vk.BindImageMemory(dev, tex.image, tex.mem, 0)
	orPanic(as.NewError(ret))

	hostVisible := memoryProps&vk.MemoryPropertyHostVisibleBit != 0
	if hostVisible {
		var layout vk.SubresourceLayout
		vk.GetImageSubresourceLayout(dev, tex.image, &vk.ImageSubresource{
			AspectMask: vk.ImageAspectFlags(vk.ImageAspectColorBit),
		}, &layout)
		layout.Deref()

		data, _, _, err := loadTextureData(path, int(layout.RowPitch))
		orPanic(err)
		if len(data) > 0 {
			var pData unsafe.Pointer
			ret = vk.MapMemory(dev, tex.mem, 0, vk.DeviceSize(len(data)), 0, &pData)
			if isError(ret) {
				log.Printf("vulkan warning: failed to map device memory for data (len=%d)", len(data))
				return tex
			}
			n := vk.Memcopy(pData, data)
			if n != len(data) {
				log.Printf("vulkan warning: failed to copy data, %d != %d", n, len(data))
			}
			vk.UnmapMemory(dev, tex.mem)
		}
	}
	return tex
}

func (s *SpinningCube) setImageLayout(image vk.Image, aspectMask vk.ImageAspectFlagBits,
	oldImageLayout, newImageLayout vk.ImageLayout,
	srcAccessMask vk.AccessFlagBits,
	srcStages, dstStages vk.PipelineStageFlagBits) {

	cmd := s.Context().CommandBuffer()
	if cmd == nil {
		orPanic(errors.New("vulkan: command buffer not initialized"))
	}

	imageMemoryBarrier := vk.ImageMemoryBarrier{
		SType:         vk.StructureTypeImageMemoryBarrier,
		SrcAccessMask: vk.AccessFlags(srcAccessMask),
		DstAccessMask: 0,
		OldLayout:     oldImageLayout,
		NewLayout:     newImageLayout,
		SubresourceRange: vk.ImageSubresourceRange{
			AspectMask: vk.ImageAspectFlags(aspectMask),
			LayerCount: 1,
			LevelCount: 1,
		},
		Image: image,
	}
	switch newImageLayout {
	case vk.ImageLayoutTransferDstOptimal:
		// make sure anything that was copying from this image has completed
		imageMemoryBarrier.DstAccessMask = vk.AccessFlags(vk.AccessTransferWriteBit)
	case vk.ImageLayoutColorAttachmentOptimal:
		imageMemoryBarrier.DstAccessMask = vk.AccessFlags(vk.AccessColorAttachmentWriteBit)
	case vk.ImageLayoutDepthStencilAttachmentOptimal:
		imageMemoryBarrier.DstAccessMask = vk.AccessFlags(vk.AccessDepthStencilAttachmentWriteBit)
	case vk.ImageLayoutShaderReadOnlyOptimal:
		imageMemoryBarrier.DstAccessMask =
			vk.AccessFlags(vk.AccessShaderReadBit) | vk.AccessFlags(vk.AccessInputAttachmentReadBit)
	case vk.ImageLayoutTransferSrcOptimal:
		imageMemoryBarrier.DstAccessMask = vk.AccessFlags(vk.AccessTransferReadBit)
	case vk.ImageLayoutPresentSrc:
		imageMemoryBarrier.DstAccessMask = vk.AccessFlags(vk.AccessMemoryReadBit)
	default:
		imageMemoryBarrier.DstAccessMask = 0
	}

	vk.CmdPipelineBarrier(cmd,
		vk.PipelineStageFlags(srcStages), vk.PipelineStageFlags(dstStages),
		0, 0, nil, 0, nil, 1, []vk.ImageMemoryBarrier{imageMemoryBarrier})
}

func (s *SpinningCube) prepareTextures() {
	dev := s.Context().Device()
	texFormat := vk.FormatR8g8b8a8Unorm
	var props vk.FormatProperties
	gpu := s.Context().Platform().PhysicalDevice()
	vk.GetPhysicalDeviceFormatProperties(gpu, texFormat, &props)
	props.Deref()

	prepareTex := func(path string) *Texture {
		var tex *Texture

		if (props.LinearTilingFeatures&vk.FormatFeatureFlags(vk.FormatFeatureSampledImageBit) != 0) &&
			!s.useStagingBuffers {
			// -> device can texture using linear textures

			tex = s.prepareTextureImage(path, vk.ImageTilingLinear, vk.ImageUsageSampledBit,
				vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit)

			// Nothing in the pipeline needs to be complete to start, and don't allow fragment
			// shader to run until layout transition completes
			s.setImageLayout(tex.image, vk.ImageAspectColorBit,
				vk.ImageLayoutPreinitialized, tex.imageLayout,
				vk.AccessHostWriteBit,
				vk.PipelineStageTopOfPipeBit, vk.PipelineStageFragmentShaderBit)

		} else if props.OptimalTilingFeatures&vk.FormatFeatureFlags(vk.FormatFeatureSampledImageBit) != 0 {
			//  Must use staging buffer to copy linear texture to optimized
			log.Println("vulkan warn: using staging buffers")

			staging := s.prepareTextureImage(path, vk.ImageTilingLinear, vk.ImageUsageTransferSrcBit,
				vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit)
			tex = s.prepareTextureImage(path, vk.ImageTilingOptimal,
				vk.ImageUsageTransferDstBit|vk.ImageUsageSampledBit, vk.MemoryPropertyDeviceLocalBit)

			s.setImageLayout(staging.image, vk.ImageAspectColorBit,
				vk.ImageLayoutPreinitialized, vk.ImageLayoutTransferSrcOptimal,
				vk.AccessHostWriteBit,
				vk.PipelineStageTopOfPipeBit, vk.PipelineStageTransferBit)

			s.setImageLayout(tex.image, vk.ImageAspectColorBit,
				vk.ImageLayoutPreinitialized, vk.ImageLayoutTransferDstOptimal,
				vk.AccessHostWriteBit,
				vk.PipelineStageTopOfPipeBit, vk.PipelineStageTransferBit)

			cmd := s.Context().CommandBuffer()
			if cmd == nil {
				orPanic(errors.New("vulkan: command buffer not initialized"))
			}
			vk.CmdCopyImage(cmd, staging.image, vk.ImageLayoutTransferSrcOptimal,
				tex.image, vk.ImageLayoutTransferDstOptimal,
				1, []vk.ImageCopy{{
					SrcSubresource: vk.ImageSubresourceLayers{
						AspectMask: vk.ImageAspectFlags(vk.ImageAspectColorBit),
						LayerCount: 1,
					},
					SrcOffset: vk.Offset3D{
						X: 0, Y: 0, Z: 0,
					},
					DstSubresource: vk.ImageSubresourceLayers{
						AspectMask: vk.ImageAspectFlags(vk.ImageAspectColorBit),
						LayerCount: 1,
					},
					DstOffset: vk.Offset3D{
						X: 0, Y: 0, Z: 0,
					},
					Extent: vk.Extent3D{
						Width:  uint32(staging.texWidth),
						Height: uint32(staging.texHeight),
						Depth:  1,
					},
				}})
			s.setImageLayout(tex.image, vk.ImageAspectColorBit,
				vk.ImageLayoutTransferDstOptimal, tex.imageLayout,
				vk.AccessTransferWriteBit,
				vk.PipelineStageTransferBit, vk.PipelineStageFragmentShaderBit)
			// cannot destroy until cmd is submitted.. must keep a list somewhere
			// staging.DestroyImage(dev)
		} else {
			orPanic(errors.New("vulkan: R8G8B8A8_UNORM not supported as texture image format"))
		}

		var sampler vk.Sampler
		ret := vk.CreateSampler(dev, &vk.SamplerCreateInfo{
			SType:                   vk.StructureTypeSamplerCreateInfo,
			MagFilter:               vk.FilterNearest,
			MinFilter:               vk.FilterNearest,
			MipmapMode:              vk.SamplerMipmapModeNearest,
			AddressModeU:            vk.SamplerAddressModeClampToEdge,
			AddressModeV:            vk.SamplerAddressModeClampToEdge,
			AddressModeW:            vk.SamplerAddressModeClampToEdge,
			AnisotropyEnable:        vk.False,
			MaxAnisotropy:           1,
			CompareOp:               vk.CompareOpNever,
			BorderColor:             vk.BorderColorFloatOpaqueWhite,
			UnnormalizedCoordinates: vk.False,
		}, nil, &sampler)
		orPanic(as.NewError(ret))
		tex.sampler = sampler

		var view vk.ImageView
		ret = vk.CreateImageView(dev, &vk.ImageViewCreateInfo{
			SType:    vk.StructureTypeImageViewCreateInfo,
			Image:    tex.image,
			ViewType: vk.ImageViewType2d,
			Format:   texFormat,
			Components: vk.ComponentMapping{
				R: vk.ComponentSwizzleR,
				G: vk.ComponentSwizzleG,
				B: vk.ComponentSwizzleB,
				A: vk.ComponentSwizzleA,
			},
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask: vk.ImageAspectFlags(vk.ImageAspectColorBit),
				LevelCount: 1,
				LayerCount: 1,
			},
		}, nil, &view)
		orPanic(as.NewError(ret))
		tex.view = view

		return tex
	}

	s.textures = make([]*Texture, 0, len(texEnabled))
	for _, texFile := range texEnabled {
		s.textures = append(s.textures, prepareTex(texFile))
	}
}

func (s *SpinningCube) drawBuildCommandBuffer(res *as.SwapchainImageResources, cmd vk.CommandBuffer) {
	ret := vk.BeginCommandBuffer(cmd, &vk.CommandBufferBeginInfo{
		SType: vk.StructureTypeCommandBufferBeginInfo,
		Flags: vk.CommandBufferUsageFlags(vk.CommandBufferUsageSimultaneousUseBit),
	})
	orPanic(as.NewError(ret))

	clearValues := make([]vk.ClearValue, 2)
	clearValues[1].SetDepthStencil(1, 0)
	clearValues[0].SetColor([]float32{
		0.2, 0.2, 0.2, 0.2,
	})
	vk.CmdBeginRenderPass(cmd, &vk.RenderPassBeginInfo{
		SType:       vk.StructureTypeRenderPassBeginInfo,
		RenderPass:  s.renderPass,
		Framebuffer: res.Framebuffer(),
		RenderArea: vk.Rect2D{
			Offset: vk.Offset2D{
				X: 0, Y: 0,
			},
			Extent: vk.Extent2D{
				Width:  s.width,
				Height: s.height,
			},
		},
		ClearValueCount: 2,
		PClearValues:    clearValues,
	}, vk.SubpassContentsInline)

	vk.CmdBindPipeline(cmd, vk.PipelineBindPointGraphics, s.pipeline)
	vk.CmdBindDescriptorSets(cmd, vk.PipelineBindPointGraphics, s.pipelineLayout,
		0, 1, []vk.DescriptorSet{res.DescriptorSet()}, 0, nil)
	vk.CmdSetViewport(cmd, 0, 1, []vk.Viewport{{
		Width:    float32(s.width),
		Height:   float32(s.height),
		MinDepth: 0.0,
		MaxDepth: 1.0,
	}})
	vk.CmdSetScissor(cmd, 0, 1, []vk.Rect2D{{
		Offset: vk.Offset2D{
			X: 0, Y: 0,
		},
		Extent: vk.Extent2D{
			Width:  s.width,
			Height: s.height,
		},
	}})

	vk.CmdDraw(cmd, 12*3, 1, 0, 0)
	// Note that ending the renderpass changes the image's layout from
	// vk.ImageLayoutColorAttachmentOptimal to vk.ImageLayoutPresentSrc
	vk.CmdEndRenderPass(cmd)

	graphicsQueueIndex := s.Context().Platform().GraphicsQueueFamilyIndex()
	presentQueueIndex := s.Context().Platform().PresentQueueFamilyIndex()
	if graphicsQueueIndex != presentQueueIndex {
		// Separate Present Queue Case
		//
		// We have to transfer ownership from the graphics queue family to the
		// present queue family to be able to present.  Note that we don't have
		// to transfer from present queue family back to graphics queue family at
		// the start of the next frame because we don't care about the image's
		// contents at that point.
		vk.CmdPipelineBarrier(cmd,
			vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit),
			vk.PipelineStageFlags(vk.PipelineStageBottomOfPipeBit),
			0, 0, nil, 0, nil, 1, []vk.ImageMemoryBarrier{{
				SType:               vk.StructureTypeImageMemoryBarrier,
				SrcAccessMask:       0,
				DstAccessMask:       vk.AccessFlags(vk.AccessColorAttachmentWriteBit),
				OldLayout:           vk.ImageLayoutPresentSrc,
				NewLayout:           vk.ImageLayoutPresentSrc,
				SrcQueueFamilyIndex: graphicsQueueIndex,
				DstQueueFamilyIndex: presentQueueIndex,
				SubresourceRange: vk.ImageSubresourceRange{
					AspectMask: vk.ImageAspectFlags(vk.ImageAspectColorBit),
					LayerCount: 1,
					LevelCount: 1,
				},
				Image: res.Image(),
			}})
	}
	ret = vk.EndCommandBuffer(cmd)
	orPanic(as.NewError(ret))
}

func (s *SpinningCube) prepareCubeDataBuffers() {
	dev := s.Context().Device()

	var VP lin.Mat4x4
	var MVP lin.Mat4x4
	VP.Mult(&s.projectionMatrix, &s.viewMatrix)
	MVP.Mult(&VP, &s.modelMatrix)

	data := vkTexCubeUniform{
		mvp: MVP,
	}
	for i := 0; i < 12*3; i++ {
		data.position[i][0] = gVertexBufferData[i*3]
		data.position[i][1] = gVertexBufferData[i*3+1]
		data.position[i][2] = gVertexBufferData[i*3+2]
		data.position[i][3] = 1.0
		data.attr[i][0] = gUVBufferData[2*i]
		data.attr[i][1] = gUVBufferData[2*i+1]
		data.attr[i][2] = 0
		data.attr[i][3] = 0
	}

	dataRaw := data.Data()
	memProps := s.Context().Platform().MemoryProperties()
	swapchainImageResources := s.Context().SwapchainImageResources()
	for _, res := range swapchainImageResources {
		buf := as.CreateBuffer(dev, memProps, dataRaw, vk.BufferUsageUniformBufferBit)
		res.SetUniformBuffer(buf.Buffer, buf.Memory)
	}
}

func (s *SpinningCube) prepareDescriptorLayout() {
	dev := s.Context().Device()

	var descLayout vk.DescriptorSetLayout
	ret := vk.CreateDescriptorSetLayout(dev, &vk.DescriptorSetLayoutCreateInfo{
		SType:        vk.StructureTypeDescriptorSetLayoutCreateInfo,
		BindingCount: 2,
		PBindings: []vk.DescriptorSetLayoutBinding{
			{
				Binding:         0,
				DescriptorType:  vk.DescriptorTypeUniformBuffer,
				DescriptorCount: 1,
				StageFlags:      vk.ShaderStageFlags(vk.ShaderStageVertexBit),
			}, {
				Binding:         1,
				DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
				DescriptorCount: uint32(len(texEnabled)),
				StageFlags:      vk.ShaderStageFlags(vk.ShaderStageFragmentBit),
			}},
	}, nil, &descLayout)
	orPanic(as.NewError(ret))
	s.descLayout = descLayout

	var pipelineLayout vk.PipelineLayout
	ret = vk.CreatePipelineLayout(dev, &vk.PipelineLayoutCreateInfo{
		SType:          vk.StructureTypePipelineLayoutCreateInfo,
		SetLayoutCount: 1,
		PSetLayouts: []vk.DescriptorSetLayout{
			s.descLayout,
		},
	}, nil, &pipelineLayout)
	orPanic(as.NewError(ret))
	s.pipelineLayout = pipelineLayout
}

func (s *SpinningCube) prepareRenderPass() {
	dev := s.Context().Device()
	// The initial layout for the color and depth attachments will be vk.LayoutUndefined
	// because at the start of the renderpass, we don't care about their contents.
	// At the start of the subpass, the color attachment's layout will be transitioned
	// to vk.LayoutColorAttachmentOptimal and the depth stencil attachment's layout
	// will be transitioned to vk.LayoutDepthStencilAttachmentOptimal.  At the end of
	// the renderpass, the color attachment's layout will be transitioned to
	// vk.LayoutPresentSrc to be ready to present.  This is all done as part of
	// the renderpass, no barriers are necessary.
	var renderPass vk.RenderPass
	ret := vk.CreateRenderPass(dev, &vk.RenderPassCreateInfo{
		SType:           vk.StructureTypeRenderPassCreateInfo,
		AttachmentCount: 2,
		PAttachments: []vk.AttachmentDescription{{
			Format:         s.Context().SwapchainDimensions().Format,
			Samples:        vk.SampleCount1Bit,
			LoadOp:         vk.AttachmentLoadOpClear,
			StoreOp:        vk.AttachmentStoreOpStore,
			StencilLoadOp:  vk.AttachmentLoadOpDontCare,
			StencilStoreOp: vk.AttachmentStoreOpDontCare,
			InitialLayout:  vk.ImageLayoutUndefined,
			FinalLayout:    vk.ImageLayoutPresentSrc,
		}, {
			Format:         s.depth.format,
			Samples:        vk.SampleCount1Bit,
			LoadOp:         vk.AttachmentLoadOpClear,
			StoreOp:        vk.AttachmentStoreOpDontCare,
			StencilLoadOp:  vk.AttachmentLoadOpDontCare,
			StencilStoreOp: vk.AttachmentStoreOpDontCare,
			InitialLayout:  vk.ImageLayoutUndefined,
			FinalLayout:    vk.ImageLayoutDepthStencilAttachmentOptimal,
		}},
		SubpassCount: 1,
		PSubpasses: []vk.SubpassDescription{{
			PipelineBindPoint:    vk.PipelineBindPointGraphics,
			ColorAttachmentCount: 1,
			PColorAttachments: []vk.AttachmentReference{{
				Attachment: 0,
				Layout:     vk.ImageLayoutColorAttachmentOptimal,
			}},
			PDepthStencilAttachment: &vk.AttachmentReference{
				Attachment: 1,
				Layout:     vk.ImageLayoutDepthStencilAttachmentOptimal,
			},
		}},
	}, nil, &renderPass)
	orPanic(as.NewError(ret))
	s.renderPass = renderPass
}

func (s *SpinningCube) preparePipeline() {
	dev := s.Context().Device()

	vs, err := as.LoadShaderModule(dev, MustAsset("shaders/cube.vert.spv"))
	orPanic(err)
	fs, err := as.LoadShaderModule(dev, MustAsset("shaders/cube.frag.spv"))
	orPanic(err)

	var pipelineCache vk.PipelineCache
	ret := vk.CreatePipelineCache(dev, &vk.PipelineCacheCreateInfo{
		SType: vk.StructureTypePipelineCacheCreateInfo,
	}, nil, &pipelineCache)
	orPanic(as.NewError(ret))
	s.pipelineCache = pipelineCache

	pipelineCreateInfos := []vk.GraphicsPipelineCreateInfo{{
		SType:      vk.StructureTypeGraphicsPipelineCreateInfo,
		Layout:     s.pipelineLayout,
		RenderPass: s.renderPass,

		PDynamicState: &vk.PipelineDynamicStateCreateInfo{
			SType:             vk.StructureTypePipelineDynamicStateCreateInfo,
			DynamicStateCount: 2,
			PDynamicStates: []vk.DynamicState{
				vk.DynamicStateScissor,
				vk.DynamicStateViewport,
			},
		},
		PVertexInputState: &vk.PipelineVertexInputStateCreateInfo{
			SType: vk.StructureTypePipelineVertexInputStateCreateInfo,
		},
		PInputAssemblyState: &vk.PipelineInputAssemblyStateCreateInfo{
			SType:    vk.StructureTypePipelineInputAssemblyStateCreateInfo,
			Topology: vk.PrimitiveTopologyTriangleList,
		},
		PRasterizationState: &vk.PipelineRasterizationStateCreateInfo{
			SType:       vk.StructureTypePipelineRasterizationStateCreateInfo,
			PolygonMode: vk.PolygonModeFill,
			CullMode:    vk.CullModeFlags(vk.CullModeBackBit),
			FrontFace:   vk.FrontFaceCounterClockwise,
			LineWidth:   1.0,
		},
		PColorBlendState: &vk.PipelineColorBlendStateCreateInfo{
			SType:           vk.StructureTypePipelineColorBlendStateCreateInfo,
			AttachmentCount: 1,
			PAttachments: []vk.PipelineColorBlendAttachmentState{{
				ColorWriteMask: 0xF,
				BlendEnable:    vk.False,
			}},
		},
		PMultisampleState: &vk.PipelineMultisampleStateCreateInfo{
			SType:                vk.StructureTypePipelineMultisampleStateCreateInfo,
			RasterizationSamples: vk.SampleCount1Bit,
		},
		PViewportState: &vk.PipelineViewportStateCreateInfo{
			SType:         vk.StructureTypePipelineViewportStateCreateInfo,
			ScissorCount:  1,
			ViewportCount: 1,
		},
		PDepthStencilState: &vk.PipelineDepthStencilStateCreateInfo{
			SType:                 vk.StructureTypePipelineDepthStencilStateCreateInfo,
			DepthTestEnable:       vk.True,
			DepthWriteEnable:      vk.True,
			DepthCompareOp:        vk.CompareOpLessOrEqual,
			DepthBoundsTestEnable: vk.False,
			Back: vk.StencilOpState{
				FailOp:    vk.StencilOpKeep,
				PassOp:    vk.StencilOpKeep,
				CompareOp: vk.CompareOpAlways,
			},
			StencilTestEnable: vk.False,
			Front: vk.StencilOpState{
				FailOp:    vk.StencilOpKeep,
				PassOp:    vk.StencilOpKeep,
				CompareOp: vk.CompareOpAlways,
			},
		},
		StageCount: 2,
		PStages: []vk.PipelineShaderStageCreateInfo{{
			SType:  vk.StructureTypePipelineShaderStageCreateInfo,
			Stage:  vk.ShaderStageVertexBit,
			Module: vs,
			PName:  "main\x00",
		}, {
			SType:  vk.StructureTypePipelineShaderStageCreateInfo,
			Stage:  vk.ShaderStageFragmentBit,
			Module: fs,
			PName:  "main\x00",
		}},
	}}
	pipeline := make([]vk.Pipeline, 1)
	ret = vk.CreateGraphicsPipelines(dev, s.pipelineCache, 1, pipelineCreateInfos, nil, pipeline)
	orPanic(as.NewError(ret))
	s.pipeline = pipeline[0]

	vk.DestroyShaderModule(dev, vs, nil)
	vk.DestroyShaderModule(dev, fs, nil)
}

func (s *SpinningCube) prepareDescriptorPool() {
	dev := s.Context().Device()
	swapchainImageResources := s.Context().SwapchainImageResources()
	var descPool vk.DescriptorPool
	ret := vk.CreateDescriptorPool(dev, &vk.DescriptorPoolCreateInfo{
		SType:         vk.StructureTypeDescriptorPoolCreateInfo,
		MaxSets:       uint32(len(swapchainImageResources)),
		PoolSizeCount: 2,
		PPoolSizes: []vk.DescriptorPoolSize{{
			Type:            vk.DescriptorTypeUniformBuffer,
			DescriptorCount: uint32(len(swapchainImageResources)),
		}, {
			Type:            vk.DescriptorTypeCombinedImageSampler,
			DescriptorCount: uint32(len(swapchainImageResources) * len(texEnabled)),
		}},
	}, nil, &descPool)
	orPanic(as.NewError(ret))
	s.descPool = descPool
}

func (s *SpinningCube) prepareDescriptorSet() {
	dev := s.Context().Device()
	swapchainImageResources := s.Context().SwapchainImageResources()

	texInfos := make([]vk.DescriptorImageInfo, 0, len(s.textures))
	for _, tex := range s.textures {
		texInfos = append(texInfos, vk.DescriptorImageInfo{
			Sampler:     tex.sampler,
			ImageView:   tex.view,
			ImageLayout: vk.ImageLayoutGeneral,
		})
	}

	for _, res := range swapchainImageResources {
		var set vk.DescriptorSet
		ret := vk.AllocateDescriptorSets(dev, &vk.DescriptorSetAllocateInfo{
			SType:              vk.StructureTypeDescriptorSetAllocateInfo,
			DescriptorPool:     s.descPool,
			DescriptorSetCount: 1,
			PSetLayouts:        []vk.DescriptorSetLayout{s.descLayout},
		}, &set)
		orPanic(as.NewError(ret))

		res.SetDescriptorSet(set)

		vk.UpdateDescriptorSets(dev, 2, []vk.WriteDescriptorSet{{
			SType:           vk.StructureTypeWriteDescriptorSet,
			DstSet:          set,
			DescriptorCount: 1,
			DescriptorType:  vk.DescriptorTypeUniformBuffer,
			PBufferInfo: []vk.DescriptorBufferInfo{{
				Offset: 0,
				Range:  vk.DeviceSize(vkTexCubeUniformSize),
				Buffer: res.UniformBuffer(),
			}},
		}, {
			SType:           vk.StructureTypeWriteDescriptorSet,
			DstBinding:      1,
			DstSet:          set,
			DescriptorCount: uint32(len(texEnabled)),
			DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
			PImageInfo:      texInfos,
		}}, 0, nil)
	}
}

func (s *SpinningCube) prepareFramebuffers() {
	dev := s.Context().Device()
	swapchainImageResources := s.Context().SwapchainImageResources()

	for _, res := range swapchainImageResources {
		var fb vk.Framebuffer

		ret := vk.CreateFramebuffer(dev, &vk.FramebufferCreateInfo{
			SType:           vk.StructureTypeFramebufferCreateInfo,
			RenderPass:      s.renderPass,
			AttachmentCount: 2,
			PAttachments: []vk.ImageView{
				res.View(),
				s.depth.view,
			},
			Width:  s.width,
			Height: s.height,
			Layers: 1,
		}, nil, &fb)
		orPanic(as.NewError(ret))

		res.SetFramebuffer(fb)
	}
}

func (s *SpinningCube) VulkanContextPrepare() error {
	dim := s.Context().SwapchainDimensions()
	s.height = dim.Height
	s.width = dim.Width

	s.prepareDepth()
	s.prepareTextures()
	s.prepareCubeDataBuffers()
	s.prepareDescriptorLayout()
	s.prepareRenderPass()
	s.preparePipeline()
	s.prepareDescriptorPool()
	s.prepareDescriptorSet()
	s.prepareFramebuffers()

	swapchainImageResources := s.Context().SwapchainImageResources()
	for _, res := range swapchainImageResources {
		s.drawBuildCommandBuffer(res, res.CommandBuffer())
	}
	return nil
}

func (s *SpinningCube) VulkanContextCleanup() error {
	dev := s.Context().Device()
	vk.DestroyDescriptorPool(dev, s.descPool, nil)
	vk.DestroyPipeline(dev, s.pipeline, nil)
	vk.DestroyPipelineCache(dev, s.pipelineCache, nil)
	vk.DestroyRenderPass(dev, s.renderPass, nil)
	vk.DestroyPipelineLayout(dev, s.pipelineLayout, nil)
	vk.DestroyDescriptorSetLayout(dev, s.descLayout, nil)

	for i := 0; i < len(s.textures); i++ {
		s.textures[i].Destroy(dev)
	}
	s.depth.Destroy(dev)
	return nil
}

func (s *SpinningCube) VulkanContextInvalidate(imageIdx int) error {
	dev := s.Context().Device()
	res := s.Context().SwapchainImageResources()[imageIdx]

	var MVP, Model, VP lin.Mat4x4
	VP.Mult(&s.projectionMatrix, &s.viewMatrix)

	// Rotate around the Y axis
	Model.Dup(&s.modelMatrix)
	s.modelMatrix.Rotate(&Model, 0.0, 1.0, 0.0, lin.DegreesToRadians(s.spinAngle))
	MVP.Mult(&VP, &s.modelMatrix)

	data := MVP.Data()
	var pData unsafe.Pointer
	ret := vk.MapMemory(dev, res.UniformMemory(), 0, vk.DeviceSize(len(data)), 0, &pData)
	orPanic(as.NewError(ret))

	n := vk.Memcopy(pData, data)
	if n != len(data) {
		log.Printf("vulkan warning: failed to copy data, %d != %d", n, len(data))
	}
	vk.UnmapMemory(dev, res.UniformMemory())
	return nil
}

func (s *SpinningCube) Destroy() {}

type Texture struct {
	sampler vk.Sampler

	image       vk.Image
	imageLayout vk.ImageLayout

	memAlloc *vk.MemoryAllocateInfo
	mem      vk.DeviceMemory
	view     vk.ImageView

	texWidth  int32
	texHeight int32
}

func (t *Texture) Destroy(dev vk.Device) {
	vk.DestroyImageView(dev, t.view, nil)
	vk.FreeMemory(dev, t.mem, nil)
	vk.DestroyImage(dev, t.image, nil)
	vk.DestroySampler(dev, t.sampler, nil)
}

func (t *Texture) DestroyImage(dev vk.Device) {
	vk.FreeMemory(dev, t.mem, nil)
	vk.DestroyImage(dev, t.image, nil)
}

type Depth struct {
	format   vk.Format
	image    vk.Image
	memAlloc *vk.MemoryAllocateInfo
	mem      vk.DeviceMemory
	view     vk.ImageView
}

func (d *Depth) Destroy(dev vk.Device) {
	vk.DestroyImageView(dev, d.view, nil)
	vk.DestroyImage(dev, d.image, nil)
	vk.FreeMemory(dev, d.mem, nil)
}

// func loadTextureSize(name string) (w int, h int, err error) {
// 	data := MustAsset(name)
// 	r := bytes.NewReader(data)
// 	ppmCfg, err := ppm.DecodeConfig(r)
// 	if err != nil {
// 		return 0, 0, err
// 	}
// 	return ppmCfg.Width, ppmCfg.Height, nil
// }

// func loadTextureData(name string, layout vk.SubresourceLayout) ([]byte, error) {
// 	data := MustAsset(name)
// 	r := bytes.NewReader(data)
// 	img, err := ppm.Decode(r)
// 	if err != nil {
// 		return nil, err
// 	}
// 	newImg := image.NewRGBA(img.Bounds())
// 	newImg.Stride = int(layout.RowPitch)
// 	draw.Draw(newImg, newImg.Bounds(), img, image.ZP, draw.Src)
// 	return []byte(newImg.Pix), nil
// }

func loadTextureData(name string, rowPitch int) ([]byte, int, int, error) {
	data := MustAsset(name)
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, 0, 0, err
	}
	newImg := image.NewRGBA(img.Bounds())
	if rowPitch <= 4*img.Bounds().Dy() {
		// apply the proposed row pitch only if supported,
		// as we're using only optimal textures.
		newImg.Stride = rowPitch
	}
	draw.Draw(newImg, newImg.Bounds(), img, image.ZP, draw.Src)
	size := newImg.Bounds().Size()
	return []byte(newImg.Pix), size.X, size.Y, nil
}

func actualTimeLate(desired, actual, rdur uint64) bool {
	// The desired time was the earliest time that the present should have
	// occured.  In almost every case, the actual time should be later than the
	// desired time.  We should only consider the actual time "late" if it is
	// after "desired + rdur".
	if actual <= desired {
		// The actual time was before or equal to the desired time.  This will
		// probably never happen, but in case it does, return false since the
		// present was obviously NOT late.
		return false
	}
	deadline := actual + rdur
	if actual > deadline {
		return true
	} else {
		return false
	}
}

const million = 1000 * 1000

func canPresentEarlier(earliest, actual, margin, rdur uint64) bool {
	if earliest < actual {
		// Consider whether this present could have occured earlier.  Make sure
		// that earliest time was at least 2msec earlier than actual time, and
		// that the margin was at least 2msec:
		diff := actual - earliest
		if (diff >= (2 * million)) && (margin >= (2 * million)) {
			// This present could have occured earlier because both: 1) the
			// earliest time was at least 2 msec before actual time, and 2) the
			// margin was at least 2msec.
			return true
		}
	}
	return false
}
