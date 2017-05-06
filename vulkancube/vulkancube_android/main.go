package main

import (
	"log"
	"time"

	as "github.com/vulkan-go/asche"
	"github.com/vulkan-go/demos/vulkancube"
	vk "github.com/vulkan-go/vulkan"
	"github.com/xlab/android-go/android"
	"github.com/xlab/android-go/app"
)

func init() {
	app.SetLogTag("VulkanCube")
}

type Application struct {
	*vulkancube.SpinningCube
	debugEnabled bool
	windowHandle uintptr
}

func (a *Application) VulkanSurface(instance vk.Instance) (surface vk.Surface) {
	ret := vk.CreateWindowSurface(instance, a.windowHandle, nil, &surface)
	if err := vk.Error(ret); err != nil {
		log.Println("vulkan error:", err)
		return vk.NullSurface
	}
	return surface
}

func (a *Application) VulkanAppName() string {
	return "VulkanInfo"
}

func (a *Application) VulkanLayers() []string {
	return []string{
		"VK_LAYER_GOOGLE_threading",
		"VK_LAYER_LUNARG_parameter_validation",
		"VK_LAYER_LUNARG_object_tracker",
		"VK_LAYER_LUNARG_core_validation",
		"VK_LAYER_LUNARG_api_dump",
		// "VK_LAYER_LUNARG_image",
		"VK_LAYER_LUNARG_swapchain",
		"VK_LAYER_GOOGLE_unique_objects",
	}
}

func (a *Application) VulkanDebug() bool {
	return a.debugEnabled
}

func (a *Application) VulkanSwapchainDimensions() *as.SwapchainDimensions {
	return &as.SwapchainDimensions{
		Width: 640, Height: 480, Format: vk.FormatB8g8r8a8Unorm,
	}
}

func (a *Application) VulkanDeviceExtensions() []string {
	return []string{
		"VK_KHR_swapchain",
	}
}

func (a *Application) VulkanInstanceExtensions() []string {
	extensions := vk.GetRequiredInstanceExtensions()
	if a.debugEnabled {
		extensions = append(extensions, "VK_EXT_debug_report")
	}
	return extensions
}

func NewApplication(debugEnabled bool) *Application {
	return &Application{
		SpinningCube: vulkancube.NewSpinningCube(1.0),

		debugEnabled: debugEnabled,
	}
}

func main() {
	nativeWindowEvents := make(chan app.NativeWindowEvent)
	inputQueueEvents := make(chan app.InputQueueEvent, 1)
	inputQueueChan := make(chan *android.InputQueue, 1)

	app.Main(func(a app.NativeActivity) {
		// disable this to get the stack
		// defer catcher.Catch(
		// 	catcher.RecvLog(true),
		// 	catcher.RecvDie(-1),
		// )

		orPanic(vk.Init())
		a.HandleNativeWindowEvents(nativeWindowEvents)
		a.HandleInputQueueEvents(inputQueueEvents)
		// just skip input events (so app won't be dead on touch input)
		go app.HandleInputQueues(inputQueueChan, func() {
			a.InputQueueHandled()
		}, app.SkipInputEvents)
		a.InitDone()

		var (
			cubeApp  *Application
			platform as.Platform
			ctx      as.Context
			err      error
			vkActive bool

			lastRender time.Duration
		)

		fpsDelay := time.Second / 60
		fpsTicker := time.NewTicker(fpsDelay)
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
					cubeApp = NewApplication(true)
					cubeApp.windowHandle = event.Window.Ptr()
					// creates a new platform, also initializes Vulkan context in the cubeApp
					platform, err = as.NewPlatform(cubeApp)
					orPanic(err)
					ctx = cubeApp.Context()
					dim := ctx.SwapchainDimensions()
					log.Printf("Initialized %s with %+v swapchain", cubeApp.VulkanAppName(), dim)
					vkActive = true
				case app.NativeWindowDestroyed:
					cubeApp.Destroy()
					platform.Destroy()
					vkActive = false
				case app.NativeWindowRedrawNeeded:
					a.NativeWindowRedrawDone()
				}
			case <-fpsTicker.C:
				if vkActive {
					cubeApp.NextFrame()

					// https://source.android.com/devices/graphics/arch-gameloops
					// FPS may drop down when no interacton with the app, should skip frames there.
					// TODO: use VK_GOOGLE_display_timing_enabled as cool guys would do. Don't be an uncool fool.
					if lastRender > fpsDelay {
						// skip frame
						lastRender = lastRender - fpsDelay
						continue
					}
					ts := time.Now()

					imageIdx, outdated, err := ctx.AcquireNextImage()
					orPanic(err)
					if outdated {
						imageIdx, _, err = ctx.AcquireNextImage()
						orPanic(err)
					}
					_ = imageIdx
					_, err = ctx.PresentImage(imageIdx)
					orPanic(err)

					lastRender = time.Since(ts)
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
