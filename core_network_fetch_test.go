package quickjs

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestWebCoreFetch(t *testing.T) {
	assert := assert.New(t)

	var t1 time.Time
	var t2 time.Time
	var t3 time.Time

	r := NewRuntime()
	defer r.RunGC()
	ctx := r.NewContext()
	AttachCoreFeaturesToContext(ctx)
	defer ctx.Free()
	var v interface{}

	ctx.Globals().SetGoValue("cb1", func() {
		t1 = time.Now()
	})
	ctx.Globals().SetGoValue("cb2", func() {
		t2 = time.Now()
	})
	ctx.Globals().SetGoValue("nativeCb", func(value map[string]interface{}) {
		v = value
		t3 = time.Now()
	})

	ctx.EvalGlobal(`
fetch(
	"https://postman-echo.com/post", { 
		method: "POST", 
		headers: { 'Content-Type': "application/json" }, 
		body: JSON.stringify( { a : 1 } )
	}
).then(resp => resp.Json().then(nativeCb))
cb1()
setTimeout(cb2, 50)
`)

	ctx.WaitLoopFinished()
	assert.NotNil(v)
	assert.True(t1.Unix() <= t2.Unix())
	assert.True(t1.Unix() <= t3.Unix())

}
