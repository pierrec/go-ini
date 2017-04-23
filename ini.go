package ini

import (
	"io"
	"strings"
)

const (
	// DefaultComment is the default value used to prefix comments.
	DefaultComment = ";"
	// DefaultSliceSeparator is the default slice separator used to decode and encode slices.
	DefaultSliceSeparator = ','
	// DefaultMapKeySeparator is the default map key separator used to decode and encode slices.
	DefaultMapKeySeparator = ':'
)

// DefaultOptions lists the Options for the Encode and Decode functions to use.
var DefaultOptions []Option

const (
	iniTagID      = "ini"
	mergeSections = 1 + iota
	mergeSectionsWithComments
	mergeSectionsWithLastComments
)

var _ io.ReaderFrom = (*INI)(nil)
var _ io.WriterTo = (*INI)(nil)

// INI represents the content of an ini source.
type INI struct {
	comment         []byte
	isCaseSensitive bool
	mergeSections   int
	sliceSep        rune
	mapkeySep       rune

	// This is the global section, without a name.
	global iniSection

	// Named sections.
	sections []*iniSection
}

// New instantiates a new Ini type ready for parsing.
func New(options ...Option) (*INI, error) {
	ini := &INI{}
	for _, option := range options {
		if err := option(ini); err != nil {
			return nil, err
		}
	}

	if len(ini.comment) == 0 {
		ini.comment = []byte(DefaultComment)
	}
	if ini.sliceSep == 0 {
		ini.sliceSep = DefaultSliceSeparator
	}
	if ini.mapkeySep == 0 {
		ini.mapkeySep = DefaultMapKeySeparator
	}

	return ini, nil
}

// Reset clears all sections with their associated comments and keys.
// Initial Options are retained.
func (ini *INI) Reset() {
	ini.global = iniSection{}
	ini.sections = nil
}

func (ini *INI) getSection(section string) *iniSection {
	if section == "" {
		return &ini.global
	}

	section = ident(ini.isCaseSensitive, section)
	for _, s := range ini.sections {
		if ident(ini.isCaseSensitive, s.Name) == section {
			return s
		}
	}
	return nil
}

func (ini *INI) addSection(section string) *iniSection {
	sec := &iniSection{Name: section}
	ini.sections = append(ini.sections, sec)
	return sec
}

func (ini *INI) rmSection(section string) bool {
	section = ident(ini.isCaseSensitive, section)
	for i, s := range ini.sections {
		if ident(ini.isCaseSensitive, s.Name) == section {
			n := len(ini.sections) - 1
			copy(ini.sections[i:], ini.sections[i+1:])
			ini.sections[n] = nil
			ini.sections = ini.sections[:n]
			return true
		}
	}
	return false
}

// Has returns whether or not the section (if the key is empty) or
// the key exists for the given section.
func (ini *INI) Has(section, key string) bool {
	if key == "" {
		return ini.getSection(section) != nil
	}
	return ini.get(section, key) != nil
}

// Get fetches the key value in the given section.
// If the section or the key is not found an empty string is returned.
func (ini *INI) Get(section, key string) string {
	if v := ini.get(section, key); v != nil {
		return *v
	}
	return ""
}

func (ini *INI) get(section, key string) *string {
	return ini.getSection(section).get(key, ini.isCaseSensitive)
}

// GetComments gets the comments for the given section or key.
// Use an empty key to get the section comments.
func (ini *INI) GetComments(section, key string) []string {
	sec := ini.getSection(section)

	if sec == nil {
		return nil
	}
	if key == "" {
		return sec.Comments
	}

	if item := sec.getItem(key, ini.isCaseSensitive); item != nil {
		return item.Comments
	}

	return nil
}

// Set adds the key with its value to the given section.
// If the section does not exist it is created.
// Setting an empty key adds a newline for the next keys.
func (ini *INI) Set(section, key, value string) {
	sec := ini.getSection(section)
	if sec == nil {
		sec = ini.addSection(section)
	} else {
		// The section name may be different.
		sec.Name = section
	}

	if key == "" {
		// Only add the newline if it is the first one.
		if n := len(sec.Data); n > 0 && sec.Data[n-1] != nil {
			sec.Data = append(sec.Data, nil)
		}
		return
	}

	if item := sec.getItem(key, ini.isCaseSensitive); item != nil {
		// The key does exist.
		item.Key = key
		item.Value = value
		return
	}
	// The key does not exist.
	sec.Data = append(sec.Data, &iniItem{Key: key, Value: value})
}

// SetComments sets the comments for the given section or key.
// Use an empty key to set the section comments.
func (ini *INI) SetComments(section, key string, comments ...string) {
	sec := ini.getSection(section)

	if key == "" {
		if sec == nil {
			sec = ini.addSection(section)
		}
		sec.Comments = comments
		return
	}

	if item := sec.getItem(key, ini.isCaseSensitive); item != nil {
		item.Comments = comments
	}
}

// Sections returns the list of defined sections, excluding the global one.
func (ini *INI) Sections() []string {
	sections := make([]string, len(ini.sections))
	for i, s := range ini.sections {
		sections[i] = s.Name
	}
	return sections
}

// Keys returns the list of keys for the given section.
func (ini *INI) Keys(section string) []string {
	s := ini.getSection(section)
	if s == nil {
		return nil
	}

	keys := make([]string, len(s.Data))
	for i, items := range s.Data {
		var key string
		if items != nil {
			key = items.Key
		}
		keys[i] = key
	}
	return keys
}

// Del removes a section or key from Ini returning whether or not it did.
// Set the key to an empty string to remove a section.
func (ini *INI) Del(section, key string) bool {
	// Remove the section.
	if key == "" {
		if section == "" {
			ini.global = iniSection{}
			return true
		}

		return ini.rmSection(section)
	}

	// Remove the key for the section.
	return ini.getSection(section).rmItem(key, ini.isCaseSensitive)
}

// ident returns a lowercased identifier if required.
func ident(isCaseSensitive bool, s string) string {
	if isCaseSensitive {
		return s
	}
	return strings.ToLower(s)
}
