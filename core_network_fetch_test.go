package quickjs

import (
	"github.com/stretchr/testify/assert"
	"runtime"
	"testing"
)

func TestWebCoreFetch(t *testing.T) {
	assert := assert.New(t)
	runtime.LockOSThread()

	r := NewRuntime()
	defer r.RunGC()
	ctx := r.NewContext()
	AttachCoreFeaturesToContext(ctx)
	defer ctx.Free()
	var v interface{}

	ctx.Globals().SetGoValue("nativeCb", func(value map[string]interface{}) {
		v = value
	})

	ctx.EvalGlobal(`
const response = request(
	"https://postman-echo.com/post",
	{
		method: "POST",
		headers: { 'Content-Type': "application/json" },
		body: JSON.stringify( { a : 1 } )
	}
);
nativeCb(response)`)

	ctx.Globals().DeleteProperty("nativeCb")

	assert.NotNil(v)
	assert.Equal(
		map[string]interface{}{"a": int64(1)},
		v.(map[string]interface{})["Json"].(func(...interface{}) interface{})().(map[string]interface{})["json"],
	)

}
