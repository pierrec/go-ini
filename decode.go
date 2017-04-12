package ini

import (
	"encoding"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	errDecodeNoPtr    = errors.New("ini: decode value must be a pointer")
	errDecodeNoStruct = errors.New("ini: decode value must be a pointer to a struct")
)

// Special struct field types.
var (
	durationType = reflect.TypeOf(time.Second)
	// Get the reflect type of an interface.
	textUnmarshalType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
)

// Decode populates v with the Ini values from the Reader.
// DefaultOptions are used.
// See Ini.Decode() for more information.
func Decode(r io.Reader, v interface{}) error {
	ini, err := New(DefaultOptions...)
	if err != nil {
		return err
	}
	if _, err := ini.ReadFrom(r); err != nil {
		return err
	}
	return ini.Decode(v)
}

// Decode decodes the Ini values into v, which must be a pointer to a struct.
// If the struct field tag has not defined the key name
// then the name of the field is used.
// The Ini section is defined as the second item in the struct tag.
// Supported types for the struct fields are:
//  - types implementing the encoding.TextUnmarshaler interface
//  - all signed and unsigned integers
//  - float32 and float64
//  - string
//  - bool
//  - time.Time and time.Duration
//  - slices of the above types
func (ini *INI) Decode(v interface{}) error {
	return ini.decode("", v)
}

func (ini *INI) decode(defaultSection string, v interface{}) error {
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

		// If the field implements encoding.TextMarshaler, it prevails.
		valuePtr, isTexter := getUnmarshalTexter(value)
		if !isTexter {
			switch field.Type.Kind() {
			case reflect.Bool,
				reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
				reflect.Float32, reflect.Float64,
				reflect.String,
				reflect.Slice:
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
				err := ini.decode(section, valuePtr.Interface())
				if err != nil {
					return err
				}
				continue
			default:
				continue
			}
		}

		// Lookup the section and key value.
		section, key, _ := getTagInfo(field)
		if key == "" {
			// Omit the field.
			continue
		}
		if section == "" {
			section = defaultSection
		}
		keyValuePtr := ini.get(section, key)
		if keyValuePtr == nil {
			// Not found.
			continue
		}

		// The value was found. Try to convert it to the field type.
		// Special case: texter.
		// NB. time.Time implements encoding.TextUnmarshaler.
		if err := ini.decodeValue(value, valuePtr, isTexter, keyValuePtr); err != nil {
			return fmt.Errorf("ini: decode: %s.%s: %v", section, key, err)
		}
	}

	return nil
}

// decodeValue sets value to the keyValuePtr value.
func (ini *INI) decodeValue(value, valuePtr reflect.Value, isTexter bool, keyValuePtr *string) error {
	if isTexter {
		bts := []byte(*keyValuePtr)
		txtValue := reflect.ValueOf(bts)
		args := []reflect.Value{txtValue}
		vals := valuePtr.MethodByName("UnmarshalText").Call(args)
		if v := vals[0]; !v.IsNil() {
			return fmt.Errorf("%v", v.Interface())
		}
		return nil
	}

	// Special cases: time.Duration.
	valueType := value.Type()
	switch valueType {
	case durationType:
		d, err := time.ParseDuration(*keyValuePtr)
		if err != nil {
			return err
		}
		value.SetInt(int64(d))
		return nil
	}

	switch valueType.Kind() {
	case reflect.Bool:
		v, err := strconv.ParseBool(*keyValuePtr)
		if err != nil {
			return err
		}
		value.SetBool(v)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(*keyValuePtr, 0, 64)
		if err != nil {
			return err
		}
		value.SetInt(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(*keyValuePtr, 0, 64)
		if err != nil {
			return err
		}
		value.SetUint(v)
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(*keyValuePtr, 64)
		if err != nil {
			return err
		}
		value.SetFloat(v)
	case reflect.String:
		value.SetString(*keyValuePtr)
	case reflect.Slice:
		var keyValues []string
		if *keyValuePtr != "" {
			keyValues = strings.Split(*keyValuePtr, ini.sliceSep)
		}
		elem := valueType.Elem()
		sliceValues := reflect.MakeSlice(valueType, 0, len(keyValues))
		for i := range keyValues {
			v := reflect.New(elem).Elem()
			_, isTexter = getUnmarshalTexter(v)
			err := ini.decodeValue(v, v.Addr(), isTexter, &keyValues[i])
			if err != nil {
				return fmt.Errorf("%v at index %d", err, i)
			}
			sliceValues = reflect.Append(sliceValues, v)
		}
		value.Set(sliceValues)
	}
	return nil
}

// getUnmarshalTexter returns the pointer to the Value and whether it
// implements the TextUnmarshaler interface.
func getUnmarshalTexter(v reflect.Value) (reflect.Value, bool) {
	p := v
	if v.Kind() != reflect.Ptr {
		p = v.Addr()
	}
	return p, p.Type().Implements(textUnmarshalType)
}
