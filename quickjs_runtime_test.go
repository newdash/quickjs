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
	runtime := NewRuntime()
	defer runtime.Free()

	context := runtime.NewContext()
	defer context.Free()

	test := context.Object()
	test.Set("A", context.String("String A"))
	test.Set("B", context.String("String B"))
	test.Set("C", context.String("String C"))
	context.Globals().Set("test", test)

	result, err := context.Eval(`Object.keys(test).map(key => test[key]).join(" ")`)
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
		require.EqualValues(t, fmt.Sprintf("test %d", i), test.GetByUint32(uint32(i)).String())
	}

	context.Globals().Set("test", test)

	result, err := context.Eval(`test.map(v => v.toUpperCase())`)
	require.NoError(t, err)
	defer result.Free()

	require.EqualValues(t, `TEST 0,TEST 1,TEST 2`, result.String())
}

func TestBadSyntax(t *testing.T) {
	runtime := NewRuntime()
	defer runtime.Free()

	context := runtime.NewContext()
	defer context.Free()

	_, err := context.Eval(`"bad syntax'`)
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

	_, actual := context.Eval("A()")
	require.Error(t, actual)
	require.EqualValues(t, "Error: "+expected.Error(), actual.Error())
}

func TestFunction(t *testing.T) {
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

	result, err := context.Eval(`A("hello world!", 1, 2 ** 3, null)`)
	require.NoError(t, err)
	defer result.Free()

	require.True(t, result.IsString() && result.String() == "A says hello")
	<-A

	result, err = context.Eval(`B()`)
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
				result, err := context.Eval(`new Date().getTime()`)
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
	v := ctx.CreateObjectWith(1)
	assert.True(v.IsNumber())
	v = ctx.CreateObjectWith(1.1)
	assert.True(v.IsNumber())
	assert.Equal(1.1, v.Float64())
	v = ctx.CreateObjectWith(true)
	assert.True(v.IsBool())
	v = ctx.CreateObjectWith(nil)
	assert.True(v.IsNull())
	v = ctx.CreateObjectWith("hello")
	assert.True(v.IsString())
	r.RunGC()
}

func TestContext_GlobalsGet(t *testing.T) {
	assert := assert.New(t)
	r := NewRuntime()
	defer r.RunGC()
	ctx := r.NewContext()
	defer ctx.Free()
	global := ctx.Globals()
	globalObject := global.Get("Object")
	assert.True(globalObject.IsObject())
	globalObjectKey := globalObject.Get("keys")
	assert.True(globalObjectKey.IsFunction())
	result := globalObjectKey.Call(ctx.Null(), ctx.ToJSValue(map[string]string{"a": "v"}))
	defer result.Free()
	assert.True(result.IsArray())
	assert.Equal(int64(1), result.Len())
	assert.Equal([]interface{}{"a"}, result.Interface())
}

func TestValue_PropertyNames(t *testing.T) {
	assert := assert.New(t)
	r := NewRuntime()
	defer r.RunGC()
	ctx := r.NewContext()
	defer ctx.Free()
	_, err := ctx.Eval("class A { constructor() { this.a = 1 } };")
	assert.Nil(err)
	_, err = ctx.Eval("class B extends A { constructor() { super(); this.b = 1 } }; ")
	assert.Nil(err)
	v, err := ctx.Eval("new B()")
	assert.Nil(err)
	assert.True(v.IsObject())
	names, err := v.PropertyNames()
	assert.Nil(err)
	assert.Equal(int(2), len(names))
	assert.Equal("a", names[0].String())

}

func TestContext_CreateObjectWithMap(t *testing.T) {
	assert := assert.New(t)
	r := NewRuntime()
	defer r.Free()
	ctx := r.NewContext()
	defer ctx.Free()
	v := ctx.CreateObjectWith(map[string]interface{}{"a": 1, "V": 2})
	defer v.Free()
	assert.True(v.IsObject())
	assert.True(v.Get("a").IsNumber())
	assert.Equal(int64(2), v.Get("V").Int64())
}

type TestStruct struct {
	A int
	V string
}

func TestContext_CreateObjectWithStruct(t *testing.T) {
	assert := assert.New(t)

	r := NewRuntime()
	defer r.Free()
	ctx := r.NewContext()
	defer ctx.Free()
	v := ctx.CreateObjectWith(TestStruct{1, "2"})
	defer v.Free()
	assert.True(v.IsObject())
	assert.True(v.Get("A").IsNumber())
	assert.Equal(int64(1), v.Get("A").Int64())
	assert.Equal("2", v.Get("V").String())

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
	result := f.Call(ctx.Null(), arg0)
	assert.False(result.IsError())
	assert.Equal(int64(4444), result.Interface())

	ctx.Globals().DeleteProperty("f")

	arg0.Free()
	result.Free()
	r.RunGC()
}

func TestValue_InterfaceNumber(t *testing.T) {
	assert := assert.New(t)

	r := NewRuntime()
	defer r.Free()
	defer r.RunGC()
	ctx := r.NewContext()
	defer ctx.Free()
	v, err := ctx.Eval("1")
	defer v.Free()
	assert.Nil(err)
	assert.True(v.IsNumber())
	assert.True(v.IsIntNumber())
	assert.False(v.IsFloat64Number())
	v2, err2 := ctx.Eval("1.1")
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
	v, err := ctx.Eval("[1,2,3,4.3]")
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
	v, err := ctx.Eval("JSON.stringify({a:1,b:2,c:[1,2]})")
	assert.Nil(err)
	assert.True(v.IsString())
	assert.Equal(`{"a":1,"b":2,"c":[1,2]}`, v.String())
}

func TestValue_InterfaceObject(t *testing.T) {
	assert := assert.New(t)

	r := NewRuntime()
	defer r.RunGC()
	ctx := r.NewContext()
	defer ctx.Free()
	v := ctx.ToJSValue(map[string]interface{}{"a": map[string]interface{}{"b": int64(1)}})
	assert.True(v.IsObject())
	assert.Equal(map[string]interface{}{"a": map[string]interface{}{"b": int64(1)}}, v.Interface())

	v2, err := ctx.Eval("var object = {a:{b:1}};object")
	assert.Nil(err)
	assert.True(v2.IsObject())
	assert.Equal(map[string]interface{}{"a": map[string]interface{}{"b": int64(1)}}, v2.Interface())

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
	v2, err := ctx.Eval("var object = {a:{b:1}};object")
	defer v2.Free()
	assert.Nil(err)
	assert.True(v2.IsObject())
	structA := &DecodeStructBase{}
	v2.Decode(structA)
	assert.Equal(1, structA.A.B)

}
