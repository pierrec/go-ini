package ini

// Option allows setting various options when creating an Ini type.
type Option func(*INI) error

// Comment sets the comment character.
// It defaults to ";".
func Comment(prefix string) Option {
	return func(ini *INI) error {
		ini.comment = []byte(prefix)
		return nil
	}
}

// CaseSensitive makes section and key names case sensitive
// when using the Get() or Decode() methods.
func CaseSensitive() Option {
	return func(ini *INI) error {
		ini.isCaseSensitive = true
		return nil
	}
}

// MergeSections merges sections when multiple ones are defined
// instead of overwriting them, in which case the last one wins.
// This is only relevant when the Ini is being initialized by ReadFrom.
func MergeSections() Option {
	return func(ini *INI) error {
		ini.mergeSections = mergeSections
		return nil
	}
}

// MergeSectionsWithComments is equivalent to MergeSections but all the
// section comments merged.
func MergeSectionsWithComments() Option {
	return func(ini *INI) error {
		ini.mergeSections = mergeSectionsWithComments
		return nil
	}
}

// MergeSectionsWithLastComments is equivalent to MergeSections but the
// section comments are set to the ones from the last section.
func MergeSectionsWithLastComments() Option {
	return func(ini *INI) error {
		ini.mergeSections = mergeSectionsWithLastComments
		return nil
	}
}

// SliceSeparator defines the separator used to split strings when
// decoding into a slice/map or encoding a slice/map into a key value.
func SliceSeparator(sep rune) Option {
	return func(ini *INI) error {
		ini.sliceSep = sep
		return nil
	}
}

// MapKeySeparator defines the separator used to split strings when
// decoding or encoding a map key.
func MapKeySeparator(sep rune) Option {
	return func(ini *INI) error {
		ini.mapkeySep = sep
		return nil
	}
}
