package structs

import (
	"encoding"
	"fmt"
	htemplate "html/template"
	"net"
	"net/url"
	"reflect"
	"regexp"
	"sort"
	"text/template"
	"time"
)

// MarshalValue converts v into a higher level value or a string as follows:
//  - int, int8, int16, int32 -> int64
//  - uint, uint8, uint16, uint32 -> uint64
//  - float32 -> float64
//  - any slice/map/array -> string
//  - time.Time, *text/template.Template, *html/template.Template, *regexp.RegExp, *url.URL -> string
//  - *net.IPAddr, *net.IPNet -> string
//  - encoding.TextMarshaler -> string
//
// The following types are returned as is:
//  - bool, time.Duration, float64, int, int64, string, uint, uint64
//
// sliceSep, mapKeySep
func MarshalValue(v interface{}, seps ...rune) (interface{}, error) {
	// v = indirect(v)
	sliceSeparator, mapKeySeparator := separators(seps)

	switch w := v.(type) {
	case nil:
		// May error further down.
	case bool, time.Duration, float64, int64, string, uint64:
		return w, nil
	case float32:
		return float64(w), nil
	case int:
		return int64(w), nil
	case int8:
		return int64(w), nil
	case int16:
		return int64(w), nil
	case int32:
		return int64(w), nil
	case uint:
		return uint64(w), nil
	case uint8:
		return uint64(w), nil
	case uint16:
		return uint64(w), nil
	case uint32:
		return uint64(w), nil
	// Check the following types first in case they implement encoding.TextMarshaler.
	case time.Time:
		return w.String(), nil
	case *url.URL:
		if w == nil {
			return "", nil
		}
		return w.String(), nil
	case *regexp.Regexp:
		if w == nil {
			// Return a valid regexp.
			return ".*", nil
		}
		return w.String(), nil
	case *template.Template:
		if w == nil {
			return "", nil
		}
		return w.Tree.Root.String(), nil
	case *htemplate.Template:
		if w == nil {
			return "", nil
		}
		return w.Tree.Root.String(), nil
	case *net.IPAddr:
		if w == nil {
			return "", nil
		}
		return w.String(), nil
	case *net.IPNet:
		if w == nil {
			return "", nil
		}
		return w.String(), nil

	case encoding.TextMarshaler:
		bts, err := w.MarshalText()
		if err != nil {
			return nil, err
		}
		return string(bts), nil
	}

	var lst []string
	value := reflect.ValueOf(v)
	switch value.Kind() {
	case reflect.Slice, reflect.Array:
		n := value.Len()
		lst = make([]string, n)
		for i := 0; i < n; i++ {
			v := value.Index(i)
			w, err := MarshalValue(v.Interface())
			if err != nil {
				return nil, err
			}
			lst[i] = fmt.Sprintf("%v", w)
		}

	case reflect.Map:
		keycsv := newcsvreadwriter(mapKeySeparator)
		keys := value.MapKeys()
		lst = make([]string, len(keys))
		for i, key := range keys {
			v := value.MapIndex(key)
			w, err := MarshalValue(v.Interface())
			if err != nil {
				return nil, err
			}
			keyString := fmt.Sprintf("%v", key.Interface())
			valString := fmt.Sprintf("%v", w)
			lst[i], err = keycsv.write(keyString, valString)
			if err != nil {
				return nil, err
			}
		}
		// To garantee a stable output...
		sort.Sort(sort.StringSlice(lst))

	default:
		return nil, fmt.Errorf("marshal: unsupported type %T", v)
	}

	csv := newcsvreadwriter(sliceSeparator)
	return csv.write(lst...)
}

// From html/template/content.go
// Copyright 2011 The Go Authors. All rights reserved.
// indirect returns the value, after dereferencing as many times
// as necessary to reach the base type (or nil).
func indirect(a interface{}) interface{} {
	if a == nil {
		return nil
	}
	if t := reflect.TypeOf(a); t.Kind() != reflect.Ptr {
		// Avoid creating a reflect.Value if it's not a pointer.
		return a
	}
	v := reflect.ValueOf(a)
	for v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	return v.Interface()
}
