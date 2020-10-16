package quickjs

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValue_ToJSONString(t *testing.T) {
	assert := assert.New(t)
	r := NewRuntime()
	defer r.RunGC()
	ctx := r.NewContext()
	defer ctx.Free()

	v := ctx.ToJSValue(map[string]string{"a": "1"})
	defer v.Free()
	assert.True(v.IsObject())
	assert.Equal(`{"a":"1"}`, v.ToJsonString())

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
