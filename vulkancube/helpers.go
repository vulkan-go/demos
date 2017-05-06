package vulkancube

import (
	"fmt"
	"runtime"

	vk "github.com/vulkan-go/vulkan"
)

type Unwind []func()

func (u Unwind) Add(cleanup func()) {
	u = append(u, cleanup)
}

func (u Unwind) Unwind() {
	for i := len(u) - 1; i >= 0; i-- {
		u[i]()
	}
}

func (u Unwind) Discard() {
	if len(u) > 0 {
		u = u[:0]
	}
}

func isError(ret vk.Result) bool {
	return ret != vk.Success
}

func orPanic(err error, finalizers ...func()) {
	if err != nil {
		for _, fn := range finalizers {
			fn()
		}
		panic(err)
	}
}

func checkErr(err *error) {
	if v := recover(); v != nil {
		*err = fmt.Errorf("%+v", v)
	}
}

func checkErrStack(err *error) {
	if v := recover(); v != nil {
		stack := make([]byte, 32*1024)
		n := runtime.Stack(stack, false)
		switch event := v.(type) {
		case error:
			*err = fmt.Errorf("%s\n%s", event.Error(), stack[:n])
		default:
			*err = fmt.Errorf("%+v %s", v, stack[:n])
		}
	}
}

type sliceHeader struct {
	Data uintptr
	Len  int
	Cap  int
}
