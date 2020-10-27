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
	MD5 "crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/imroc/req"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
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

// GoJSObject shortcut
type GoJSObject map[string]interface{}

// GoJSArray shortcut
type GoJSArray []interface{}

func md5(value string) string {
	h := MD5.New()
	io.WriteString(h, value)
	return hex.EncodeToString(h.Sum(nil))
}

// loadTypeScriptSource with local fs cache
func loadTypeScriptSource(version string) (code string, err error) {

	var content []byte
	// jsDelivr is a really awesome CDN
	url := fmt.Sprintf("https://cdn.jsdelivr.net/npm/typescript@%v/lib/typescript.min.js", version)
	cacheLocation := filepath.Join(os.TempDir(), fmt.Sprintf("%v.ts", md5(url)))
	stat, err := os.Stat(cacheLocation)

	if err == nil && !stat.IsDir() {
		// read from local fs cache
		content, err = ioutil.ReadFile(cacheLocation)
		if err != nil {
			return code, err
		}
	} else {
		// get source from remote
		res, err := req.Get(url)
		if err != nil {
			return code, err
		}
		resp := res.Response()
		if resp.StatusCode != http.StatusOK {
			return code, fmt.Errorf(resp.Status)
		}
		content, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return code, err
		}
		// write back to local cache
		ioutil.WriteFile(cacheLocation, content, 0644)
	}

	code = string(content)

	return
}
