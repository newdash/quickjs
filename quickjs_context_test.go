package quickjs

import (
	"github.com/stretchr/testify/assert"
	stdruntime "runtime"
	"testing"
)

func TestContext_CreateObjectWithMap(t *testing.T) {
	stdruntime.LockOSThread()

	assert := assert.New(t)
	r := NewRuntime()
	defer r.Free()
	ctx := r.NewContext()
	defer ctx.Free()
	v := ctx.ToJSValue(map[string]interface{}{"attachTimerTo": 1, "V": 2})
	defer v.Free()
	assert.True(v.IsObject())
	assert.True(v.Get("attachTimerTo").IsNumber())
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
	v := ctx.ToJSValue(TestStruct{1, "2"})
	defer v.Free()
	assert.True(v.IsObject())
	assert.True(v.Get("A").IsNumber())
	assert.Equal(int64(1), v.Get("A").Int64())
	ov := v.Get("V")
	defer ov.Free()
	assert.True(ov.IsString())
	assert.Equal("2", ov.String())

}

func TestContext_ParseJson(t *testing.T) {
	assert := assert.New(t)
	r := NewRuntime()
	defer r.RunGC()
	ctx := r.NewContext()
	defer ctx.Free()

	v := ctx.ParseJson(`{"attachTimerTo":"1"}`)
	defer v.Free()
	assert.True(v.IsObject())
	assert.Equal("1", v.Get("attachTimerTo").String())

}

type DemoObject struct{}

func (s DemoObject) Add(v1, v2 int64) int64 {
	return v1 + v2
}

func TestContext_ToJSValueWithFunc(t *testing.T) {
	stdruntime.LockOSThread()
	assert := assert.New(t)
	r := NewRuntime()
	defer r.Free()
	ctx := r.NewContext()
	defer ctx.Free()

	Global := ctx.Globals()
	Global.SetGoValue("Demo", DemoObject{})

	Demo := Global.Get("Demo")
	defer Demo.Free()
	assert.True(Demo.IsObject())
	Add := Demo.Get("Add")
	defer Add.Free()
	assert.True(Add.IsFunction())
	assert.Equal(int64(42), Add.Call(ctx.ToJSValue(1), ctx.ToJSValue(41)).Int64())
}

type DemoObject2 struct {
	a string
	B string `mapstructure:"b"`
}

func TestContext_ToJSValueWithPrivateField(t *testing.T) {
	stdruntime.LockOSThread()
	assert := assert.New(t)
	r := NewRuntime()
	defer r.Free()
	ctx := r.NewContext()
	defer ctx.Free()

	Global := ctx.Globals()
	Global.SetGoValue("Demo", DemoObject2{"v1", "v2"})

	Demo := Global.Get("Demo")
	defer Demo.Free()
	assert.True(Demo.IsObject())
	a := Demo.Get("a")
	defer a.Free()
	assert.True(a.IsUndefined())
}
