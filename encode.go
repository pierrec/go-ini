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
func (ini *INI) Encode(v interface{}) error {
	return ini.encode("", v)
}

func (ini *INI) encode(defaultSection string, v interface{}) error {
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
				reflect.String, reflect.Slice, reflect.Array, reflect.Map:
			case reflect.Struct:
				section, key, _ := getTagInfo(field)
				if key == "" {
					// Omit the field.
					continue
				}
				if section == "" && field.Anonymous {
					// If embedded, use the field name as the default section name.
					section = field.Name
				}
				err := ini.encode(section, valuePtr.Interface())
				if err != nil {
					return err
				}
				continue
			default:
				continue
			}
		}

		section, key, isLastKey := getTagInfo(field)
		if key == "" {
			// Omit the field.
			continue
		}
		if section == "" {
			section = defaultSection
		}

		keyValue, err := ini.encodeValue(value, valuePtr, isTexter)
		if err != nil {
			return fmt.Errorf("ini: encode: %s.%s: %v", section, key, err)
		}
		ini.Set(section, key, keyValue)

		if isLastKey {
			ini.Set(section, "", "")
		}
	}

	return nil
}

func (ini *INI) encodeValue(value, valuePtr reflect.Value, isTexter bool) (string, error) {
	if isTexter {
		vals := valuePtr.MethodByName("MarshalText").Call(nil)
		if v := vals[1]; !v.IsNil() {
			return "", v.Interface().(error)
		}
		value = vals[0]
		return string(value.Interface().([]byte)), nil
	}

	fieldKind := value.Type().Kind()
	switch fieldKind {
	case reflect.Slice, reflect.Array:
		n := value.Len()
		keyValues := make([]string, n)
		for i := 0; i < n; i++ {
			v := value.Index(i)
			w, isTexter := getMarshalTexter(v)
			s, err := ini.encodeValue(v, w, isTexter)
			if err != nil {
				return "", err
			}
			keyValues[i] = s
		}
		//TODO write csv record
		return strings.Join(keyValues, ini.sliceSep), nil
	case reflect.Map:
		keys := value.MapKeys()
		keyValues := make([]string, len(keys))
		for i, key := range keys {
			v := value.MapIndex(key)
			w, isTexter := getMarshalTexter(v)
			s, err := ini.encodeValue(v, w, isTexter)
			if err != nil {
				return "", err
			}
			keyValues[i] = fmt.Sprintf("%v%s%s", key.Interface(), ini.mapkeySep, s)
		}
		//TODO write csv record
		return strings.Join(keyValues, ini.sliceSep), nil
	}
	return fmt.Sprintf("%v", value.Interface()), nil
}

// Figure out the key and section to look for in Ini.
// If the field is to be ignored, the returned key name is empty.
// Otherwise, if it is not specified, the field name is used as the key.
// A struct tag may contain 3 entries:
//  - the key name (defaults to the field name)
//  - the section name (defaults to the global section)
//  - whether the key is the last of a block, which introduces a newline
func getTagInfo(field reflect.StructField) (section, key string, isLastKey bool) {
	tag := field.Tag.Get("ini")
	if tag == "" {
		key = field.Name
		return
	}
	if tag[0] == '-' {
		return
	}
	lst := strings.Split(tag, ",")
	n := len(lst)
	if n > 0 {
		key = lst[0]
		if key == "" {
			key = field.Name
		}
	}
	if n > 1 {
		section = lst[1]
	}
	if n > 2 {
		isLastKey, _ = strconv.ParseBool(lst[2])
	}
	return
}

// getMarshalTexter returns the pointer to the Value and whether it
// implements the TextMarshaler interface.
func getMarshalTexter(v reflect.Value) (reflect.Value, bool) {
	p := v
	if v.Kind() != reflect.Ptr && v.CanAddr() {
		p = v.Addr()
	}
	return p, p.Type().Implements(textMarshalType)
}
