package quickjs

/*
#cgo CFLAGS: -D_GNU_SOURCE
#cgo CFLAGS: -DCONFIG_BIGNUM
#cgo CFLAGS: -fno-asynchronous-unwind-tables
#cgo LDFLAGS: -lm -lpthread

#include "bridge.h"
*/
import "C"
import (
	"fmt"
	"io"
	"reflect"
	"unsafe"
)

type Context struct {
	ref     *C.JSContext
	globals *Value
	proxy   *Value
	runtime *Runtime
}

func (ctx *Context) Free() {
	if ctx.proxy != nil {
		ctx.proxy.Free()
	}
	if ctx.globals != nil {
		ctx.globals.Free()
	}

	C.JS_FreeContext(ctx.ref)
}

func (ctx *Context) Function(fn JSFunction) Value {
	val := ctx.eval(`(proxy, id) => function() { return proxy.call(this, id, ...arguments); }`)
	if val.IsException() {
		return val
	}
	defer val.Free()

	funcPtr := storeFuncPtr(funcEntry{ctx: ctx, fn: fn})
	funcPtrVal := ctx.Int64(funcPtr)

	if ctx.proxy == nil {
		ctx.proxy = &Value{
			ctx: ctx,
			ref: C.JS_NewCFunction(ctx.ref, (*C.JSCFunction)(unsafe.Pointer(C.InvokeProxy)), nil, C.int(0)),
		}
	}

	args := []C.JSValue{ctx.proxy.ref, funcPtrVal.ref}

	return Value{ctx: ctx, ref: C.JS_Call(ctx.ref, val.ref, ctx.Null().ref, C.int(len(args)), &args[0])}
}

func (ctx *Context) Null() Value {
	return Value{ctx: ctx, ref: C.JS_NewNull()}
}

func (ctx *Context) Undefined() Value {
	return Value{ctx: ctx, ref: C.JS_NewUndefined()}
}

func (ctx *Context) Uninitialized() Value {
	return Value{ctx: ctx, ref: C.JS_NewUninitialized()}
}

func (ctx *Context) Error(err error) Value {
	val := Value{ctx: ctx, ref: C.JS_NewError(ctx.ref)}
	val.Set("message", ctx.String(err.Error()))
	return val
}

func (ctx *Context) Bool(b bool) Value {
	bv := 0
	if b {
		bv = 1
	}
	return Value{ctx: ctx, ref: C.JS_NewBool(ctx.ref, C.int(bv))}
}

func (ctx *Context) Int32(v int32) Value {
	return Value{ctx: ctx, ref: C.JS_NewInt32(ctx.ref, C.int32_t(v))}
}

func (ctx *Context) Int64(v int64) Value {
	return Value{ctx: ctx, ref: C.JS_NewInt64(ctx.ref, C.int64_t(v))}
}

func (ctx *Context) Uint32(v uint32) Value {
	return Value{ctx: ctx, ref: C.JS_NewUint32(ctx.ref, C.uint32_t(v))}
}

func (ctx *Context) BigUint64(v uint64) Value {
	return Value{ctx: ctx, ref: C.JS_NewBigUint64(ctx.ref, C.uint64_t(v))}
}

func (ctx *Context) Float64(v float64) Value {
	return Value{ctx: ctx, ref: C.JS_NewFloat64(ctx.ref, C.double(v))}
}

func (ctx *Context) String(v string) Value {
	ptr := C.CString(v)
	defer C.free(unsafe.Pointer(ptr))
	return Value{ctx: ctx, ref: C.JS_NewString(ctx.ref, ptr)}
}

func (ctx *Context) Atom(v string) Atom {
	ptr := C.CString(v)
	defer C.free(unsafe.Pointer(ptr))
	return Atom{ctx: ctx, ref: C.JS_NewAtom(ctx.ref, ptr)}
}

func (ctx *Context) eval(code string) Value { return ctx.evalFile(code, "code", 0) }

func (ctx *Context) evalFile(code, filename string, mod int) Value {
	codePtr := C.CString(code)
	defer C.free(unsafe.Pointer(codePtr))

	filenamePtr := C.CString(filename)
	defer C.free(unsafe.Pointer(filenamePtr))

	return Value{ctx: ctx, ref: C.JS_Eval(ctx.ref, codePtr, C.size_t(len(code)), filenamePtr, C.int(mod))}
}

func (ctx *Context) EvalModule(code string) (Value, error) { return ctx.EvalFile(code, "code", 1) }

func (ctx *Context) EvalGlobal(code string) (Value, error) { return ctx.EvalFile(code, "code", 0) }

func (ctx *Context) EvalFile(code, filename string, mod int) (Value, error) {
	val := ctx.evalFile(code, filename, mod)
	if val.IsException() {
		return val, ctx.Exception()
	}
	return val, nil
}

func (ctx *Context) Globals() Value {
	if ctx.globals == nil {
		ctx.globals = &Value{
			ctx: ctx,
			ref: C.JS_GetGlobalObject(ctx.ref),
		}
	}
	return *ctx.globals
}

func (ctx *Context) Throw(v Value) Value {
	return Value{ctx: ctx, ref: C.JS_Throw(ctx.ref, v.ref)}
}

func (ctx *Context) ThrowError(err error) Value { return ctx.Throw(ctx.Error(err)) }

func (ctx *Context) ThrowSyntaxError(format string, args ...interface{}) Value {
	cause := fmt.Sprintf(format, args...)
	causePtr := C.CString(cause)
	defer C.free(unsafe.Pointer(causePtr))
	return Value{ctx: ctx, ref: C.ThrowSyntaxError(ctx.ref, causePtr)}
}

func (ctx *Context) ThrowTypeError(format string, args ...interface{}) Value {
	cause := fmt.Sprintf(format, args...)
	causePtr := C.CString(cause)
	defer C.free(unsafe.Pointer(causePtr))
	return Value{ctx: ctx, ref: C.ThrowTypeError(ctx.ref, causePtr)}
}

func (ctx *Context) ThrowReferenceError(format string, args ...interface{}) Value {
	cause := fmt.Sprintf(format, args...)
	causePtr := C.CString(cause)
	defer C.free(unsafe.Pointer(causePtr))
	return Value{ctx: ctx, ref: C.ThrowReferenceError(ctx.ref, causePtr)}
}

func (ctx *Context) ThrowRangeError(format string, args ...interface{}) Value {
	cause := fmt.Sprintf(format, args...)
	causePtr := C.CString(cause)
	defer C.free(unsafe.Pointer(causePtr))
	return Value{ctx: ctx, ref: C.ThrowRangeError(ctx.ref, causePtr)}
}

func (ctx *Context) ThrowInternalError(format string, args ...interface{}) Value {
	cause := fmt.Sprintf(format, args...)
	causePtr := C.CString(cause)
	defer C.free(unsafe.Pointer(causePtr))
	return Value{ctx: ctx, ref: C.ThrowInternalError(ctx.ref, causePtr)}
}

func (ctx *Context) Exception() error {
	val := Value{ctx: ctx, ref: C.JS_GetException(ctx.ref)}
	defer val.Free()
	return val.Error()
}

// Object create new JSObject
func (ctx *Context) Object() Value {
	return Value{ctx: ctx, ref: C.JS_NewObject(ctx.ref)}
}

// ToJSValue convert golang object to quickjs.Value
func (ctx *Context) ToJSValue(value interface{}) Value {

	if value == nil {
		return ctx.Undefined()
	}

	reflectValue := reflect.ValueOf(value)
	reflectType := reflectValue.Type()

	// if it is quickjs.Value, return it directly
	if reflectType == reflect.TypeOf(Value{}) {
		return value.(Value)
	}
	// if is reflect.Value, unwrap the real value
	if reflectType == reflect.TypeOf(reflect.Value{}) {
		value := value.(reflect.Value).Interface()
		return ctx.ToJSValue(value)
	}

	switch reflectValue.Kind() {
	case reflect.String:
		return ctx.String(reflectValue.String())
	case reflect.Int8, reflect.Int, reflect.Int16, reflect.Int32:
		return ctx.Int32(int32(reflectValue.Int()))
	case reflect.Int64:
		return ctx.Int64(reflectValue.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return ctx.Uint32(uint32(reflectValue.Uint()))
	case reflect.Uint64:
		return ctx.BigUint64(reflectValue.Uint())
	case reflect.Float32, reflect.Float64:
		return ctx.Float64(reflectValue.Float())
	case reflect.Bool:
		return ctx.Bool(reflectValue.Bool())
	case reflect.Map:
		obj := ctx.Object()
		for _, key := range reflectValue.MapKeys() {
			innerValue := reflectValue.MapIndex(key)
			obj.Set(key.String(), ctx.ToJSValue(innerValue.Interface()))
		}
		return obj
	case reflect.Struct:
		obj := ctx.Object()
		for fIndex := 0; fIndex < reflectValue.NumField(); fIndex++ {
			field := reflectValue.Field(fIndex)
			fieldName := reflectType.Field(fIndex).Name
			if isExportedName(fieldName) {
				obj.Set(fieldName, ctx.ToJSValue(field.Interface()))
			}
		}
		methodCount := reflectType.NumMethod()
		for mIndex := 0; mIndex < methodCount; mIndex++ {
			method := reflectValue.Method(mIndex)
			methodName := reflectType.Method(mIndex).Name
			if isExportedName(methodName) {
				obj.Set(methodName, ctx.ToJSValue(method.Interface()))
			}
		}
		return obj
	case reflect.Slice:
		obj := ctx.Array()
		for arrayItemIndex := 0; arrayItemIndex < reflectValue.Len(); arrayItemIndex++ {
			arrayItem := reflectValue.Index(arrayItemIndex)
			arrayItemValue := ctx.ToJSValue(arrayItem.Interface())
			obj.SetByInt64(int64(arrayItemIndex), arrayItemValue)
		}
		return obj
	case reflect.Func:
		funcArgsNum := reflectType.NumIn()
		return ctx.Function(func(ctx *Context, this Value, jsArgs []Value) Value {
			if len(jsArgs) < funcArgsNum {
				return ctx.ThrowError(fmt.Errorf("arguments is not enough, the function require '%v' parameters", funcArgsNum))
			}

			var goFuncArgs []reflect.Value

			for i := 0; i < funcArgsNum; i++ {
				argType := reflectType.In(i)
				jsArg := jsArgs[i]
				goFuncArgs = append(goFuncArgs, jsArg.ToReflectValue(argType))
			}

			goFuncResult := reflectValue.Call(goFuncArgs)

			if len(goFuncResult) == 0 {
				return ctx.Undefined()
			} else if len(goFuncResult) == 1 {
				return ctx.ToJSValue(goFuncResult[0])
			} else {
				return ctx.ToJSValue(goFuncResult)
			}

			return ctx.ThrowError(fmt.Errorf("not support call golang function directly"))
		})
	default:
		// ignore
	}
	return ctx.Undefined()
}

// ParseJson parse Value from JSON string
func (ctx *Context) ParseJson(jsonStr string) Value {
	jsJsonString := ctx.ToJSValue(jsonStr)
	defer jsJsonString.Free()
	JSON := ctx.Globals().Get("JSON")
	return JSON.Get("parse").CallWithContext(JSON, jsJsonString)
}

// NewPromise shortcut for creating a new promise object
func (ctx *Context) NewPromise(runner PromiseRunner) Value {
	cb := ctx.Function(func(ctx *Context, this Value, args []Value) Value {
		runner(args[0], args[1])
		return ctx.Undefined()
	})
	return ctx.Globals().Get("Promise").New(cb)
}

func (ctx *Context) ExecutePendingJob() error {

	code := C.JS_ExecutePendingJob(ctx.runtime.ref, &ctx.ref)
	if code <= 0 {
		if code == 0 {
			return io.EOF
		}
		return ctx.Exception()
	}

	return nil
}

type PromiseRunner = func(resolve, reject Value)

func (ctx *Context) Array() Value {
	return Value{ctx: ctx, ref: C.JS_NewArray(ctx.ref)}
}
