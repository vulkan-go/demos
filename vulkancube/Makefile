.PHONY: shaders

all:

shaders:
	# Obtain glslangValidator at https://github.com/google/shaderc
	# Mirror: https://github.com/vulkan-go/shaderc
	glslangValidator -s -V -o shaders/cube.vert.spv shaders/cube.vert
	glslangValidator -s -V -o shaders/cube.frag.spv shaders/cube.frag
	go get github.com/jteeuwen/go-bindata
	go-bindata -pkg vulkancube shaders/ textures/
