package lib

import "reflect"

// ReflectMakeValPtr returns a `Reflect.Value` that's a pointer to the given value.
//
//   ReflectMakeValPtr(int(0)) // => *int
func ReflectMakeValPtr(val reflect.Value) reflect.Value {
	ptr := reflect.New(val.Type())
	ptr.Elem().Set(val)
	return ptr
}
