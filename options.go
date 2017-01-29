package ini

// Option allows setting various options when creating an Ini type.
type Option func(*Ini) error

// Comment sets the comment character.
// It defaults to ';'.
func Comment(prefix rune) Option {
	return func(ini *Ini) error {
		ini.comment = prefix
		return nil
	}
}

// CaseSensitive makes section and key names case sensitive
// when using the Get() or Decode() methods.
func CaseSensitive() Option {
	return func(ini *Ini) error {
		ini.isCaseSensitive = true
		return nil
	}
}

// MergeSections merges sections when multiple ones are defined
// instead of overwriting them, in which case the last one wins.
// This is only relevant when the Ini is being initialized by ReadFrom.
func MergeSections() Option {
	return func(ini *Ini) error {
		ini.mergeSections = true
		return nil
	}
}

// SliceSeparator defines the separator used to split strings when
// decoding into a slice or encoding a slice into a key value.
func SliceSeparator(sep string) Option {
	return func(ini *Ini) error {
		ini.sliceSep = sep
		return nil
	}
}
