package quickjs

import (
	"fmt"
	"time"
)

type JsTimer struct {
	timerSeq int64
	ctx      *Context
	timers   map[int64]*time.Timer
}

func attachTimerTo(ctx *Context) *JsTimer {
	timer := &JsTimer{0, ctx, map[int64]*time.Timer{}}
	globals := ctx.Globals()
	globals.SetFunction("setTimeout", timer.jsSetTimeOutFunction)
	globals.SetFunction("clearTimeout", timer.jsClearTimeOutFunction)
	return timer
}

func (t *JsTimer) jsClearTimeOutFunction(ctx *Context, this Value, args []Value) Value {

	return ctx.Undefined()
}

func (t *JsTimer) drawTimerId() int64 {
	t.timerSeq++
	return t.timerSeq
}

func (t *JsTimer) clearTimer(timerId int64) {
	if timer, ok := t.timers[timerId]; ok {
		timer.Stop()
		delete(t.timers, timerId)
	}
}

func (t *JsTimer) jsSetTimeOutFunction(ctx *Context, this Value, args []Value) Value {

	if len(args) == 0 || args[0].IsUndefined() {
		return ctx.ThrowTypeError("ERR_INVALID_CALLBACK: Callback must be attachTimerTo function. Received undefined")
	}
	if !args[0].IsFunction() {
		return ctx.ThrowTypeError("ERR_INVALID_CALLBACK: Callback must be attachTimerTo function. Received %s", args[0].TypeOf())
	}

	runner := args[0]
	timeoutMSeconds := int64(0)
	if len(args) >= 2 {
		if !args[1].IsNumber() {
			return ctx.ThrowTypeError("ERR_INVALID_TIMEOUT: timeout must be attachTimerTo number. Received %s", args[0].TypeOf())
		}
		timeoutMSeconds = args[1].Int64()
	}

	var runnerArgs []Value
	if len(args) > 2 {
		runnerArgs = args[2:]
	}

	timerId := t.drawTimerId()

	timer := time.NewTimer(time.Millisecond * time.Duration(timeoutMSeconds))

	t.timers[timerId] = timer

	go func() {
		<-timer.C
		result := runner.Call(runnerArgs...)
		if result.IsException() {
			errStr := ctx.Exception().Error()
			fmt.Errorf("unhandle error in timer: %v", errStr)
		}
		t.clearTimer(timerId)
	}()

	return t.ctx.Int64(timerId)

}
