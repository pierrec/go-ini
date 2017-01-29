[![GoDoc](https://godoc.org/github.com/pierrec/go-ini?status.png)](https://godoc.org/github.com/pierrec/go-ini)
[![Build Status](https://travis-ci.org/pierrec/go-ini.svg?branch=master)](https://travis-ci.org/pierrec/go-ini)
[![Coverage Status](https://coveralls.io/repos/github/pierrec/go-ini/badge.svg?branch=master)](https://coveralls.io/github/pierrec/go-ini?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/pierrec/go-ini)](https://goreportcard.com/report/github.com/pierrec/go-ini)

# ini
--
    import "github.com/pierrec/go-ini"

Package ini provides parsing and pretty printing methods for ini config files
including comments for sections and keys. The ini data can also be loaded
from/to structures using struct tags.

Since there is not really a strict definition for the ini file format, this
implementation follows these rules:

    - a section name cannot be empty unless it is the global one
    - leading and trailing whitespaces for key names are ignored
    - leading whitespace for key values are ignored
    - all characters from the first non whitespace to the end of the line are
    accepted for a value, unless the value is single or double quoted
    - anything after a quoted value is ignored
    - section and key names are not case sensitive by default
    - in case of conflicting key names, only the last one is used
    - in case of conflicting section names, only the last one is considered
    by default. However, if specified during initialization, the keys of
    conflicting sections can be merged.

## Usage

```go
const (
	// DefaultComment is the default value used to prefix comments.
	DefaultComment = ';'
	// DefaultSliceSeparator is the default slice separator used to decode and encode slices.
	DefaultSliceSeparator = ","
)
```

```go
var DefaultOptions []Option
```
DefaultOptions lists the Options for the Encode and Decode functions to use.

#### func  Decode

```go
func Decode(r io.Reader, v interface{}) error
```
Decode populates v with the Ini values from the Reader. DefaultOptions are used.
See Ini.Decode() for more information.

#### func  Encode

```go
func Encode(w io.Writer, v interface{}) error
```
Encode writes the contents of v to the given Writer. DefaultOptions are used.
See Ini.Encode() for more information.

#### type Ini

```go
type Ini struct {
}
```

Ini represents the content of an ini source.

#### func  New

```go
func New(options ...Option) (*Ini, error)
```
New instantiates a new Ini type ready for parsing.

#### func (*Ini) Decode

```go
func (ini *Ini) Decode(v interface{}) error
```
Decode decodes the Ini values into v, which must be a pointer to a struct. If
the struct field tag has not defined the key name then the name of the field is
used. The Ini section is defined as the second item in the struct tag. Supported
types for the struct fields are:

    - types implementing the encoding.TextUnmarshaler interface
    - all signed and unsigned integers
    - float32 and float64
    - string
    - bool
    - time.Time and time.Duration
    - slices of the above types

#### func (*Ini) Del

```go
func (ini *Ini) Del(section, key string) bool
```
Del removes a section or key from Ini returning whether or not it did. Set the
key to an empty string to remove a section.

#### func (*Ini) Encode

```go
func (ini *Ini) Encode(v interface{}) error
```
Encode sets Ini sections and keys according to the values defined in v. v must
be a pointer to a struct.

#### func (*Ini) Get

```go
func (ini *Ini) Get(section, key string) string
```
Get fetches the key value in the given section. If the section or the key is not
found an empty string is returned.

#### func (*Ini) GetComments

```go
func (ini *Ini) GetComments(section, key string) []string
```
GetComments gets the comments for the given section or key. Use an empty key to
get the section comments.

#### func (*Ini) Keys

```go
func (ini *Ini) Keys(section string) []string
```
Keys returns the list of keys for the given section.

#### func (*Ini) ReadFrom

```go
func (ini *Ini) ReadFrom(r io.Reader) (int64, error)
```
ReadFrom populates Ini with the data read from the reader. Leading and trailing
whitespaces for the key names are removed. Leading whitespaces for key values
are removed. If multiple sections have the same name, by default, the last one
is used. This can be overriden with the MergeSections option.

#### func (*Ini) Reset

```go
func (ini *Ini) Reset()
```
Reset clears all sections with their associated comments and keys. Initial
Options are retained.

#### func (*Ini) Sections

```go
func (ini *Ini) Sections() []string
```
Sections returns the list of defined sections, excluding the global one.

#### func (*Ini) Set

```go
func (ini *Ini) Set(section, key, value string)
```
Set adds the key with its value to the given section. If the section does not
exist it is created. Setting an empty key adds a newline for the next keys.

#### func (*Ini) SetComments

```go
func (ini *Ini) SetComments(section, key string, comments ...string)
```
SetComments sets the comments for the given section or key. Use an empty key to
set the section comments.

#### func (*Ini) WriteTo

```go
func (ini *Ini) WriteTo(w io.Writer) (int64, error)
```
WriteTo writes the contents of Ini to the given Writer.

#### type Option

```go
type Option func(*Ini) error
```

Option allows setting various options when creating an Ini type.

#### func  CaseSensitive

```go
func CaseSensitive() Option
```
CaseSensitive makes section and key names case sensitive when using the Get() or
Decode() methods.

#### func  Comment

```go
func Comment(prefix rune) Option
```
Comment sets the comment character. It defaults to ';'.

#### func  MergeSections

```go
func MergeSections() Option
```
MergeSections merges sections when multiple ones are defined instead of
overwriting them, in which case the last one wins. This is only relevant when
the Ini is being initialized by ReadFrom.

#### func  SliceSeparator

```go
func SliceSeparator(sep string) Option
```
SliceSeparator defines the separator used to split strings when decoding into a
slice or encoding a slice into a key value.
