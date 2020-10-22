package quickjs

func AttachCoreFeaturesToContext(ctx *Context) {

	globals := ctx.Globals()
	globals.Set("fetch", ctx.Function(WebCoreFetch))

}
