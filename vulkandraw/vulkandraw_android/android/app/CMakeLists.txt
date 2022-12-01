
# Mention the cmake version name.
cmake_minimum_required(VERSION 3.4.1)

project( VulkanDraw )

# Application
add_library(vulkandraw SHARED IMPORTED)

set_target_properties( vulkandraw
        PROPERTIES IMPORTED_LOCATION
        ${CMAKE_SOURCE_DIR}/src/main/jniLibs/${ANDROID_ABI}/libvulkandraw.so
)
