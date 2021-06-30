package kv

import (
	"fmt"
	"reflect"
	"strconv"
)

type IndexKey struct {
	stringVal *string
	intVal    *int
	floatVal  *float64
}

func NewIndexKey(val interface{}) IndexKey {
	var stringVal string
	var intVal int
	var floatVal float64
	switch v := val.(type) {
	case string:
		stringVal = v
	case int:
		intVal = v
	case uint:
		intVal = int(v)
	case int32:
		intVal = int(v)
	case uint32:
		intVal = int(v)
	case int64:
		intVal = int(v)
	case uint64:
		intVal = int(v)
	case float64:
		floatVal = v
	case float32:
		floatVal = float64(v)
	default:
		switch reflect.ValueOf(val).Kind() {
		case reflect.String:
			stringVal = fmt.Sprintf("%s", val)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			intVal, _ = strconv.Atoi(fmt.Sprintf("%d", val))
		case reflect.Float32:
			floatVal, _ = strconv.ParseFloat(fmt.Sprintf("%f", val), 32)
		case reflect.Float64:
			floatVal, _ = strconv.ParseFloat(fmt.Sprintf("%f", val), 64)
		}
	}
	var k = IndexKey{}
	if stringVal != "" {
		k.stringVal = &stringVal
	}
	if intVal != 0 {
		k.intVal = &intVal
	}
	if floatVal != 0 {
		k.floatVal = &floatVal
	}
	return k
}

func (k IndexKey) String() string {
	if k.stringVal != nil {
		return *k.stringVal
	}
	return ""
}
func (k IndexKey) Key() string {
	if k.String() != "" {
		return k.String()
	}
	if k.Int() != 0 {
		return strconv.Itoa(k.Int())
	}
	if k.Float() != 0 {
		return fmt.Sprintf("%f", k.Float())
	}
	return ""
}
func (k IndexKey) Int() int {
	if k.stringVal != nil {
		return *k.intVal
	}
	return 0
}
func (k IndexKey) Float() float64 {
	if k.floatVal != nil {
		return *k.floatVal
	}
	return 0
}
