# Golang Vulkan API Demos

- vulcaninfo
- vulcandraw
- vulcancube

Using CGO then requare C/C++ compiler on path.

## Supported platforms

- Windows (GLFW)
- Android (Native)
- Linux graphics (GLFW)
- Linux computing (X11)
- OS X / macOS (GLFW + MoltenVK)
- iOS (Metal + MoltenVK)

## How to run on desktops

### GLFW v3.3

A standard `go run main.go` / `go build` in each of the glfw demo folders should work out of the box for most platforms.
Optionally make file can help.

For macOS / iOS the [MoltenVK](https://github.com/KhronosGroup/MoltenVK) dylib needs to be installed.  
On macOS you can use [Homebrew](https://docs.brew.sh/Installation) and install `molten-vk` package like `brew install molten-vk`.

### Manual configuration

For **OS X / macOS** you'll need to install.  the latest GLFW 3.3 from master https://github.com/glfw/glfw
and prepare MoltenVK https://moltengl.com/moltenvk/ SDK beforehand so CMake could find it.

There is a Makefile https://github.com/vulkan-go/demos/blob/master/vulkancube/vulkancube_desktop/Makefile to show how to properly invoke `go install` specifying the path to GLFW.

Make sure your graphics card and driver are supported:

- https://developer.nvidia.com/vulkan-driver
- http://www.amd.com/en-us/innovations/software-technologies/technologies-gaming/vulkan

In all cases you will run `XXX_desktop` demos.

## How to run on Android

Prerequisites are
- installed [Android SDK](https://developer.android.com/studio/releases/platform-tools)
- installed [Android NDK](https://developer.android.com/ndk/downloads)
- installed application "make"
- set system variable [ANDROID_HOME](https://developer.android.com/studio/command-line/variables)
- set system variable [NDK](https://developer.android.com/ndk/guides/other_build_systems)
- set system variable [HOST_TAG](https://developer.android.com/ndk/guides/other_build_systems)

Recommended:
- installed [Android Studio](https://developer.android.com/studio)
- [validation layers binaries](https://github.com/KhronosGroup/Vulkan-ValidationLayers/releases)

In the "android" folder is "make" file which run as `make all` will clean, build and make application APK file in ./android/app/build/outputs/apk/debug.  
Using Android Studio is very easy to deploy APK file into physical device or emulator.

Refer to [xlab/android-go/examples/minimal](https://github.com/xlab/android-go/tree/master/examples/minimal)

## License

[WTFPL](LICENSE.txt)
