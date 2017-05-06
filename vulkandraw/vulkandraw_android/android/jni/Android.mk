LOCAL_PATH := $(call my-dir)

include $(CLEAR_VARS)

LOCAL_MODULE    := vulkandraw
LOCAL_SRC_FILES := lib/libvulkandraw.so
LOCAL_LDLIBS    := -llog -landroid

include $(PREBUILT_SHARED_LIBRARY)

# Enable Vulkan validation layers, you can obtain them at
# 	https://github.com/LunarG/VulkanTools
# 	mirror: https://github.com/vulkan-go/VulkanTools

# include $(LOCAL_PATH)/ValidationLayers.mk
