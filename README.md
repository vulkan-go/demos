# Golang Vulkan API Demos

- vulcaninfo
- vulcandraw
- vulcancube

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

For macOS / iOS the MoltenVK.framework needs to be installed in the /Library/Frameworks folder for the build to find it. It can be downloaded as part of he Vulkan SDK: https://vulkan.lunarg.com/sdk/home

### Manual configuration

For **OS X / macOS** you'll need to install the latest GLFW 3.3 from master https://github.com/glfw/glfw
and prepare MoltenVK https://moltengl.com/moltenvk/ SDK beforehand so CMake could find it.

There is a Makefile https://github.com/vulkan-go/demos/blob/master/vulkancube/vulkancube_desktop/Makefile to show how to properly invoke `go install` specifying the path to GLFW.

For **Windows** you don't need to compile MoltenVK and can use GLFW 3.2.1 distro from the site http://www.glfw.org then just specify paths in Makefile or run commands by hand, specifying paths to GLFW distro folders.

Make sure your graphics card and driver are supported:

- https://developer.nvidia.com/vulkan-driver
- http://www.amd.com/en-us/innovations/software-technologies/technologies-gaming/vulkan

In all cases you will run `XXX_desktop` demos.

## How to run on Android

Refer to [xlab/android-go/examples/minimal](https://github.com/xlab/android-go/tree/master/examples/minimal)

## License

[WTFPL](LICENSE.txt)
