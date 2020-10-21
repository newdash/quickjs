package quickjs

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestContext_CreateObjectWithMap(t *testing.T) {
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
	assert.Equal("2", v.Get("V").String())

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
