package main

import (
	"log"
	"runtime"
	"time"

	"github.com/vulkan-go/demos/vulkandraw"
	"github.com/vulkan-go/glfw/v3.3/glfw"
	vk "github.com/vulkan-go/vulkan"
	"github.com/xlab/closer"
)

var appInfo = &vk.ApplicationInfo{
	SType:              vk.StructureTypeApplicationInfo,
	ApiVersion:         vk.MakeVersion(1, 0, 0),
	ApplicationVersion: vk.MakeVersion(1, 0, 0),
	PApplicationName:   "VulkanDraw\x00",
	PEngineName:        "vulkango.com\x00",
}

func init() {
	runtime.LockOSThread()
}

func main() {
	procAddr := glfw.GetVulkanGetInstanceProcAddress()
	if procAddr == nil {
		panic("GetInstanceProcAddress is nil")
	}
	vk.SetGetInstanceProcAddr(procAddr)

	orPanic(glfw.Init())
	orPanic(vk.Init())
	defer closer.Close()

	var (
		v   vulkandraw.VulkanDeviceInfo
		s   vulkandraw.VulkanSwapchainInfo
		r   vulkandraw.VulkanRenderInfo
		b   vulkandraw.VulkanBufferInfo
		gfx vulkandraw.VulkanGfxPipelineInfo
	)

	glfw.WindowHint(glfw.ClientAPI, glfw.NoAPI)
	glfw.WindowHint(glfw.Resizable, glfw.False)
	window, err := glfw.CreateWindow(640, 480, "Vulkan Info", nil, nil)
	orPanic(err)

	createSurface := func(instance interface{}) uintptr {
		surface, err := window.CreateWindowSurface(instance, nil)
		orPanic(err)
		return surface
	}

	v, err = vulkandraw.NewVulkanDevice(appInfo,
		window.GLFWWindow(),
		window.GetRequiredInstanceExtensions(),
		createSurface)
	orPanic(err)
	s, err = v.CreateSwapchain()
	orPanic(err)
	r, err = vulkandraw.CreateRenderer(v.Device, s.DisplayFormat)
	orPanic(err)
	err = s.CreateFramebuffers(r.RenderPass, nil)
	orPanic(err)
	b, err = v.CreateBuffers()
	orPanic(err)
	gfx, err = vulkandraw.CreateGraphicsPipeline(v.Device, s.DisplaySize, r.RenderPass)
	orPanic(err)
	log.Println("[INFO] swapchain lengths:", s.SwapchainLen)
	err = r.CreateCommandBuffers(s.DefaultSwapchainLen())
	orPanic(err)

	doneC := make(chan struct{}, 2)
	exitC := make(chan struct{}, 2)
	defer closer.Bind(func() {
		exitC <- struct{}{}
		<-doneC
		log.Println("Bye!")
	})
	vulkandraw.VulkanInit(&v, &s, &r, &b, &gfx)

	fpsTicker := time.NewTicker(time.Second / 30)
	for {
		select {
		case <-exitC:
			vulkandraw.DestroyInOrder(&v, &s, &r, &b, &gfx)
			window.Destroy()
			glfw.Terminate()
			fpsTicker.Stop()
			doneC <- struct{}{}
			return
		case <-fpsTicker.C:
			if window.ShouldClose() {
				exitC <- struct{}{}
				continue
			}
			glfw.PollEvents()
			if window.GetAttrib(glfw.Iconified) != 1 {
				vulkandraw.VulkanDrawFrame(v, s, r)
			}
		}
	}
}

func orPanic(err interface{}) {
	switch v := err.(type) {
	case error:
		if v != nil {
			panic(err)
		}
	case vk.Result:
		if err := vk.Error(v); err != nil {
			panic(err)
		}
	case bool:
		if !v {
			panic("condition failed: != true")
		}
	}
}
