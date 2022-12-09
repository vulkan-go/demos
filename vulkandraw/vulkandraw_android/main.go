//go:build android
// +build android

package main

import (
	"log"

	vk "github.com/vulkan-go/vulkan"
	"github.com/xlab/android-go/android"
	"github.com/xlab/android-go/app"
	"github.com/xlab/catcher"
	"github.com/vulkan-go/demos/vulkandraw"
)

func init() {
	app.SetLogTag("VulkanDraw")
}

var appInfo = &vk.ApplicationInfo{
	SType:              vk.StructureTypeApplicationInfo,
	ApplicationVersion: vk.MakeVersion(1, 0, 0),
	PApplicationName:   "VulkanDraw\x00",
	PEngineName:        "vulkango.com\x00",
	ApiVersion:         vk.ApiVersion10,
}

func main() {
	nativeWindowEvents := make(chan app.NativeWindowEvent)
	inputQueueEvents := make(chan app.InputQueueEvent, 1)
	inputQueueChan := make(chan *android.InputQueue, 1)

	app.Main(func(a app.NativeActivity) {
		// disable this to get the stack
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

		a.HandleNativeWindowEvents(nativeWindowEvents)
		a.HandleInputQueueEvents(inputQueueEvents)
		// just skip input events (so app won't be dead on touch input)
		go app.HandleInputQueues(inputQueueChan, func() {
			a.InputQueueHandled()
		}, app.SkipInputEvents)
		a.InitDone()

		for {
			select {
			case <-a.LifecycleEvents():
				// ignore
			case event := <-inputQueueEvents:
				switch event.Kind {
				case app.QueueCreated:
					inputQueueChan <- event.Queue
				case app.QueueDestroyed:
					inputQueueChan <- nil
				}
			case event := <-nativeWindowEvents:
				switch event.Kind {
				case app.NativeWindowCreated:
					err := vk.SetDefaultGetInstanceProcAddr()
					orPanic(err)
					err = vk.Init()
					orPanic(err)

					// differs between Android, iOS and GLFW
					createSurface := func(instance vk.Instance) vk.Surface {
						var surface vk.Surface
						result := vk.CreateWindowSurface(instance, event.Window.Ptr(), nil, &surface)
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

				case app.NativeWindowDestroyed:
					vkActive = false
					vulkandraw.DestroyInOrder(&v, &s, &r, &b, &gfx)
				case app.NativeWindowRedrawNeeded:
					if vkActive {
						vulkandraw.DrawFrame(v, s, r)
					}
					a.NativeWindowRedrawDone()
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
