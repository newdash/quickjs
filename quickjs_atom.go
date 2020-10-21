package quickjs

/*
#cgo CFLAGS: -D_GNU_SOURCE
#cgo CFLAGS: -DCONFIG_BIGNUM
#cgo CFLAGS: -fno-asynchronous-unwind-tables
#cgo LDFLAGS: -lm -lpthread -ldl

#include "bridge.h"
*/
import "C"

type Atom struct {
	ctx *Context
	ref C.JSAtom
}

func (a Atom) Free() { C.JS_FreeAtom(a.ctx.ref, a.ref) }

func (a Atom) String() string {
	ptr := C.JS_AtomToCString(a.ctx.ref, a.ref)
	defer C.JS_FreeCString(a.ctx.ref, ptr)
	return C.GoString(ptr)
}

func (a Atom) Value() Value {
	return Value{ctx: a.ctx, ref: C.JS_AtomToValue(a.ctx.ref, a.ref)}
}
