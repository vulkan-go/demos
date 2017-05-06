GLFW_INCLUDE_DIR = "?"
GLFW_LIB_DIR = "?"

install:
	CGO_CFLAGS="-I$(GLFW_INCLUDE_DIR)" CGO_LDFLAGS="-L$(GLFW_LIB_DIR)" go install

install-pkg:
	CGO_CFLAGS="$(shell pkg-config --cflags glfw3)" CGO_LDFLAGS="$(shell pkg-config --libs glfw3)" go install

build:
	CGO_CFLAGS="-I$(GLFW_INCLUDE_DIR)" CGO_LDFLAGS="-L$(GLFW_LIB_DIR)" go build

build-pkg:
	CGO_CFLAGS="$(shell pkg-config --cflags glfw3)" CGO_LDFLAGS="$(shell pkg-config --libs glfw3)" go build
