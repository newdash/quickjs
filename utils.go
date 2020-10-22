package quickjs

import (
	"reflect"
	"unicode"
)

var keyMapstructure = "mapstructure"

func isExportedName(name string) bool {

	if len(name) > 0 {
		firstChar := name[0]
		return unicode.IsLetter(rune(firstChar)) && unicode.IsUpper(rune(firstChar))
	}

	return false
}

func getFieldName(field reflect.StructField) string {
	tag := field.Tag.Get(keyMapstructure)
	if len(tag) > 0 {
		return tag
	}
	return field.Name
}
