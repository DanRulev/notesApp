package utils

import (
	"fmt"
	"reflect"
)

func AddFieldsToQuery(field string, value interface{}, fields *[]string, args *[]interface{}, argIdx *int) {
	if value == nil {
		return
	}

	indirectValue := reflect.Indirect(reflect.ValueOf(value))
	if !indirectValue.IsValid() {
		return
	}

	*argIdx++
	*fields = append(*fields, fmt.Sprintf("%s = $%d", field, *argIdx))
	*args = append(*args, indirectValue.Interface())
}
