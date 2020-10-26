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
	"errors"
	"github.com/mitchellh/mapstructure"
	"math/big"
	"reflect"
	"unsafe"
)

const (
	JsTagFIRST            = -11 /* first negative tag */
	JsTagBigDECIMAL       = -11
	JsTagBigINT           = -10
	JsTagBigFLOAT         = -9
	JsTagSYMBOL           = -8
	JsTagSTRING           = -7
	JsTagMODULE           = -3 /* used internally */
	JsTagFunctionByteCode = -2 /* used internally */
	JsTagOBJECT           = -1
	JsTagINT              = 0
	JsTagBOOL             = 1
	JsTagNULL             = 2
	JsTagUNDEFINED        = 3
	JsTagUNINITIALIZED    = 4
	JsTagCatchOffset      = 5
	JsTagEXCEPTION        = 6
	JsTagFLOAT64          = 7
)

type Value struct {
	ctx *Context
	ref C.JSValue
}

func (v Value) Free() {
	if !IsUndefinedOrNull(v.ref) && GetRefCount(v.ctx.ref, v.ref) > 0 {
		C.JS_FreeValue(v.ctx.ref, v.ref)
	}
}

func (v Value) Context() *Context { return v.ctx }

func (v Value) Bool() bool { return C.JS_ToBool(v.ctx.ref, v.ref) == 1 }

func (v Value) String() string {
	ptr := C.JS_ToCString(v.ctx.ref, v.ref)
	defer C.JS_FreeCString(v.ctx.ref, ptr)
	return C.GoString(ptr)
}

// Call function without `this`
func (v Value) Call(args ...Value) Value {
	return v.CallWithContext(v.ctx.Undefined(), args...)
}

// New for constructor
func (v Value) New(args ...Value) Value {

	if len(args) > 0 {
		var jsArgs []C.JSValue
		for _, goArg := range args {
			jsArgs = append(jsArgs, goArg.ref)
		}

		return Value{
			ctx: v.ctx,
			ref: C.JS_CallConstructor(v.ctx.ref, v.ref, C.int(len(args)), &jsArgs[0]),
		}
	}

	return Value{
		ctx: v.ctx,
		ref: C.JS_CallConstructor(v.ctx.ref, v.ref, C.int(len(args)), nil),
	}

}

// CallWithContext call function with this parameter
// if client call a function from class instance/object, please remember set the `this` object
func (v Value) CallWithContext(thisArg Value, args ...Value) Value {
	if len(args) > 0 {
		var jsArgs []C.JSValue
		for _, goArg := range args {
			jsArgs = append(jsArgs, goArg.ref)
		}

		return Value{
			ctx: v.ctx,
			ref: C.JS_Call(v.ctx.ref, v.ref, thisArg.ref, C.int(len(args)), &jsArgs[0]),
		}
	}

	return Value{
		ctx: v.ctx,
		ref: C.JS_Call(v.ctx.ref, v.ref, thisArg.ref, C.int(len(args)), nil),
	}
}

func (v Value) Int64() int64 {
	val := C.int64_t(0)
	C.JS_ToInt64(v.ctx.ref, &val, v.ref)
	return int64(val)
}

func (v Value) Int32() int32 {
	val := C.int32_t(0)
	C.JS_ToInt32(v.ctx.ref, &val, v.ref)
	return int32(val)
}

func (v Value) Uint32() uint32 {
	val := C.uint32_t(0)
	C.JS_ToUint32(v.ctx.ref, &val, v.ref)
	return uint32(val)
}

func (v Value) Float64() float64 {
	val := C.double(0)
	C.JS_ToFloat64(v.ctx.ref, &val, v.ref)
	return float64(val)
}

func (v Value) BigInt() *big.Int {
	if !v.IsBigInt() {
		return nil
	}
	val, ok := new(big.Int).SetString(v.String(), 10)
	if !ok {
		return nil
	}
	return val
}

func (v Value) BigFloat() *big.Float {
	if !v.IsBigDecimal() && !v.IsBigFloat() {
		return nil
	}
	val, ok := new(big.Float).SetString(v.String())
	if !ok {
		return nil
	}
	return val
}

func (v Value) Get(name string) Value {
	nameAtom := v.ctx.Atom(name)
	defer nameAtom.Free()
	return v.GetByAtom(nameAtom)
}

func (v Value) GetByAtom(atom Atom) Value {
	return v.ctx.newValue(C.JS_GetProperty(v.ctx.ref, v.ref, atom.ref))
}

func (v Value) GetByUint32(idx uint32) Value {
	return v.ctx.newValue(C.JS_GetPropertyUint32(v.ctx.ref, v.ref, C.uint32_t(idx)))
}

func (v Value) SetByAtom(atom Atom, val Value) {
	C.JS_SetProperty(v.ctx.ref, v.ref, atom.ref, val.ref)
}

func (v Value) SetByInt64(idx int64, val Value) {
	C.JS_SetPropertyInt64(v.ctx.ref, v.ref, C.int64_t(idx), val.ref)
}

func (v Value) SetByUint32(idx uint32, val Value) {
	C.JS_SetPropertyUint32(v.ctx.ref, v.ref, C.uint32_t(idx), val.ref)
}

func (v Value) Len() int64 { return v.Get("length").Int64() }

func (v Value) Set(name string, val Value) {
	nameAtom := v.ctx.Atom(name)
	defer nameAtom.Free()
	v.SetByAtom(nameAtom, val)
}

// SetGoValue with set value with go object
func (v Value) SetGoValue(name string, value interface{}) {
	namePtr := C.CString(name)
	defer C.free(unsafe.Pointer(namePtr))
	val := v.ctx.ToJSValue(value)
	v.Set(name, val)
}

// HasProperty with name
func (v Value) HasProperty(name string) bool {
	nameAtom := v.ctx.Atom(name)
	defer nameAtom.Free()
	return C.JS_HasProperty(v.ctx.ref, v.ref, nameAtom.ref) == 1
}

// DeleteProperty property
func (v Value) DeleteProperty(name string) {
	nameAtom := v.ctx.Atom(name)
	defer nameAtom.Free()
	C.JS_DeleteProperty(v.ctx.ref, v.ref, nameAtom.ref, C.int(0))
}

func (v Value) SetFunction(name string, fn JSFunction) {
	v.Set(name, v.ctx.Function(fn))
}

type Error struct {
	Cause string
	Stack string
}

func (err Error) Error() string { return err.Cause }

func (v Value) Error() error {
	if !v.IsError() {
		return nil
	}
	cause := v.String()

	stack := v.Get("stack")
	defer stack.Free()

	if stack.IsUndefined() {
		return &Error{Cause: cause}
	}
	return &Error{Cause: cause, Stack: stack.String()}
}
func IsUndefinedOrNull(ref C.JSValue) bool {
	return ref.tag == JsTagNULL || ref.tag == JsTagUNDEFINED
}

func (v Value) getTag() C.int64_t { return v.ref.tag }

func (v Value) IsNumber() bool        { return v.IsIntNumber() || v.IsFloat64Number() }
func (v Value) IsIntNumber() bool     { return v.getTag() == JsTagINT }
func (v Value) IsFloat64Number() bool { return v.ref.tag == JsTagFLOAT64 }
func (v Value) IsBigInt() bool        { return C.JS_IsBigInt(v.ctx.ref, v.ref) == 1 }
func (v Value) IsBigFloat() bool      { return C.JS_IsBigFloat(v.ref) == 1 }
func (v Value) IsBigDecimal() bool    { return C.JS_IsBigDecimal(v.ref) == 1 }
func (v Value) IsBool() bool          { return v.getTag() == JsTagBOOL }
func (v Value) IsNull() bool          { return v.getTag() == JsTagNULL }
func (v Value) IsUndefined() bool     { return v.getTag() == JsTagUNDEFINED }
func (v Value) IsException() bool     { return v.getTag() == JsTagEXCEPTION }
func (v Value) IsUninitialized() bool { return v.getTag() == JsTagUNINITIALIZED }
func (v Value) IsString() bool        { return v.getTag() == JsTagSTRING }
func (v Value) IsSymbol() bool        { return v.getTag() == JsTagSYMBOL }
func (v Value) IsObject() bool {
	if v.getTag() == JsTagOBJECT {
		if v.IsFunction() {
			return false
		}
		return true
	}
	return false
}
func (v Value) IsArray() bool       { return C.JS_IsArray(v.ctx.ref, v.ref) == 1 }
func (v Value) IsError() bool       { return C.JS_IsError(v.ctx.ref, v.ref) == 1 }
func (v Value) IsFunction() bool    { return C.JS_IsFunction(v.ctx.ref, v.ref) == 1 }
func (v Value) IsConstructor() bool { return C.JS_IsConstructor(v.ctx.ref, v.ref) == 1 }

type PropertyEnum struct {
	IsEnumerable bool
	Atom         Atom
}

func (p PropertyEnum) String() string { return p.Atom.String() }

// Decode js Value to golang struct
func (v Value) Decode(target interface{}) {
	mapstructure.Decode(v.Interface(), target)
}

// Interface return golang value with correct type (with interface{} any type)
func (v Value) Interface() interface{} {
	if v.IsNumber() {
		if v.IsBigInt() {
			return v.BigInt()
		}
		if v.IsBigFloat() {
			return v.BigFloat()
		}
		if v.IsBigDecimal() {
			return v.BigFloat()
		}
		if v.IsIntNumber() {
			return v.Int64()
		}
		return v.Float64()
	}
	if v.IsString() {
		return v.String()
	}
	if v.IsUndefined() || v.IsNull() {
		return nil
	}
	if v.IsBool() {
		return v.Bool()
	}
	if v.IsError() {
		return v.Error()
	}

	if v.IsArray() {
		var rt []interface{}
		arrayLen := v.Len()
		for idx := int64(0); idx < arrayLen; idx++ {
			rt = append(rt, v.GetByUint32(uint32(idx)).Interface())
		}
		return rt
	}

	// function also will be an object, just return function firstly
	if v.IsFunction() {
		return func(args ...interface{}) interface{} {
			var jsArgs []Value
			for arg := range args {
				jsArgs = append(jsArgs, v.ctx.ToJSValue(arg))
			}
			result := v.Call(jsArgs...)
			if result.IsException() {
				return v.ctx.Exception()
			}
			return result.Interface()
		}
	}

	if v.IsObject() {

		rt := map[string]interface{}{}
		if names, err := v.PropertyNames(); err == nil {
			for _, name := range names {
				propertyKey := name.String()
				propertyValue := v.GetByAtom(name.Atom)
				if !propertyValue.IsUndefined() {
					rt[propertyKey] = propertyValue.Interface()
				}
			}
		}
		return rt
	}

	return nil
}

// ToReflectValue used in native function call processing
// must provide a reflect.Type to check and return the reflect.Value instance
func (v Value) ToReflectValue(rType reflect.Type) reflect.Value {

	switch rType.Kind() {
	case reflect.Int64:
		return reflect.ValueOf(v.Int64())
	case reflect.Int32:
		return reflect.ValueOf(v.Int32())
	case reflect.Int16:
		return reflect.ValueOf(int16(v.Int64()))
	case reflect.Int8:
		return reflect.ValueOf(int8(v.Int64()))
	case reflect.Int:
		return reflect.ValueOf(int(v.Int64()))
	case reflect.Float32:
		return reflect.ValueOf(float32(v.Float64()))
	case reflect.Float64:
		return reflect.ValueOf(v.Float64())
	case reflect.Struct, reflect.Slice:
		reflectInstance := reflect.New(rType)
		instance := reflectInstance.Interface()
		v.Decode(instance)
		return reflect.Indirect(reflect.ValueOf(instance))
	}

	return reflect.ValueOf(v.Interface())

}

type JsType = string

const (
	JsTypeString    JsType = "string"
	JsTypeNumber    JsType = "number"
	JsTypeObject    JsType = "object"
	JsTypeSymbol    JsType = "symbol"
	JsTypeUndefined JsType = "undefined"
	JsTypeBigInt    JsType = "bigint"
	JsTypeFunction  JsType = "function"
	JsTypeBoolean   JsType = "boolean"
)

// TypeOf current value, same as the `typeof` keyword in js
func (v Value) TypeOf() JsType {
	if v.IsNull() {
		return JsTypeObject
	}
	if v.IsFunction() {
		return JsTypeFunction
	}
	if v.IsNumber() {
		if v.IsBigInt() {
			return JsTypeBigInt
		}
		return JsTypeNumber
	}
	if v.IsBool() {
		return JsTypeBoolean
	}
	if v.IsString() {
		return JsTypeString
	}
	if v.IsObject() {
		return JsTypeObject
	}
	if v.IsSymbol() {
		return JsTypeSymbol
	}
	return JsTypeUndefined
}

func (v Value) ToJsonString() string {
	undefined := v.ctx.Undefined()
	defer undefined.Free()

	jsonStr := Value{
		ref: C.JS_JSONStringify(v.ctx.ref, v.ref, undefined.ref, undefined.ref),
		ctx: v.ctx,
	}
	defer jsonStr.Free()
	return jsonStr.String()
}

// Dup value instance avoid freed by quickjs
// so user MUST manually free it
func (v Value) Dup() Value {
	return Value{
		ctx: v.ctx,
		ref: C.JS_DupValue(v.ctx.ref, v.ref),
	}
}

// PropertyNames of object, includes prototype
func (v Value) PropertyNames() ([]PropertyEnum, error) {
	var (
		ptr  *C.JSPropertyEnum
		size C.uint32_t
	)

	result := int(C.JS_GetOwnPropertyNames(v.ctx.ref, &ptr, &size, v.ref, C.int(1<<0|1<<1|1<<2)))
	if result < 0 {
		return nil, errors.New("value does not contain properties")
	}
	defer C.js_free(v.ctx.ref, unsafe.Pointer(ptr))

	entries := (*[(1 << 29) - 1]C.JSPropertyEnum)(unsafe.Pointer(ptr))

	names := make([]PropertyEnum, uint32(size))

	for i := 0; i < len(names); i++ {
		names[i].IsEnumerable = entries[i].is_enumerable == 1

		names[i].Atom = Atom{ctx: v.ctx, ref: entries[i].atom}
		names[i].Atom.Free()
	}

	return names, nil
}
