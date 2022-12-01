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
	ApplicationVersion: vk.MakeVersion(1, 0, 0),
	PApplicationName:   "VulkanDraw\x00",
	PEngineName:        "vulkango.com\x00",
	ApiVersion:         vk.ApiVersion10,
}

func main() {
	app.Main(func(a app.AppDelegate) {
		defer catcher.Catch(
			catcher.RecvLog(true),
			catcher.RecvDie(-1),
		)
		var (
			v        vulkandraw.VulkanDeviceInfo
			s        vulkandraw.VulkanSwapchainInfo
			r        vulkandraw.VulkanRenderInfo
			b        vulkandraw.VulkanBufferInfo
			gfx      vulkandraw.VulkanGfxPipelineInfo
			vkActive = false
		)

		a.InitDone()
		for {
			select {
			case event := <-a.LifecycleEvents():
				switch event.Kind {
				case app.ViewDidLoad:
					err := vk.SetDefaultGetInstanceProcAddr()
					orPanic(err)
					err = vk.Init()
					orPanic(err)

					// differs between Android, iOS and GLFW
					createSurface := func(instance vk.Instance) vk.Surface {
						var surface vk.Surface
						result := vk.CreateWindowSurface(instance, event.View, nil, &surface)
						if result == vk.Success {
							//fmt.Println("CreateWindowSurface - Success")
						}
						if err := vk.Error(result); err != nil {
							vk.DestroyInstance(instance, nil)
							//fmt.Printf("vkCreateWindowSurface failed with %s\n", err)
							panic(err)
						}
						return surface
					}

					v, err = vulkandraw.NewVulkanDevice(appInfo, vk.GetRequiredInstanceExtensions(), createSurface)
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
					vulkandraw.DrawFrame(v, s, r)
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
