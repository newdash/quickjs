package quickjs

import (
	"errors"
	"github.com/imroc/req"
)

type FetchInit struct {
	Method  string            `mapstructure:"method"`
	Headers map[string]string `mapstructure:"headers"`
	Body    interface{}       `mapstructure:"body"`
}

// WebCoreFetch for javascript
// implement the `fetch` function of Web API
func WebCoreFetch(ctx *Context, this Value, args []Value) Value {
	if len(args) == 0 {
		return ctx.ThrowError(errors.New("must provide url at least"))
	}
	if !(args[0].IsString()) {
		return ctx.ThrowError(errors.New("must provide a string as url"))
	}
	if len(args) > 1 && !(args[1].IsObject()) {
		return ctx.ThrowError(errors.New("must provide an object as init"))
	}
	return ctx.NewPromise(func(resolve, reject Value) {
		url := args[0].String()
		init := &FetchInit{}
		if len(args) > 1 {
			args[1].Decode(init)
		}
		if len(init.Method) == 0 {
			init.Method = "GET"
		}

		_, err := req.Do(init.Method, url)

		if err != nil {
			reject.Call(ctx.Error(err))
		} else {

		}
	})
}
