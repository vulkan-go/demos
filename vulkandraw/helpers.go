package vulkandraw

import (
	"fmt"
	"log"
	"strings"
	"unsafe"

	vk "github.com/vulkan-go/vulkan"
)

func check(ret vk.Result, name string) bool {
	if err := vk.Error(ret); err != nil {
		log.Println("[WARN]", name, "failed with", err)
		return true
	}
	return false
}

func orPanic(err interface{}) {
	switch v := err.(type) {
	case error:
		if v != nil {
			panic(err)
		}
	case vk.Result:
		if err := vk.Error(v); err != nil {
			panic(err)
		}
	case bool:
		if !v {
			panic("condition failed: != true")
		}
	}
}

func orPanicWith(err interface{}, notes ...string) {
	getNotes := func() string {
		return strings.Join(notes, " ")
	}
	switch v := err.(type) {
	case error:
		if v != nil {
			if len(notes) > 0 {
				err = fmt.Errorf("%s: %s", err, getNotes())
			}
			panic(err)
		}
	case vk.Result:
		if err := vk.Error(v); err != nil {
			if len(notes) > 0 {
				err = fmt.Errorf("%s: %s", err, getNotes())
			}
			panic(err)
		}
	case bool:
		if !v {
			if len(notes) > 0 {
				err := fmt.Errorf("condition failed: %s", getNotes())
				panic(err)
			}
			panic("condition failed: != true")
		}
	}
}

func repackUint32(data []byte) []uint32 {
	buf := make([]uint32, len(data)/4)
	vk.Memcopy(unsafe.Pointer((*sliceHeader)(unsafe.Pointer(&buf)).Data), data)
	return buf
}

type sliceHeader struct {
	Data uintptr
	Len  int
	Cap  int
}
