package ini

import (
	"encoding"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

var textMarshalType = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()

// Encode writes the contents of v to the given Writer.
// DefaultOptions are used.
// See Ini.Encode() for more information.
func Encode(w io.Writer, v interface{}) error {
	ini, err := New(DefaultOptions...)
	if err != nil {
		return err
	}
	if err = ini.Encode(v); err != nil {
		return err
	}
	_, err = ini.WriteTo(w)
	return err
}

// Encode sets Ini sections and keys according to the values defined in v.
// v must be a pointer to a struct.
func (ini *Ini) Encode(v interface{}) error {
	// v must be a pointer.
	ptr := reflect.ValueOf(v)
	if ptr.Kind() != reflect.Ptr {
		return errDecodeNoPtr
	}
	// v must be a pointer to a struct.
	val := ptr.Elem()
	if val.Kind() != reflect.Struct {
		return errDecodeNoStruct
	}

	// Make sure to convert the value to its interface
	// otherwise TypeOf will look at the reflect.Value itself!
	vType := reflect.TypeOf(val.Interface())
	for i, n := 0, vType.NumField(); i < n; i++ {
		value := val.Field(i)
		if !value.CanSet() {
			// Cannot set the field, maybe unexported.
			continue
		}
		field := vType.Field(i)
		fieldKind := field.Type.Kind()

		// If the field implements encoding.TextMarshaler, it prevails.
		valuePtr, isTexter := getMarshalTexter(value)
		if !isTexter {
			switch fieldKind {
			case reflect.Bool,
				reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
				reflect.Float32, reflect.Float64,
				reflect.String, reflect.Slice:
			case reflect.Struct:
				err := ini.Encode(valuePtr.Interface())
				if err != nil {
					return err
				}
				continue
			default:
				continue
			}
		}

		section, key, isLastKey := getTagInfo(field)

		var keyValue string
		if isTexter {
			vals := valuePtr.MethodByName("MarshalText").Call(nil)
			if v := vals[1]; !v.IsNil() {
				return fmt.Errorf("ini: encode: %s.%s: %v", section, key, v.Interface())
			}
			value = vals[0]
			keyValue = string(value.Interface().([]byte))
		} else if fieldKind == reflect.Slice {
			n := value.Len()
			keyValues := make([]string, n)
			for i := 0; i < n; i++ {
				v := value.Index(i)
				keyValues[i] = fmt.Sprintf("%v", v.Interface())
			}
			keyValue = strings.Join(keyValues, ini.sliceSep)
		} else {
			keyValue = fmt.Sprintf("%v", value.Interface())
		}
		ini.Set(section, key, keyValue)

		if isLastKey {
			ini.Set(section, "", "")
		}
	}

	return nil
}

// Figure out the key and section to look for in Ini.
// If not specified, use the field name as the key.
// A struct tag may contain 3 entries:
//  - the key name (defaults to the field name)
//  - the section name (defaults to the global section)
//  - whether the key is the last of a block, which introduces a newline
func getTagInfo(field reflect.StructField) (string, string, bool) {
	var section, key string
	var isLastKey bool
	if tag := field.Tag.Get("ini"); tag != "" {
		lst := strings.Split(tag, ",")
		n := len(lst)
		if n > 0 {
			key = lst[0]
		}
		if n > 1 {
			section = lst[1]
		}
		if n > 2 {
			isLastKey, _ = strconv.ParseBool(lst[2])
		}
	}
	if key == "" {
		key = field.Name
	}
	return section, key, isLastKey
}

// getMarshalTexter returns the pointer to the Value and whether it
// implements the TextMarshaler interface.
func getMarshalTexter(v reflect.Value) (reflect.Value, bool) {
	p := v
	if v.Kind() != reflect.Ptr {
		p = v.Addr()
	}
	return p, p.Type().Implements(textMarshalType)
}
