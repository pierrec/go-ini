package structs

import (
	"fmt"
	"reflect"
)

// Set assigns v to the value.
// If v is a string but value is not, then Set attempts to deserialize it
// using UnmarshalValue().
func Set(value reflect.Value, v interface{}, seps ...rune) error {
	if !value.CanSet() {
		return errCannotSet
	}

	if s, ok := v.(string); ok {
		return UnmarshalValue(value, s, seps...)
	}

	val := reflect.ValueOf(v)
	if value.Kind() != val.Kind() {
		// The value was converted.
		v, err := convert(val, value)
		if err != nil {
			return err
		}
		val = v
	}
	value.Set(val)
	return nil
}

// convert a to b safely.
func convert(a, b reflect.Value) (_ reflect.Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	return a.Convert(b.Type()), nil
}
