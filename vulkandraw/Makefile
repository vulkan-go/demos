.PHONY: shaders

all:

shaders:
	# Obtain glslangValidator at https://github.com/google/shaderc
	# Mirror: https://github.com/vulkan-go/shaderc
	glslangValidator -s -V -o shaders/tri-vert.spv shaders/tri.vert
	glslangValidator -s -V -o shaders/tri-frag.spv shaders/tri.frag
	go get github.com/jteeuwen/go-bindata
	go-bindata -pkg main shaders/
