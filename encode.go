package ini

import (
	"encoding"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/pierrec/go-ini/internal/structs"
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
	root, err := structs.NewStruct(v, iniTagID)
	if err != nil {
		return err
	}

	for _, field := range root.Fields() {
		section, key, isLastKey := getTagInfo(field.Tag(), field.Name())
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
			if err := ini.encode(section, emb); err != nil {
				return fmt.Errorf("ini: encode: %s.%s: %v", section, key, err)
			}
			continue
		}

		mvalue, err := structs.MarshalValue(field.Value(), ini.sliceSep, ini.mapkeySep)
		if err != nil {
			return fmt.Errorf("ini: encode: %s.%s: %v", section, key, err)
		}
		keyValue := fmt.Sprintf("%v", mvalue)
		ini.Set(section, key, keyValue)

		if isLastKey {
			ini.Set(section, "", "")
		}
	}

	return nil
}

// Figure out the key and section to look for in Ini.
// Otherwise, if it is not specified, the field name is used as the key.
// A struct tag may contain 3 entries:
//  - the key name (defaults to the field name)
//  - the section name (defaults to the global section)
//  - whether the key is the last of a block, which introduces a newline
func getTagInfo(tags reflect.StructTag, defaultKey string) (section, key string, isLastKey bool) {
	tag := tags.Get(iniTagID)
	if tag == "" {
		key = defaultKey
		return
	}
	lst := strings.Split(tag, ",")
	n := len(lst)
	if n > 0 {
		key = lst[0]
		if key == "" {
			key = defaultKey
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
