HOST_TAG ?= linux-x86_64
API_LEVEL ?= 26
TOOLCHAIN ?= $(NDK)/toolchains/llvm/prebuilt/$(HOST_TAG)/bin

all: clean build apk
	echo "WORK DONE"

build: build32 build64 build-amd64
	echo "builded all"

build32:
	echo "compile for old phones with arm (32bit)"
	CC="$(TOOLCHAIN)/armv7a-linux-androideabi$(API_LEVEL)-clang" \
	CXX="$(TOOLCHAIN)/armv7a-linux-androideabi$(API_LEVEL)-clang++" \
	GOOS=android \
	GOARCH=arm \
	CGO_ENABLED=1 \
	go build -buildmode=c-shared -o android/app/src/main/jniLibs/armeabi-v7a/libvulkandraw.so

build64:
	echo "compile for arm64 (64bit)"
	CC="$(TOOLCHAIN)/aarch64-linux-android$(API_LEVEL)-clang" \
	CXX="$(TOOLCHAIN)/aarch64-linux-android$(API_LEVEL)-clang++" \
	GOOS=android \
	GOARCH=arm64 \
	CGO_ENABLED=1 \
	go build -buildmode=c-shared -o android/app/src/main/jniLibs/arm64-v8a/libvulkandraw.so

build-amd64:
	echo "compile for emulator"
	CC="$(TOOLCHAIN)/x86_64-linux-android$(API_LEVEL)-clang" \
	CXX="$(TOOLCHAIN)/x86_64-linux-android$(API_LEVEL)-clang++" \
	GOOS=android \
	GOARCH=amd64 \
	CGO_ENABLED=1 \
	go build -buildmode=c-shared -o android/app/src/main/jniLibs/x86_64/libvulkandraw.so

clean:
	echo "delete generated files"
	rm -rf android/app/build
	rm -rf android/app/build-native

apk:
	echo "run gradle"
	cd android; ./gradlew -q clean assembleDebug > /dev/null 2>&1

listen:
	adb logcat -c
	adb logcat *:S VulkanDraw
