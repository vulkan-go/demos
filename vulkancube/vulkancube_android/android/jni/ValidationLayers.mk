LOCAL_PATH := $(call my-dir)

include $(CLEAR_VARS)

LOCAL_MODULE    := VK_LAYER_LUNARG_api_dump
LOCAL_SRC_FILES := lib/libVkLayer_api_dump.so

include $(PREBUILT_SHARED_LIBRARY)

include $(CLEAR_VARS)

LOCAL_MODULE    := VK_LAYER_LUNARG_core_validation
LOCAL_SRC_FILES := lib/libVkLayer_core_validation.so

include $(PREBUILT_SHARED_LIBRARY)

include $(CLEAR_VARS)

LOCAL_MODULE    := VK_LAYER_LUNARG_object_tracker
LOCAL_SRC_FILES := lib/libVkLayer_object_tracker.so

include $(PREBUILT_SHARED_LIBRARY)

include $(CLEAR_VARS)

LOCAL_MODULE    := VK_LAYER_LUNARG_parameter_validation
LOCAL_SRC_FILES := lib/libVkLayer_parameter_validation.so

include $(PREBUILT_SHARED_LIBRARY)

include $(CLEAR_VARS)

LOCAL_MODULE    := VK_LAYER_LUNARG_screenshot
LOCAL_SRC_FILES := lib/libVkLayer_screenshot.so

include $(PREBUILT_SHARED_LIBRARY)

include $(CLEAR_VARS)

LOCAL_MODULE    := VK_LAYER_LUNARG_swapchain
LOCAL_SRC_FILES := lib/libVkLayer_swapchain.so

include $(PREBUILT_SHARED_LIBRARY)

include $(CLEAR_VARS)

LOCAL_MODULE    := VK_LAYER_GOOGLE_threading
LOCAL_SRC_FILES := lib/libVkLayer_threading.so

include $(PREBUILT_SHARED_LIBRARY)

include $(CLEAR_VARS)

LOCAL_MODULE    := VK_LAYER_GOOGLE_unique_objects
LOCAL_SRC_FILES := lib/libVkLayer_unique_objects.so

include $(PREBUILT_SHARED_LIBRARY)
