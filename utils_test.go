package quickjs

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func Test_isExportedName(t *testing.T) {
	assert := assert.New(t)

	assert.False(isExportedName("runObj"))
	assert.False(isExportedName(""))
	assert.False(isExportedName("123"))

	assert.True(isExportedName("RunObject"))
}

type T1 struct {
	a string
	B string `mapstructure:"b"`
}

func Test_getFieldName(t *testing.T) {
	assert := assert.New(t)
	t1 := T1{}
	v := reflect.ValueOf(t1)
	tp := v.Type()
	assert.Equal("a", getFieldName(tp.Field(0)))
	assert.Equal("b", getFieldName(tp.Field(1)))

}
