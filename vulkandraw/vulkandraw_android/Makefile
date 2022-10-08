OS=linux
ANDROID_API ?= 23
ANDROID_TOOLCHAIN_DIR ?= $(ANDROID_NDK_HOME)/toolchains/llvm/prebuilt/$(OS)-x86_64/bin

all: build apk

build:
	mkdir -p android/jni/lib
	CC="$(ANDROID_TOOLCHAIN_DIR)/aarch64-linux-android$(ANDROID_API)-clang" \
	CXX="$(ANDROID_TOOLCHAIN_DIR)/aarch64-linux-android$(ANDROID_API)-clang++" \
	GOOS=android \
	GOARCH=arm64 \
	CGO_ENABLED=1 \
	go build -buildmode=c-shared -o android/jni/lib/libvulkandraw.so
 
apk:
	cd android && make

clean:
	cd android && make clean

install:
	cd android && make install

listen:
	adb logcat -c
	adb logcat *:S VulkanDraw
