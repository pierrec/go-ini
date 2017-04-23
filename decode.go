package ini

import (
	"encoding"
	"errors"
	"fmt"
	"io"
	"reflect"
	"time"

	"github.com/pierrec/go-ini/internal/structs"
)

var (
	errDecodeNoPtr    = errors.New("ini: decode value must be a pointer")
	errDecodeNoStruct = errors.New("ini: decode value must be a pointer to a struct")
	errInvalidMapKey  = errors.New("ini: invalid map key")
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
	root, err := structs.NewStruct(v, iniTagID)
	if err != nil {
		return err
	}

	for _, field := range root.Fields() {
		section, key, _ := getTagInfo(field.Tag(), field.Name())
		if section == "" {
			section = defaultSection
		}

		if emb := field.Embedded(); emb != nil {
			if defaultSection != "" {
				// Only process the first level of embedded types.
				continue
			}
			if section == "" {
				section = field.Name()
			}
			if err := ini.decode(section, emb); err != nil {
				return fmt.Errorf("ini: decode: %s.%s: %v", section, key, err)
			}
			continue
		}

		keyValuePtr := ini.get(section, key)
		if keyValuePtr == nil {
			// Not found.
			continue
		}

		// The value was found. Try to convert it to the field type.
		if err := field.Set(*keyValuePtr, ini.sliceSep, ini.mapkeySep); err != nil {
			return fmt.Errorf("ini: decode: %s.%s: %v", section, key, err)
		}
	}

	return nil
}
