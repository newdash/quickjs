package quickjs

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestValue_ToJSONString(t *testing.T) {
	assert := assert.New(t)
	r := NewRuntime()
	defer r.RunGC()
	ctx := r.NewContext()
	defer ctx.Free()

	v := ctx.ToJSValue(map[string]string{"attachTimerTo": "1"})
	defer v.Free()
	assert.True(v.IsObject())
	assert.Equal(`{"attachTimerTo":"1"}`, v.ToJsonString())

}

func TestContext_CallFunction(t *testing.T) {
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
func TestContext_CallFunctionWithoutArgs(t *testing.T) {
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
	assert := assert.New(t)

	r := NewRuntime()
	defer r.RunGC()
	ctx := r.NewContext()
	defer ctx.Free()
	v, err := ctx.EvalGlobal("JSON.stringify({attachTimerTo:1,b:2,c:[1,2]})")
	assert.Nil(err)
	assert.True(v.IsString())
	assert.Equal(`{"attachTimerTo":1,"b":2,"c":[1,2]}`, v.String())
}

func TestValue_CallWithArgs(t *testing.T) {
	assert := assert.New(t)

	r := NewRuntime()
	defer r.RunGC()
	ctx := r.NewContext()
	defer ctx.Free()
	v := ctx.String("中文")
	defer v.Free()
	assert.Equal("中文", v.String())
	ctx.EvalGlobal(`function getArgs() {return arguments}`)
	getArgsFunc := ctx.Globals().Get("getArgs")
	assert.True(getArgsFunc.IsFunction())
	result := getArgsFunc.Call(ctx.String("v1"), ctx.Int64(12))
	assert.False(result.IsException())
	assert.Equal(int64(2), result.Len())
	assert.Equal("v1", result.GetByUint32(0).String())
	assert.Equal(int64(12), result.GetByUint32(1).Int64())
}

func TestValue_CallInner(t *testing.T) {
	assert := assert.New(t)
	called := false
	r := NewRuntime()
	defer r.RunGC()
	ctx := r.NewContext()
	defer ctx.Free()
	ctx.EvalGlobal(`function runCb(cb,a1) { return cb(a1) }`)
	runCbFunction := ctx.Globals().Get("runCb")
	assert.True(runCbFunction.IsFunction())

	result := runCbFunction.Call(ctx.Function(func(ctx *Context, this Value, args []Value) Value {
		called = true
		return args[0]
	}), ctx.String("v1"))

	assert.False(result.IsException())
	assert.True(called)
	assert.True(result.IsString())
	assert.Equal("v1", result.String())
}

func TestValue_StringWithChinese(t *testing.T) {
	assert := assert.New(t)

	r := NewRuntime()
	defer r.RunGC()
	ctx := r.NewContext()
	defer ctx.Free()
	v := ctx.String("中文")
	defer v.Free()
	assert.Equal("中文", v.String())
	ctx.EvalGlobal(`function fString() {return "中文"}`)
	jsFunctionFString := ctx.Globals().Get("fString")
	assert.True(jsFunctionFString.IsFunction())
	assert.Equal("中文", jsFunctionFString.Call().String())
}

func TestValue_New(t *testing.T) {
	assert := assert.New(t)

	r := NewRuntime()
	defer r.RunGC()
	ctx := r.NewContext()
	defer ctx.Free()
	Date := ctx.Globals().Get("Date")
	assert.True(Date.IsConstructor())
	date := Date.New()
	assert.True(date.IsObject())
	assert.True(date.HasProperty("getTime"))
	getTimeResult := date.Get("getTime").CallWithContext(date)
	getTimeResultTypeof := getTimeResult.TypeOf()
	assert.Equal(JsTypeNumber, getTimeResultTypeof)

}

type ReflectValueTestStruct struct {
	A int32  `mapstructure:"a"`
	B string `mapstructure:"b"`
}

func TestValue_ToReflectValue(t *testing.T) {
	assert := assert.New(t)

	r := NewRuntime()
	defer r.RunGC()
	ctx := r.NewContext()
	defer ctx.Free()

	Global := ctx.Globals()

	Global.Set("v1", ctx.Int32(42))
	reflectV1 := Global.Get("v1").ToReflectValue(reflect.TypeOf(int32(42)))
	assert.Equal(int32(42), reflectV1.Interface())

	ctx.EvalGlobal("var v2 = {a:1,b:'2'}")
	reflectV2 := Global.Get("v2").ToReflectValue(reflect.TypeOf(ReflectValueTestStruct{}))
	assert.Equal(ReflectValueTestStruct{1, "2"}, reflectV2.Interface())

	ctx.EvalGlobal("var v3 = [1,2,3]")
	reflectV3 := Global.Get("v3").ToReflectValue(reflect.TypeOf([]int32{}))
	assert.Equal([]int32{1, 2, 3}, reflectV3.Interface())
}
