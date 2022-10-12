all: project build 

build:
	ant debug

install:
	ant debug install

project:
    # command "android-project" require installation:
    # go get github.com/xlab/treeprint
    # go get github.com/jawher/mow.cli
    # go install github.com/xlab/android-go/cmd/android-project

	android-project update --target android-23 --name VulkanDraw --path .

project-n:
	android-project update --target android-N --name VulkanDraw --path .

clean:
	rm -rf libs obj bin gen jni/lib
	rm -f proguard-project.txt project.properties
	rm -f local.properties build.xml
