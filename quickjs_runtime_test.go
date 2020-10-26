package quickjs

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	stdruntime "runtime"
	"sync"
	"testing"
)

func TestObject(t *testing.T) {
	stdruntime.UnlockOSThread()

	runtime := NewRuntime()
	defer runtime.Free()
	context := runtime.NewContext()
	defer context.Free()

	test := context.Object()
	test.Set("A", context.String("String A"))
	test.Set("B", context.String("String B"))
	test.Set("C", context.String("String C"))
	context.Globals().Set("test", test)

	result, err := context.EvalGlobal(`Object.keys(test).map(key => test[key]).join(" ")`)
	require.NoError(t, err)
	defer result.Free()

	require.EqualValues(t, "String A String B String C", result.String())
}

func TestArray(t *testing.T) {
	runtime := NewRuntime()
	defer runtime.Free()

	context := runtime.NewContext()
	defer context.Free()

	test := context.Array()
	for i := int64(0); i < 3; i++ {
		test.SetByInt64(i, context.String(fmt.Sprintf("test %d", i)))
	}
	for i := int64(0); i < test.Len(); i++ {
		v := test.GetByUint32(uint32(i))
		require.EqualValues(t, fmt.Sprintf("test %d", i), v.String())
		v.Free()
	}

	context.Globals().Set("test", test)

	result, err := context.EvalGlobal(`test.map(v => v.toUpperCase())`)
	require.NoError(t, err)
	defer result.Free()

	require.EqualValues(t, `TEST 0,TEST 1,TEST 2`, result.String())
}

func TestBadSyntax(t *testing.T) {
	runtime := NewRuntime()
	defer runtime.Free()

	context := runtime.NewContext()
	defer context.Free()

	_, err := context.EvalGlobal(`"bad syntax'`)
	require.Error(t, err)
}

func TestFunctionThrowError(t *testing.T) {
	expected := errors.New("expected error")

	runtime := NewRuntime()
	defer runtime.Free()

	context := runtime.NewContext()
	defer context.Free()

	context.Globals().SetFunction("A", func(ctx *Context, this Value, args []Value) Value {
		return ctx.ThrowError(expected)
	})

	_, actual := context.EvalGlobal("A()")
	require.Error(t, actual)
	require.EqualValues(t, "Error: "+expected.Error(), actual.Error())
}

func TestFunction(t *testing.T) {
	stdruntime.LockOSThread()
	runtime := NewRuntime()
	defer runtime.Free()

	context := runtime.NewContext()
	defer context.Free()

	A := make(chan struct{})
	B := make(chan struct{})

	context.Globals().SetFunction("A", func(ctx *Context, this Value, args []Value) Value {
		require.Len(t, args, 4)
		require.True(t, args[0].IsString() && args[0].String() == "hello world!")
		require.True(t, args[1].IsNumber() && args[1].Int32() == 1)
		require.True(t, args[2].IsNumber() && args[2].Int64() == 8)
		require.True(t, args[3].IsNull())

		close(A)

		return ctx.String("A says hello")
	})

	context.Globals().SetFunction("B", func(ctx *Context, this Value, args []Value) Value {
		require.Len(t, args, 0)

		close(B)

		return ctx.Float64(256)
	})

	result, err := context.EvalGlobal(`A("hello world!", 1, 2 ** 3, null)`)
	require.NoError(t, err)
	defer result.Free()

	require.True(t, result.IsString() && result.String() == "A says hello")
	<-A

	result, err = context.EvalGlobal(`B()`)
	require.NoError(t, err)
	defer result.Free()

	require.True(t, result.IsNumber() && result.Uint32() == 256)
	<-B
}

func TestConcurrency(t *testing.T) {
	n := 32
	m := 10000

	var wg sync.WaitGroup
	wg.Add(n)

	req := make(chan struct{}, n)
	res := make(chan int64, m)

	for i := 0; i < n; i++ {
		go func() {
			stdruntime.LockOSThread()

			defer wg.Done()

			runtime := NewRuntime()
			defer runtime.Free()

			context := runtime.NewContext()
			defer context.Free()

			for range req {
				result, err := context.EvalGlobal(`new Date().getTime()`)
				require.NoError(t, err)

				res <- result.Int64()

				result.Free()
			}
		}()
	}

	for i := 0; i < m; i++ {
		req <- struct{}{}
	}
	close(req)

	wg.Wait()

	for i := 0; i < m; i++ {
		<-res
	}
}

func TestContext_CreateObjectWithPrimitive(t *testing.T) {
	assert := assert.New(t)

	r := NewRuntime()
	defer r.Free()
	ctx := r.NewContext()
	defer ctx.Free()
	v := ctx.ToJSValue(1)
	assert.True(v.IsNumber())
	v = ctx.ToJSValue(1.1)
	assert.True(v.IsNumber())
	assert.Equal(1.1, v.Float64())
	v = ctx.ToJSValue(true)
	assert.True(v.IsBool())
	v = ctx.ToJSValue(nil)
	assert.True(v.IsUndefined())
	v = ctx.ToJSValue("hello")
	defer v.Free()
	assert.True(v.IsString())
}

func TestContext_GlobalsGet(t *testing.T) {
	stdruntime.UnlockOSThread()

	assert := assert.New(t)
	r := NewRuntime()
	defer r.RunGC()
	ctx := r.NewContext()
	defer ctx.Free()
	global := ctx.Globals()
	globalObject := global.Get("Object")
	assert.True(globalObject.IsConstructor())
	globalObjectKey := globalObject.Get("keys")
	assert.True(globalObjectKey.IsFunction())
	result := globalObjectKey.CallWithContext(ctx.Null(), ctx.ToJSValue(map[string]string{"attachTimerTo": "v"}))
	defer result.Free()
	assert.True(result.IsArray())
	assert.Equal(int64(1), result.Len())
	assert.Equal([]interface{}{"attachTimerTo"}, result.Interface())
}

func TestValue_PropertyNames(t *testing.T) {
	assert := assert.New(t)
	r := NewRuntime()
	defer r.RunGC()
	ctx := r.NewContext()
	defer ctx.Free()
	_, err := ctx.EvalGlobal("class A { constructor() { this.attachTimerTo = 1 } };")
	assert.Nil(err)
	_, err = ctx.EvalGlobal("class B extends A { constructor() { super(); this.b = 1 } }; ")
	assert.Nil(err)
	v, err := ctx.EvalGlobal("new B()")
	assert.Nil(err)
	assert.True(v.IsObject())
	names, err := v.PropertyNames()
	assert.Nil(err)
	assert.Equal(int(2), len(names))
	assert.Equal("attachTimerTo", names[0].String())

}

func TestValue_InterfaceObject(t *testing.T) {
	assert := assert.New(t)

	r := NewRuntime()
	defer r.RunGC()
	ctx := r.NewContext()
	defer ctx.Free()
	v := ctx.ToJSValue(map[string]interface{}{"attachTimerTo": map[string]interface{}{"b": int64(1)}})
	assert.True(v.IsObject())
	assert.Equal(map[string]interface{}{"attachTimerTo": map[string]interface{}{"b": int64(1)}}, v.Interface())

	v2, err := ctx.EvalGlobal("var object = {attachTimerTo:{b:1}};object")
	assert.Nil(err)
	assert.True(v2.IsObject())
	assert.Equal(map[string]interface{}{"attachTimerTo": map[string]interface{}{"b": int64(1)}}, v2.Interface())

}

type DecodeStructA struct {
	B int
}
type DecodeStructBase struct {
	A DecodeStructA
}

func TestValue_Decode(t *testing.T) {
	assert := assert.New(t)

	r := NewRuntime()
	defer r.RunGC()
	ctx := r.NewContext()
	defer ctx.Free()
	v2, err := ctx.EvalGlobal("var object = {a:{b:1}};object")
	defer v2.Free()
	assert.Nil(err)
	assert.True(v2.IsObject())
	structA := &DecodeStructBase{}
	v2.Decode(structA)
	assert.Equal(1, structA.A.B)

}
