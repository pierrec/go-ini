

# ini
`import "."`

* [Overview](#pkg-overview)
* [Index](#pkg-index)
* [Examples](#pkg-examples)
* [Subdirectories](#pkg-subdirectories)

## <a name="pkg-overview">Overview</a>
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

Behaviour of INI processing can be modified using struct tags. The struct tags
are defined by the "ini" keyword. The struct tags format is:


	<key name>[,section name[,last key in a block]]

If a key name is '-' then the struct field is ignored.




## <a name="pkg-index">Index</a>
* [Constants](#pkg-constants)
* [Variables](#pkg-variables)
* [func Decode(r io.Reader, v interface{}) error](#Decode)
* [func Encode(w io.Writer, v interface{}) error](#Encode)
* [type INI](#INI)
  * [func New(options ...Option) (*INI, error)](#New)
  * [func (ini *INI) Decode(v interface{}) error](#INI.Decode)
  * [func (ini *INI) Del(section, key string) bool](#INI.Del)
  * [func (ini *INI) Encode(v interface{}) error](#INI.Encode)
  * [func (ini *INI) Get(section, key string) string](#INI.Get)
  * [func (ini *INI) GetComments(section, key string) []string](#INI.GetComments)
  * [func (ini *INI) Has(section, key string) bool](#INI.Has)
  * [func (ini *INI) Keys(section string) []string](#INI.Keys)
  * [func (ini *INI) ReadFrom(r io.Reader) (int64, error)](#INI.ReadFrom)
  * [func (ini *INI) Reset()](#INI.Reset)
  * [func (ini *INI) Sections() []string](#INI.Sections)
  * [func (ini *INI) Set(section, key, value string)](#INI.Set)
  * [func (ini *INI) SetComments(section, key string, comments ...string)](#INI.SetComments)
  * [func (ini *INI) WriteTo(w io.Writer) (int64, error)](#INI.WriteTo)
* [type Option](#Option)
  * [func CaseSensitive() Option](#CaseSensitive)
  * [func Comment(prefix string) Option](#Comment)
  * [func MapKeySeparator(sep rune) Option](#MapKeySeparator)
  * [func MergeSections() Option](#MergeSections)
  * [func MergeSectionsWithComments() Option](#MergeSectionsWithComments)
  * [func MergeSectionsWithLastComments() Option](#MergeSectionsWithLastComments)
  * [func SliceSeparator(sep rune) Option](#SliceSeparator)

#### <a name="pkg-examples">Examples</a>
* [Package](#example_)
* [Decode](#example_Decode)
* [Encode](#example_Encode)

#### <a name="pkg-files">Package files</a>
[decode.go](/src/target/decode.go) [doc.go](/src/target/doc.go) [encode.go](/src/target/encode.go) [ini.go](/src/target/ini.go) [options.go](/src/target/options.go) [read.go](/src/target/read.go) [section.go](/src/target/section.go) [write.go](/src/target/write.go) 


## <a name="pkg-constants">Constants</a>
``` go
const (
    // DefaultComment is the default value used to prefix comments.
    DefaultComment = ";"
    // DefaultSliceSeparator is the default slice separator used to decode and encode slices.
    DefaultSliceSeparator = ','
    // DefaultMapKeySeparator is the default map key separator used to decode and encode slices.
    DefaultMapKeySeparator = ':'
)
```

## <a name="pkg-variables">Variables</a>
``` go
var DefaultOptions []Option
```
DefaultOptions lists the Options for the Encode and Decode functions to use.



## <a name="Decode">func</a> [Decode](/src/target/decode.go?s=677:722#L20)
``` go
func Decode(r io.Reader, v interface{}) error
```
Decode populates v with the Ini values from the Reader.
DefaultOptions are used.
See Ini.Decode() for more information.



## <a name="Encode">func</a> [Encode](/src/target/encode.go?s=333:378#L9)
``` go
func Encode(w io.Writer, v interface{}) error
```
Encode writes the contents of v to the given Writer.
DefaultOptions are used.
See Ini.Encode() for more information.




## <a name="INI">type</a> [INI](/src/target/ini.go?s=726:969#L21)
``` go
type INI struct {
    // contains filtered or unexported fields
}
```
INI represents the content of an ini source.







### <a name="New">func</a> [New](/src/target/ini.go?s=1025:1066#L36)
``` go
func New(options ...Option) (*INI, error)
```
New instantiates a new Ini type ready for parsing.





### <a name="INI.Decode">func</a> (\*INI) [Decode](/src/target/decode.go?s=1385:1428#L43)
``` go
func (ini *INI) Decode(v interface{}) error
```
Decode decodes the Ini values into v, which must be a pointer to a struct.
If the struct field tag has not defined the key name
then the name of the field is used.
The Ini section is defined as the second item in the struct tag.
Supported types for the struct fields are:


	- types implementing the encoding.TextUnmarshaler interface
	- all signed and unsigned integers
	- float32 and float64
	- string
	- bool
	- time.Time and time.Duration
	- slices of the above types




### <a name="INI.Del">func</a> (\*INI) [Del](/src/target/ini.go?s=5270:5315#L216)
``` go
func (ini *INI) Del(section, key string) bool
```
Del removes a section or key from Ini returning whether or not it did.
Set the key to an empty string to remove a section.




### <a name="INI.Encode">func</a> (\*INI) [Encode](/src/target/encode.go?s=655:698#L23)
``` go
func (ini *INI) Encode(v interface{}) error
```
Encode sets Ini sections and keys according to the values defined in v.
v must be a pointer to a struct.




### <a name="INI.Get">func</a> (\*INI) [Get](/src/target/ini.go?s=2742:2789#L109)
``` go
func (ini *INI) Get(section, key string) string
```
Get fetches the key value in the given section.
If the section or the key is not found an empty string is returned.




### <a name="INI.GetComments">func</a> (\*INI) [GetComments](/src/target/ini.go?s=3092:3149#L122)
``` go
func (ini *INI) GetComments(section, key string) []string
```
GetComments gets the comments for the given section or key.
Use an empty key to get the section comments.




### <a name="INI.Has">func</a> (\*INI) [Has](/src/target/ini.go?s=2473:2518#L100)
``` go
func (ini *INI) Has(section, key string) bool
```
Has returns whether or not the section (if the key is empty) or
the key exists for the given section.




### <a name="INI.Keys">func</a> (\*INI) [Keys](/src/target/ini.go?s=4867:4912#L197)
``` go
func (ini *INI) Keys(section string) []string
```
Keys returns the list of keys for the given section.




### <a name="INI.ReadFrom">func</a> (\*INI) [ReadFrom](/src/target/read.go?s=530:582#L12)
``` go
func (ini *INI) ReadFrom(r io.Reader) (int64, error)
```
ReadFrom populates Ini with the data read from the reader.
Leading and trailing whitespaces for the key names are removed.
Leading whitespaces for key values are removed.
If multiple sections have the same name, by default, the last
one is used. This can be overridden with the MergeSections option.




### <a name="INI.Reset">func</a> (\*INI) [Reset](/src/target/ini.go?s=1512:1535#L59)
``` go
func (ini *INI) Reset()
```
Reset clears all sections with their associated comments and keys.
Initial Options are retained.




### <a name="INI.Sections">func</a> (\*INI) [Sections](/src/target/ini.go?s=4646:4681#L188)
``` go
func (ini *INI) Sections() []string
```
Sections returns the list of defined sections, excluding the global one.




### <a name="INI.Set">func</a> (\*INI) [Set](/src/target/ini.go?s=3530:3577#L142)
``` go
func (ini *INI) Set(section, key, value string)
```
Set adds the key with its value to the given section.
If the section does not exist it is created.
Setting an empty key adds a newline for the next keys.




### <a name="INI.SetComments">func</a> (\*INI) [SetComments](/src/target/ini.go?s=4258:4326#L171)
``` go
func (ini *INI) SetComments(section, key string, comments ...string)
```
SetComments sets the comments for the given section or key.
Use an empty key to set the section comments.




### <a name="INI.WriteTo">func</a> (\*INI) [WriteTo](/src/target/write.go?s=97:148#L1)
``` go
func (ini *INI) WriteTo(w io.Writer) (int64, error)
```
WriteTo writes the contents of Ini to the given Writer.




## <a name="Option">type</a> [Option](/src/target/options.go?s=81:109#L1)
``` go
type Option func(*INI) error
```
Option allows setting various options when creating an Ini type.







### <a name="CaseSensitive">func</a> [CaseSensitive](/src/target/options.go?s=396:423#L7)
``` go
func CaseSensitive() Option
```
CaseSensitive makes section and key names case sensitive
when using the Get() or Decode() methods.


### <a name="Comment">func</a> [Comment](/src/target/options.go?s=173:207#L1)
``` go
func Comment(prefix string) Option
```
Comment sets the comment character.
It defaults to ";".


### <a name="MapKeySeparator">func</a> [MapKeySeparator](/src/target/options.go?s=1696:1733#L53)
``` go
func MapKeySeparator(sep rune) Option
```
MapKeySeparator defines the separator used to split strings when
decoding or encoding a map key.


### <a name="MergeSections">func</a> [MergeSections](/src/target/options.go?s=706:733#L17)
``` go
func MergeSections() Option
```
MergeSections merges sections when multiple ones are defined
instead of overwriting them, in which case the last one wins.
This is only relevant when the Ini is being initialized by ReadFrom.


### <a name="MergeSectionsWithComments">func</a> [MergeSectionsWithComments](/src/target/options.go?s=922:961#L26)
``` go
func MergeSectionsWithComments() Option
```
MergeSectionsWithComments is equivalent to MergeSections but all the
section comments merged.


### <a name="MergeSectionsWithLastComments">func</a> [MergeSectionsWithLastComments](/src/target/options.go?s=1197:1240#L35)
``` go
func MergeSectionsWithLastComments() Option
```
MergeSectionsWithLastComments is equivalent to MergeSections but the
section comments are set to the ones from the last section.


### <a name="SliceSeparator">func</a> [SliceSeparator](/src/target/options.go?s=1483:1519#L44)
``` go
func SliceSeparator(sep rune) Option
```
SliceSeparator defines the separator used to split strings when
decoding into a slice/map or encoding a slice/map into a key value.









- - -
Generated by [godoc2md](http://godoc.org/github.com/davecheney/godoc2md)
