package quickjs

func AttachCoreFeaturesToContext(ctx *Context) {

	globals := ctx.Globals()
	globals.Set("request", ctx.Function(jsRequest))

}
