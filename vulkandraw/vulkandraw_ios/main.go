package main

import (
	"log"

	"github.com/vulkan-go/demos/vulkandraw"
	vk "github.com/vulkan-go/vulkan"
	"github.com/xlab/catcher"
	"github.com/xlab/ios-go/app"
)

var appInfo = &vk.ApplicationInfo{
	SType:              vk.StructureTypeApplicationInfo,
	ApiVersion:         vk.MakeVersion(1, 0, 0),
	ApplicationVersion: vk.MakeVersion(1, 0, 0),
	PApplicationName:   "VulkanDraw\x00",
	PEngineName:        "vulkango.com\x00",
}

func main() {
	app.Main(func(a app.AppDelegate) {
		defer catcher.Catch(
			catcher.RecvLog(true),
			catcher.RecvDie(-1),
		)
		var (
			v   vulkandraw.VulkanDeviceInfo
			s   vulkandraw.VulkanSwapchainInfo
			r   vulkandraw.VulkanRenderInfo
			b   vulkandraw.VulkanBufferInfo
			gfx vulkandraw.VulkanGfxPipelineInfo

			vkActive bool
		)

		a.InitDone()
		for {
			select {
			case event := <-a.LifecycleEvents():
				switch event.Kind {
				case app.ViewDidLoad:
					err := vk.Init()
					orPanic(err)
					v, err = vulkandraw.NewVulkanDevice(appInfo, event.View)
					orPanic(err)
					s, err = v.CreateSwapchain()
					orPanic(err)
					r, err = vulkandraw.CreateRenderer(v.Device, s.DisplayFormat)
					orPanic(err)
					err = s.CreateFramebuffers(r.RenderPass, vk.NullImageView)
					orPanic(err)
					b, err = v.CreateBuffers()
					orPanic(err)
					gfx, err = vulkandraw.CreateGraphicsPipeline(v.Device, s.DisplaySize, r.RenderPass)
					orPanic(err)
					log.Println("[INFO] swapchain lengths:", s.SwapchainLen)
					err = r.CreateCommandBuffers(s.DefaultSwapchainLen())
					orPanic(err)
					vulkandraw.VulkanInit(&v, &s, &r, &b, &gfx)
					vkActive = true
				case app.DidBecomeActive:
					vkActive = true
				case app.DidEnterBackground:
					vkActive = false
				case app.WillTerminate:
					vkActive = false
					vulkandraw.DestroyInOrder(&v, &s, &r, &b, &gfx)
				}
			case <-a.VSync():
				if vkActive {
					vulkandraw.VulkanDrawFrame(v, s, r)
				}
			}
		}
	})
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
