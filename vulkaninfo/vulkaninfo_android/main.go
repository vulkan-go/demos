package main

import (
	"github.com/vulkan-go/demos/vulkaninfo"
	vk "github.com/vulkan-go/vulkan"
	"github.com/xlab/android-go/app"
	"github.com/xlab/catcher"
)

func init() {
	app.SetLogTag("VulkanInfo")
}

var appInfo = &vk.ApplicationInfo{
	SType:              vk.StructureTypeApplicationInfo,
	ApiVersion:         vk.MakeVersion(1, 0, 0),
	ApplicationVersion: vk.MakeVersion(1, 0, 0),
	PApplicationName:   "VulkanInfo\x00",
	PEngineName:        "vulkango.com\x00",
}

func main() {
	nativeWindowEvents := make(chan app.NativeWindowEvent)

	app.Main(func(a app.NativeActivity) {
		defer catcher.Catch(
			catcher.RecvLog(true),
			catcher.RecvDie(-1),
		)

		var vkDevice *vulkaninfo.VulkanDeviceInfo
		a.HandleNativeWindowEvents(nativeWindowEvents)
		a.InitDone()
		for {
			select {
			case <-a.LifecycleEvents():
				// ignore
			case event := <-nativeWindowEvents:
				switch event.Kind {
				case app.NativeWindowCreated:
					err := vk.Init()
					orPanic(err)
					vkDevice, err = vulkaninfo.NewVulkanDevice(appInfo, event.Window.Ptr())
					orPanic(err)
					vulkaninfo.PrintInfo(vkDevice)
				case app.NativeWindowDestroyed:
					vkDevice.Destroy()
				case app.NativeWindowRedrawNeeded:
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
