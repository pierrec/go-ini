package structs

import (
	"fmt"
	htemplate "html/template"
	"net"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/kr/pretty"
)

const (
	// SliceSeparator is used to separate slice and map items.
	SliceSeparator = ','

	// MapKeySeparator is used to separate map keys and their value.
	MapKeySeparator = ':'
)

var (
	errNoStruct        = fmt.Errorf("not a struct")
	errNoPointer       = fmt.Errorf("not a pointer")
	errCannotUnmarshal = fmt.Errorf("cannot unmarshal value")
	errInvalidMapKey   = fmt.Errorf("invalid map key")
	errCannotSet       = fmt.Errorf("cannot set value")
)

// Supported types.
var (
	durationType     = reflect.TypeOf(time.Second)
	timeType         = reflect.TypeOf(time.Time{})
	urlType          = reflect.TypeOf(new(url.URL))
	texttemplateType = reflect.TypeOf(template.New(""))
	htmltemplateType = reflect.TypeOf(htemplate.New(""))
	regexpType       = reflect.TypeOf(regexp.MustCompile("."))
	ipaddrType       = reflect.TypeOf(new(net.IPAddr))
	ipnetType        = reflect.TypeOf(new(net.IPNet))
)

// NewStruct recursively decomposes the input struct into its fields
// and embedded structs.
// Fields tags with "-" will be skipped.
// Fields tags with a non empty value will be renamed to that value.
//
// The input must be a pointer to a struct.
func NewStruct(s interface{}, tagid string) (*StructStruct, error) {
	if s, ok := s.(*StructStruct); ok {
		return s, nil
	}

	v := reflect.ValueOf(s)
	if v.Kind() != reflect.Ptr {
		return nil, errNoPointer
	}
	if v.Elem().Kind() != reflect.Struct {
		return nil, errNoStruct
	}

	return &StructStruct{
		name:  fmt.Sprintf("%T", s),
		raw:   s,
		value: v,
		data:  fieldsOf(s, tagid),
	}, nil
}

// StructField represents a struct field.
type StructField struct {
	name     string
	field    *reflect.StructField
	value    reflect.Value
	tag      reflect.StructTag
	embedded *StructStruct
}

// Name returns the field name.
func (f *StructField) Name() string {
	return f.name
}

// Embedded returns the embedded struct if the field is embedded.
func (f *StructField) Embedded() *StructStruct {
	return f.embedded
}

// Set assigns the given value to the field.
// If the value is a string but the field is not,
// then its value is deserialized using encoding.Unmarshaler
// or in a best effort way.
func (f *StructField) Set(v interface{}, seps ...rune) error {
	return Set(f.value, v, seps...)
}

// Value returns the interface value of the field.
func (f *StructField) Value() interface{} {
	return f.value.Interface()
}

// PtrValue returns the interface pointer value of the field.
func (f *StructField) PtrValue() interface{} {
	return f.value.Addr().Interface()
}

// Tag returns the tags defined on the field.
func (f *StructField) Tag() reflect.StructTag {
	return f.tag
}

// StructStruct represents a decomposed struct.
type StructStruct struct {
	name  string
	raw   interface{}
	value reflect.Value
	data  []*StructField
}

// Name returns the underlying type name.
func (s *StructStruct) Name() string {
	return s.name
}

// GoString is used to debug a StructStruct and returns a full
// and human readable representation of its elements.
func (s *StructStruct) GoString() string {
	return pretty.Sprint(s)
}

// String gives a simple string representation of the StructStruct.
func (s *StructStruct) String() string {
	return s.string(0)
}

// n: field padding
func (s *StructStruct) string(n int) string {
	sname := s.Name()
	pad := strings.Repeat(" ", n)

	var res string
	res += fmt.Sprintf("%s%s {\n", pad, sname)

	var fn int
	for _, field := range s.data {
		var n int
		if emb := field.Embedded(); emb != nil {
			n = len(emb.Name())
		} else {
			n = len(field.Name())
		}
		if n > fn {
			fn = n
		}
	}

	f := fmt.Sprintf("%s%%%ds %%T\n", pad, fn+1)
	for _, field := range s.data {
		if emb := field.Embedded(); emb != nil {
			res += emb.string(n + fn)
			continue
		}
		res += fmt.Sprintf(f, field.Name(), field.value.Interface())
	}

	res += fmt.Sprintf("%s}\n", pad)
	return res
}

// Lookup returns the field for the corresponding path.
func (s *StructStruct) Lookup(path ...string) *StructField {
	name := path[0]
	if len(path) == 1 {
		for _, item := range s.data {
			if item.Name() == name {
				return item
			}
		}
		return nil
	}
	for _, item := range s.data {
		if item.embedded != nil && item.Name() == name {
			return item.embedded.Lookup(path[1:]...)
		}
	}
	return nil
}

// Fields returns all the fields of the parsed struct.
func (s *StructStruct) Fields() []*StructField {
	return s.data
}

// CallUntil recursively calls the given method on its StructStruct fields
// and stops at the first one satisfying the stop condition.
func (s *StructStruct) CallUntil(m string, args []interface{}, until func([]interface{}) bool) ([]interface{}, bool) {
	res, ok := s.Call(m, args)
	if ok && until(res) {
		return res, true
	}
	for _, item := range s.data {
		if item.embedded == nil {
			continue
		}
		res, ok := item.embedded.CallUntil(m, args, until)
		if ok && until(res) {
			return res, true
		}
	}
	return nil, false
}

// Call invokes the method m on s with arguments args.
//
// It returns the method results and whether is was invoked successfully.
func (s *StructStruct) Call(m string, args []interface{}) ([]interface{}, bool) {
	fn := s.value.MethodByName(m)
	if !fn.IsValid() {
		if s.value.CanAddr() {
			// Try with a pointer receiver.
			fn = s.value.Addr().MethodByName(m)
		}
		if !fn.IsValid() {
			return nil, false
		}
	}
	values := make([]reflect.Value, len(args))
	for i, arg := range args {
		values[i] = reflect.ValueOf(arg)
	}
	rvalues := fn.Call(values)
	results := make([]interface{}, len(rvalues))
	for i, rv := range rvalues {
		results[i] = rv.Interface()
	}
	return results, true
}

// List the fields of the input which must be a pointer to a struct.
func fieldsOf(v interface{}, tagid string) (res []*StructField) {
	value := reflect.ValueOf(v).Elem()
	vType := value.Type()
	for i, n := 0, value.NumField(); i < n; i++ {
		value := value.Field(i)
		if !value.CanSet() {
			// Cannot set the field, maybe unexported.
			continue
		}
		field := vType.Field(i)
		fname := field.Name

		tag := field.Tag
		tagval := tag.Get(tagid)
		// The name is the first item in a coma separated list.
		switch v := strings.SplitN(tagval, ",", 2); v[0] {
		case "":
		case "-":
			continue
		default:
			// Set the field name according to the struct tag.
			fname = v[0]
		}

		var fs *StructStruct
		switch kind := value.Kind(); kind {
		case reflect.Invalid,
			reflect.Complex64, reflect.Complex128,
			reflect.Chan, reflect.Func, reflect.Interface,
			reflect.UnsafePointer:
			// Unsupported field types.
			continue
		case reflect.Struct:
			if field.Anonymous {
				// Embedded field: recursively descend into its fields.
				v := value.Addr().Interface()
				fs = &StructStruct{fname, v, value, fieldsOf(v, tagid)}
			}
		}
		res = append(res, &StructField{fname, &field, value, tag, fs})
	}
	return
}
