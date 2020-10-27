#include "stdlib.h"
#include "quickjs.h"
#include "list.h"

extern JSValue InvokeProxy(JSContext *ctx, JSValueConst this_val, int argc, JSValueConst *argv);

static JSValue JS_NewNull() { return JS_NULL; }
static JSValue JS_NewUndefined() { return JS_UNDEFINED; }
static JSValue JS_NewUninitialized() { return JS_UNINITIALIZED; }

static JSValue ThrowSyntaxError(JSContext *ctx, const char *fmt) { return JS_ThrowSyntaxError(ctx, "%s", fmt); }
static JSValue ThrowTypeError(JSContext *ctx, const char *fmt) { return JS_ThrowTypeError(ctx, "%s", fmt); }
static JSValue ThrowReferenceError(JSContext *ctx, const char *fmt) { return JS_ThrowReferenceError(ctx, "%s", fmt); }
static JSValue ThrowRangeError(JSContext *ctx, const char *fmt) { return JS_ThrowRangeError(ctx, "%s", fmt); }
static JSValue ThrowInternalError(JSContext *ctx, const char *fmt) { return JS_ThrowInternalError(ctx, "%s", fmt); }

static int GetValueRefCount(JSContext *ctx, JSValue v)
{
    if (JS_VALUE_HAS_REF_COUNT(v))
    {
        JSRefCountHeader *p = (JSRefCountHeader *)JS_VALUE_GET_PTR(v);
        return p->ref_count;
    }
    return 0;
}



static JS_BOOL JS_IsIntNumber(JSValueConst v)
{
    int tag = JS_VALUE_GET_TAG(v);
    return tag == JS_TAG_INT;
}

static JS_BOOL JS_IsFloat64Number(JSValueConst v)
{
    int tag = JS_VALUE_GET_TAG(v);
    return JS_TAG_IS_FLOAT64(tag);
}