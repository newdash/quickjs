package quickjs

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

/*
#cgo CFLAGS: -D_GNU_SOURCE
#cgo CFLAGS: -DCONFIG_BIGNUM
#cgo CFLAGS: -fno-asynchronous-unwind-tables
#cgo LDFLAGS: -lm -lpthread

#include "bridge.h"
*/
import "C"

// Runtime for quickjs
type Runtime struct {
	ref *C.JSRuntime
}

// NewRuntime for javascript
func NewRuntime() Runtime {
	rt := Runtime{ref: C.JS_NewRuntime()}
	C.JS_SetCanBlock(rt.ref, C.int(1))
	return rt
}

// RunGC to perform garbage collection for runtime
func (r Runtime) RunGC() { C.JS_RunGC(r.ref) }

// Free runtime, it will raise error when assert failed when something is not free
func (r Runtime) Free() {
	C.JS_FreeRuntime(r.ref)
}

// SetMaxStackSize of Runtime in bytes
func (r Runtime) SetMaxStackSize(stackSize int64) {
	C.JS_SetMaxStackSize(r.ref, C.size_t(stackSize))
}

// SetMemoryLimit of Runtime
func (r Runtime) SetMemoryLimit(limit int64) {
	C.JS_SetMemoryLimit(r.ref, C.size_t(limit))
}

// NewContext for quickjs
func (r Runtime) NewContext() *Context {
	ref := C.JS_NewContext(r.ref)

	C.JS_AddIntrinsicBigFloat(ref)
	C.JS_AddIntrinsicBigDecimal(ref)
	C.JS_AddIntrinsicOperators(ref)
	C.JS_EnableBignumExt(ref, C.int(1))

	ctx := &Context{ref: ref, runtime: &r}

	return ctx
}

// JSFunction proxy
type JSFunction func(ctx *Context, this Value, args []Value) Value

type funcEntry struct {
	ctx *Context
	fn  JSFunction
}

var funcPtrLen int64
var funcPtrLock sync.Mutex
var funcPtrStore = make(map[int64]funcEntry)
var funcPtrClassID C.JSClassID

func init() { C.JS_NewClassID(&funcPtrClassID) }

func storeFuncPtr(v funcEntry) int64 {
	id := atomic.AddInt64(&funcPtrLen, 1) - 1
	funcPtrLock.Lock()
	defer funcPtrLock.Unlock()
	funcPtrStore[id] = v
	return id
}

func restoreFuncPtr(ptr int64) funcEntry {
	funcPtrLock.Lock()
	defer funcPtrLock.Unlock()
	return funcPtrStore[ptr]
}

//func freeFuncPtr(ptr int64) {
//	funcPtrLock.Lock()
//	defer funcPtrLock.Unlock()
//	delete(funcPtrStore, ptr)
//}

//export proxy
func proxy(ctx *C.JSContext, thisVal C.JSValueConst, argc C.int, argv *C.JSValueConst) C.JSValue {
	// The maximum capacity of the following two slices is limited to (2^29)-1 to remain compatible
	// with 32-bit platforms. The size of a `*C.char` (a pointer) is 4 Byte on a 32-bit system
	// and (2^29)*4 == math.MaxInt32 + 1. -- See issue golang/go#13656
	refs := (*[(1 << 29) - 1]C.JSValueConst)(unsafe.Pointer(argv))[:argc:argc]

	id := C.int64_t(0)
	C.JS_ToInt64(ctx, &id, refs[0])

	entry := restoreFuncPtr(int64(id))

	args := make([]Value, len(refs)-1)
	for i := 0; i < len(args); i++ {
		args[i].ctx = entry.ctx
		args[i].ref = refs[1+i]
	}

	result := entry.fn(entry.ctx, Value{ctx: entry.ctx, ref: thisVal}, args)

	return result.ref
}
