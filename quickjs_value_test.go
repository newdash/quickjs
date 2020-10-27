package quickjs

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	stdruntime "runtime"
	"testing"
)

func TestValue_ToJSONString(t *testing.T) {
	stdruntime.LockOSThread()
	defer stdruntime.UnlockOSThread()
	assert := assert.New(t)
	r := NewRuntime()
	defer r.Free()
	ctx := r.NewContext()
	defer ctx.Free()

	v := ctx.ToJSValue(map[string]string{"attachTimerTo": "1"})
	defer v.Free()
	assert.True(v.IsObject())
	assert.Equal(`{"attachTimerTo":"1"}`, v.ToJsonString())
}

func TestValue_ToJSONStringDeep(t *testing.T) {
	stdruntime.LockOSThread()
	defer stdruntime.UnlockOSThread()
	assert := assert.New(t)
	r := NewRuntime()
	defer r.Free()
	ctx := r.NewContext()
	defer ctx.Free()

	v := ctx.ToJSValue(map[string]interface{}{"v1": map[string]int64{"v2": 3}})
	defer v.Free()
	assert.True(v.IsObject())
	assert.Equal(`{"v1":{"v2":3}}`, v.ToJsonString())
}

func TestContext_CallFunction(t *testing.T) {
	stdruntime.LockOSThread()
	defer stdruntime.UnlockOSThread()
	assert := assert.New(t)

	r := NewRuntime()
	defer r.Free()
	ctx := r.NewContext()
	defer ctx.Free()
	f := ctx.Function(func(ctx *Context, this Value, args []Value) Value {
		return args[0]
	})
	ctx.Globals().Set("f", f)
	assert.True(ctx.Globals().HasProperty("f"))
	arg0 := ctx.ToJSValue(4444)
	result := f.CallWithContext(ctx.Null(), arg0)
	assert.False(result.IsError())
	assert.Equal(int64(4444), result.Interface())

	ctx.Globals().DeleteProperty("f")

	arg0.Free()
	result.Free()
	r.RunGC()
}

func TestContext_DynaCallFunction(t *testing.T) {
	stdruntime.LockOSThread()
	defer stdruntime.UnlockOSThread()
	assert := assert.New(t)

	r := NewRuntime()
	defer r.Free()
	ctx := r.NewContext()
	defer ctx.Free()
	f := ctx.Function(func(ctx *Context, this Value, args []Value) Value {
		return args[0].Dup()
	})
	ctx.Globals().Set("f", f)
	result := f.DynamicCall(4444, 1234)
	defer result.Free()
	assert.False(result.IsError())
	assert.Equal(int64(4444), result.Interface())

}

func TestContext_CallFunctionWithoutArgs(t *testing.T) {
	stdruntime.LockOSThread()
	defer stdruntime.UnlockOSThread()
	assert := assert.New(t)

	r := NewRuntime()
	defer r.RunGC()
	ctx := r.NewContext()
	defer ctx.Free()
	f := ctx.Function(func(ctx *Context, this Value, args []Value) Value {
		return ctx.Int64(1234)
	})
	ctx.Globals().Set("f", f)
	result, err := ctx.EvalGlobal("f()")
	assert.Nil(err)
	assert.Equal(int64(1234), result.Int64())

	result = f.Call()
	assert.Equal(int64(1234), result.Int64())
}

func TestContext_VerifyFunctionArgs(t *testing.T) {
	stdruntime.LockOSThread()
	defer stdruntime.UnlockOSThread()
	assert := assert.New(t)

	r := NewRuntime()
	defer r.RunGC()
	ctx := r.NewContext()
	defer ctx.Free()
	f := ctx.Function(func(ctx *Context, this Value, args []Value) Value {
		return ctx.ToJSValue(
			map[string]interface{}{
				"args":   args,
				"length": len(args),
			},
		)
	})
	ctx.Globals().Set("f", f)
	assert.True(ctx.Globals().HasProperty("f"))
	result, err := ctx.EvalGlobal("f(1234,2345)")
	assert.Nil(err)
	assert.Equal(int64(2), result.Get("length").Int64())
	assert.Equal(int64(1234), result.Get("args").GetByUint32(0).Int64())
	assert.Equal(int64(2345), result.Get("args").GetByUint32(1).Int64())

	result = f.Call(ctx.Int64(1234), ctx.Int64(2345))
	assert.Equal(int64(2), result.Get("length").Int64())
	assert.Equal(int64(1234), result.Get("args").GetByUint32(0).Int64())
	assert.Equal(int64(2345), result.Get("args").GetByUint32(1).Int64())

}

func TestValue_InterfaceNumber(t *testing.T) {
	stdruntime.LockOSThread()
	defer stdruntime.UnlockOSThread()
	assert := assert.New(t)

	r := NewRuntime()
	defer r.Free()
	defer r.RunGC()
	ctx := r.NewContext()
	defer ctx.Free()
	v, err := ctx.EvalGlobal("1")
	defer v.Free()
	assert.Nil(err)
	assert.True(v.IsNumber())
	assert.True(v.IsIntNumber())
	assert.False(v.IsFloat64Number())
	v2, err2 := ctx.EvalGlobal("1.1")
	defer v2.Free()
	assert.Nil(err2)
	assert.Nil(err)
	assert.True(v2.IsNumber())
	assert.False(v2.IsIntNumber())
	assert.True(v2.IsFloat64Number())

	goV := ctx.Int64(123)
	defer goV.Free()
	assert.True(goV.IsNumber())
	assert.True(goV.IsIntNumber())
	assert.False(goV.IsFloat64Number())
}

func TestValue_InterfaceArray(t *testing.T) {
	stdruntime.LockOSThread()
	defer stdruntime.UnlockOSThread()
	assert := assert.New(t)

	r := NewRuntime()
	defer r.Free()
	defer r.RunGC()
	ctx := r.NewContext()
	defer ctx.Free()
	v, err := ctx.EvalGlobal("[1,2,3,4.3]")
	defer v.Free()
	assert.Nil(err)
	assert.Equal([]interface{}{int64(1), int64(2), int64(3), 4.3}, v.Interface())
	assert.Equal(int64(4), v.Len())
}

func TestContext_EvalJson(t *testing.T) {
	stdruntime.LockOSThread()
	defer stdruntime.UnlockOSThread()
	assert := assert.New(t)

	r := NewRuntime()
	defer r.Free()
	ctx := r.NewContext()
	defer ctx.Free()
	v, err := ctx.EvalGlobal("JSON.stringify({attachTimerTo:1,b:2,c:[1,2]})")
	defer v.Free()
	assert.Nil(err)
	assert.True(v.IsString())
	assert.Equal(`{"attachTimerTo":1,"b":2,"c":[1,2]}`, v.String())
}

func TestValue_CallWithArgs(t *testing.T) {
	stdruntime.LockOSThread()
	defer stdruntime.UnlockOSThread()
	assert := assert.New(t)

	r := NewRuntime()
	defer r.Free()
	ctx := r.NewContext()
	defer ctx.Free()
	v := ctx.String("中文")
	defer v.Free()
	assert.Equal("中文", v.String())
	ctx.EvalGlobal(`function getArgs() {return arguments}`)
	getArgsFunc := ctx.Globals().Get("getArgs")
	defer getArgsFunc.Free()
	assert.True(getArgsFunc.IsFunction())
	result := getArgsFunc.DynamicCall("v1", 12)
	defer result.Free()
	assert.False(result.IsException())
	assert.Equal(int64(2), result.Len())
	assert.Equal("v1", result.GetStringByUint32(0))
	assert.Equal(int64(12), result.GetInt64ByUint32(1))
}

func TestValue_CallInner(t *testing.T) {
	stdruntime.LockOSThread()
	defer stdruntime.UnlockOSThread()
	assert := assert.New(t)
	called := false
	r := NewRuntime()
	defer r.Free()
	ctx := r.NewContext()
	defer ctx.Free()
	ctx.EvalGlobal(`function runCb(cb,a1) { return cb(a1) }`)
	runCbFunction := ctx.Globals().Get("runCb")
	defer runCbFunction.Free()
	assert.True(runCbFunction.IsFunction())
	f := ctx.Function(func(ctx *Context, this Value, args []Value) Value {
		called = true
		return args[0]
	})
	defer f.Free()

	result := runCbFunction.Call(f, ctx.String("v1"))

	defer result.Free()
	assert.False(result.IsException())
	assert.True(called)
	assert.True(result.IsString())
	assert.Equal("v1", result.String())
}

func TestValue_StringWithChinese(t *testing.T) {
	stdruntime.LockOSThread()
	defer stdruntime.UnlockOSThread()

	assert := assert.New(t)

	r := NewRuntime()
	defer r.Free()
	ctx := r.NewContext()
	defer ctx.Free()
	v := ctx.String("中文")
	defer v.Free()
	assert.Equal("中文", v.String())
	ctx.EvalGlobal(`function fString() {return "中文"}`)
	jsFunctionFString := ctx.Globals().Get("fString")
	defer jsFunctionFString.Free()
	result := jsFunctionFString.Call()
	defer result.Free()
	assert.True(jsFunctionFString.IsFunction())
	assert.Equal("中文", result.String())
}

func TestValue_New(t *testing.T) {
	stdruntime.LockOSThread()
	defer stdruntime.UnlockOSThread()

	assert := assert.New(t)

	r := NewRuntime()
	defer r.Free()
	ctx := r.NewContext()
	defer ctx.Free()
	Date := ctx.Globals().Get("Date")
	defer Date.Free()
	assert.True(Date.IsConstructor())
	date := Date.New()
	defer date.Free()
	assert.True(date.IsObject())
	assert.True(date.HasProperty("getTime"))
	getTimeFunc := date.Get("getTime")
	defer getTimeFunc.Free()
	getTimeResult := getTimeFunc.CallWithContext(date)
	defer getTimeResult.Free()
	getTimeResultTypeof := getTimeResult.TypeOf()
	assert.Equal(JsTypeNumber, getTimeResultTypeof)

}

type ReflectValueTestStruct struct {
	A int32  `mapstructure:"a"`
	B string `mapstructure:"b"`
}

func TestValue_ToReflectValue(t *testing.T) {
	stdruntime.LockOSThread()
	defer stdruntime.UnlockOSThread()

	assert := assert.New(t)

	r := NewRuntime()
	defer r.Free()
	ctx := r.NewContext()
	defer ctx.Free()

	Global := ctx.Globals()

	Global.Set("v1", ctx.Int32(42))
	v1 := Global.Get("v1")
	defer v1.Free()
	reflectV1 := v1.ToReflectValue(reflect.TypeOf(int32(42)))
	assert.Equal(int32(42), reflectV1.Interface())

	ctx.EvalGlobal("var v2 = {a:1,b:'2'}")
	v2 := Global.Get("v2")
	defer v2.Free()
	reflectV2 := v2.ToReflectValue(reflect.TypeOf(ReflectValueTestStruct{}))
	assert.Equal(ReflectValueTestStruct{1, "2"}, reflectV2.Interface())

	ctx.EvalGlobal("var v3 = [1,2,3]")
	v3 := Global.Get("v3")
	defer v3.Free()
	reflectV3 := v3.ToReflectValue(reflect.TypeOf([]int32{}))
	assert.Equal([]int32{1, 2, 3}, reflectV3.Interface())
}

func TestValue_InterfaceAndFree(t *testing.T) {
	stdruntime.LockOSThread()
	defer stdruntime.UnlockOSThread()

	assert := assert.New(t)

	r := NewRuntime()
	defer r.Free()
	ctx := r.NewContext()
	defer ctx.Free()

	ctx.EvalGlobal("function fa() {return {a:{b:123}} }")
	fa := ctx.Globals().Get("fa")
	defer fa.Free()
	assert.True(fa.IsFunction())
	result := fa.DynamicCall().InterfaceAndFree()

	assert.Equal(map[string]interface{}{"a": map[string]interface{}{"b": int64(123)}}, result)
}

func TestValue_IsX(t *testing.T) {
	stdruntime.LockOSThread()
	defer stdruntime.UnlockOSThread()

	assert := assert.New(t)

	r := NewRuntime()
	defer r.Free()
	ctx := r.NewContext()
	defer ctx.Free()

	assert.True(ctx.Int64(10).IsNumber())
	assert.True(ctx.Int64(10).IsIntNumber())
	assert.False(ctx.Float64(10.1).IsIntNumber())
	assert.True(ctx.Float64(10.1).IsFloat64Number())
	assert.True(ctx.Float64(10.1).IsNumber())
	assert.True(ctx.Null().IsNull())
	JSON := ctx.Globals().Get("JSON")
	defer JSON.Free()
	parse := JSON.Get("parse")
	defer parse.Free()
	assert.True(parse.IsFunction())
	assert.False(parse.IsObject())

}
