package main

import (
	"github.com/vulkan-go/demos/vulkaninfo"
	vk "github.com/vulkan-go/vulkan"
)

var appInfo = &vk.ApplicationInfo{
	SType:              vk.StructureTypeApplicationInfo,
	ApiVersion:         vk.MakeVersion(1, 0, 0),
	ApplicationVersion: vk.MakeVersion(1, 0, 0),
	PApplicationName:   "VulkanInfo\x00",
	PEngineName:        "vulkango.com\x00",
}

func main() {
	orPanic(vk.Init())
	vkDevice, err := vulkaninfo.NewVulkanDevice(appInfo, 0)
	orPanic(err)
	vulkaninfo.PrintInfo(vkDevice)
	vkDevice.Destroy()
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
