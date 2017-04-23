package structs

import (
	"encoding"
	"fmt"
	htemplate "html/template"
	"net"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"text/template"
	"time"

	"github.com/spf13/cast"
)

// UnmarshalValue unmarshals s into value.
// sliceSep, mapKeySep
func UnmarshalValue(value reflect.Value, s string, seps ...rune) error {
	sliceSeparator, mapKeySeparator := separators(seps)

	switch value.Type() {
	case urlType:
		v, err := url.Parse(s)
		if err != nil {
			return err
		}
		value.Set(reflect.ValueOf(v))
		return nil
	case htmltemplateType:
		v, err := htemplate.New("").Parse(s)
		if err != nil {
			return err
		}
		value.Set(reflect.ValueOf(v))
		return nil
	case texttemplateType:
		v, err := template.New("").Parse(s)
		if err != nil {
			return err
		}
		value.Set(reflect.ValueOf(v))
		return nil
	case regexpType:
		v, err := regexp.Compile(s)
		if err != nil {
			return err
		}
		value.Set(reflect.ValueOf(v))
		return nil
	case timeType:
		//TODO add UTC parsing
		v, err := cast.StringToDate(s)
		if err != nil {
			return err
		}
		value.Set(reflect.ValueOf(v))
		return nil
	case durationType:
		v, err := time.ParseDuration(s)
		if err != nil {
			return err
		}
		value.SetInt(int64(v))
		return nil
	case ipaddrType:
		v := net.ParseIP(s)
		value.Set(reflect.ValueOf(v))
		return nil
	case ipnetType:
		_, v, err := net.ParseCIDR(s)
		if err != nil {
			return err
		}
		value.Set(reflect.ValueOf(v))
		return nil
	}

	if dec, ok := ptrValue(value).Interface().(encoding.TextUnmarshaler); ok {
		return dec.UnmarshalText([]byte(s))
	}

	switch value.Kind() {
	default:
		return fmt.Errorf("%v: %v", errCannotUnmarshal, value.Interface())

	case reflect.Bool:
		v, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		value.SetBool(v)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(s, 0, 64)
		if err != nil {
			return err
		}
		value.SetInt(v)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(s, 0, 64)
		if err != nil {
			return err
		}
		value.SetUint(v)

	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		value.SetFloat(v)

	case reflect.String:
		value.SetString(s)

	case reflect.Array:
		r := newcsvreadwriter(sliceSeparator)
		values, err := r.read(s)
		if err != nil {
			return err
		}
		// Make sure the input and array sizes do fit.
		n := value.Len()
		if m := len(values); m < n {
			n = m
		}
		for i := 0; i < n; i++ {
			s := values[i]
			v := value.Index(i)
			if v.Kind() == reflect.Ptr {
				v = v.Elem()
			}
			if err := UnmarshalValue(v, s); err != nil {
				return fmt.Errorf("%s: %v", s, err)
			}
		}

	case reflect.Slice:
		r := newcsvreadwriter(sliceSeparator)
		values, err := r.read(s)
		if err != nil {
			return err
		}
		elem := value.Type().Elem()
		sliceValues := reflect.MakeSlice(value.Type(), 0, len(values))
		for _, s := range values {
			v := reflect.New(elem).Elem()
			if err := UnmarshalValue(v, s); err != nil {
				return fmt.Errorf("%s: %v", s, err)
			}
			sliceValues = reflect.Append(sliceValues, v)
		}
		value.Set(sliceValues)

	case reflect.Map:
		r := newcsvreadwriter(sliceSeparator)
		values, err := r.read(s)
		if err != nil {
			return err
		}
		vType := value.Type()
		keyType := vType.Key()
		elemType := vType.Elem()
		mapValues := reflect.MakeMap(value.Type())

		keyreader := newcsvreadwriter(mapKeySeparator)
		for _, s := range values {
			data, err := keyreader.read(s)
			if err != nil {
				return fmt.Errorf("%s: %v", s, err)
			}
			if len(data) != 2 {
				return fmt.Errorf("%s: %v", s, errInvalidMapKey)
			}
			key := reflect.New(keyType).Elem()
			if err := UnmarshalValue(key, data[0]); err != nil {
				return fmt.Errorf("%s: %v", s, err)
			}
			v := reflect.New(elemType).Elem()
			if err := UnmarshalValue(v, data[1]); err != nil {
				return fmt.Errorf("%s: %v", s, err)
			}
			mapValues.SetMapIndex(key, v)
		}
		value.Set(mapValues)
	}
	return nil
}

// ptrValue returns the interface of the pointer value.
func ptrValue(value reflect.Value) reflect.Value {
	if value.Kind() != reflect.Ptr && value.CanAddr() {
		return value.Addr()
	}
	return value
}
