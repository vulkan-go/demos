
# Mention the cmake version name.
cmake_minimum_required(VERSION 3.4.1)

project( VulkanInfo )

# Application
add_library(vulkaninfo SHARED IMPORTED)

set_target_properties( vulkaninfo
        PROPERTIES IMPORTED_LOCATION
        ${CMAKE_SOURCE_DIR}/src/main/jniLibs/${ANDROID_ABI}/libvulkaninfo.so
)
