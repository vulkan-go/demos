package main

import (
	"github.com/vulkan-go/demos/vulkaninfo"
	vk "github.com/vulkan-go/vulkan"
	"github.com/xlab/catcher"
	"github.com/xlab/ios-go/app"
)

var appInfo = &vk.ApplicationInfo{
	SType:              vk.StructureTypeApplicationInfo,
	ApiVersion:         vk.MakeVersion(1, 0, 0),
	ApplicationVersion: vk.MakeVersion(1, 0, 0),
	PApplicationName:   "VulkanInfo\x00",
	PEngineName:        "vulkango.com\x00",
}

func main() {
	app.Main(func(a app.AppDelegate) {
		defer catcher.Catch(
			catcher.RecvLog(true),
			catcher.RecvDie(-1),
		)

		var vkDevice *vulkaninfo.VulkanDeviceInfo
		a.InitDone()
		for {
			select {
			case event := <-a.LifecycleEvents():
				switch event.Kind {
				case app.ViewDidLoad:
					err := vk.Init()
					orPanic(err)
					vkDevice, err = vulkaninfo.NewVulkanDevice(appInfo, event.View)
					orPanic(err)
					vulkaninfo.PrintInfo(vkDevice)
				case app.ApplicationWillTerminate:
					vkDevice.Destroy()
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
