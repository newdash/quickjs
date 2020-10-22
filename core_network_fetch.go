package quickjs

import (
	"errors"
	"net/http"
	"strings"

	"github.com/imroc/req"
)

type FetchInit struct {
	Method  string      `mapstructure:"method"`
	Headers req.Header  `mapstructure:"headers"`
	Body    interface{} `mapstructure:"body"`
}

type FetchResponse struct {
	Url        string      `mapstructure:"url"`
	Headers    http.Header `mapstructure:"headers"`
	Status     int         `mapstructure:"status"`
	StatusText string      `mapstructure:"statusText"`
	resp       *req.Resp
	ctx        *Context
}

func (fr FetchResponse) Json() Value {
	return fr.ctx.NewPromise(func(resolve, reject Value) {
		body := fr.resp.String()
		resolve.Call(fr.ctx.ParseJson(body))
	})
}

func (fr FetchResponse) Text() Value {
	return fr.ctx.NewPromise(func(resolve, reject Value) {
		resolve.Call(fr.ctx.ToJSValue(fr.resp.String()))
	})
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

		init.Method = strings.ToUpper(init.Method)

		resp, err := req.Do(init.Method, url, init.Headers, init.Body)

		if err != nil {
			reject.Call(ctx.Error(err))
		} else {
			fetchResponse := FetchResponse{
				Url:        resp.Request().URL.String(),
				ctx:        ctx,
				resp:       resp,
				Headers:    resp.Request().Header,
				Status:     resp.Response().StatusCode,
				StatusText: resp.Response().Status,
			}
			resolve.Call(ctx.ToJSValue(fetchResponse))
		}

	})
}
