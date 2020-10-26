package quickjs

/*
#cgo CFLAGS: -D_GNU_SOURCE
#cgo CFLAGS: -DCONFIG_BIGNUM
#cgo CFLAGS: -fno-asynchronous-unwind-tables
#cgo LDFLAGS: -lm -lpthread

#include "bridge.h"
*/
import "C"
import (
	"reflect"
	"unicode"
)

var keyMapStructure = "mapstructure"

func isExportedName(name string) bool {

	if len(name) > 0 {
		firstChar := name[0]
		return unicode.IsLetter(rune(firstChar)) && unicode.IsUpper(rune(firstChar))
	}

	return false
}

func getFieldName(field reflect.StructField) string {
	tag := field.Tag.Get(keyMapStructure)
	if len(tag) > 0 {
		return tag
	}
	return field.Name
}

func GetRefCount(ctx *C.JSContext, value C.JSValue) int64 {
	rt := int64(C.GetValueRefCount(ctx, value))
	return rt
}
