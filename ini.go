package ini

import (
	"io"
	"strings"
)

const (
	// DefaultComment is the default value used to prefix comments.
	DefaultComment = ';'
	// DefaultSliceSeparator is the default slice separator used to decode and encode slices.
	DefaultSliceSeparator = ","
)

// DefaultOptions lists the Options for the Encode and Decode functions to use.
var DefaultOptions []Option

var _ io.ReaderFrom = (*Ini)(nil)
var _ io.WriterTo = (*Ini)(nil)

// Ini represents the content of an ini source.
type Ini struct {
	comment         rune
	isCaseSensitive bool
	mergeSections   bool
	sliceSep        string

	// This is the global section, without a name.
	global *iniSection

	// Named sections.
	sections []*iniSection
}

// New instantiates a new Ini type ready for parsing.
func New(options ...Option) (*Ini, error) {
	ini := &Ini{}
	for _, option := range options {
		if err := option(ini); err != nil {
			return nil, err
		}
	}

	if ini.comment == 0 {
		ini.comment = DefaultComment
	}
	if ini.sliceSep == "" {
		ini.sliceSep = DefaultSliceSeparator
	}

	return ini, nil
}

// Reset clears all sections with their associated comments and keys.
// Initial Options are retained.
func (ini *Ini) Reset() {
	ini.global = nil
	ini.sections = nil
}

func (ini *Ini) getSection(section string) *iniSection {
	if section == "" {
		return ini.global
	}

	flag := ini.isCaseSensitive
	if !flag {
		section = strings.ToLower(section)
	}
	for _, s := range ini.sections {
		if (flag && s.Name == section) || (!flag && strings.ToLower(s.Name) == section) {
			return s
		}
	}
	return nil
}

func (ini *Ini) addSection(section string) *iniSection {
	sec := &iniSection{Name: section}
	if section == "" {
		ini.global = sec
	} else {
		ini.sections = append(ini.sections, sec)
	}
	return sec
}

func (ini *Ini) rmSection(section string) bool {
	flag := ini.isCaseSensitive
	if !flag {
		section = strings.ToLower(section)
	}
	for i, s := range ini.sections {
		if (flag && s.Name == section) || (!flag && strings.ToLower(s.Name) == section) {
			n := len(ini.sections) - 1
			copy(ini.sections[i:], ini.sections[i+1:])
			ini.sections[n] = nil
			ini.sections = ini.sections[:n]
			return true
		}
	}
	return false
}

// Get fetches the key value in the given section.
// If the section or the key is not found an empty string is returned.
func (ini *Ini) Get(section, key string) string {
	if v := ini.get(section, key); v != nil {
		return *v
	}
	return ""
}

func (ini *Ini) get(section, key string) *string {
	return ini.getSection(section).get(key, ini.isCaseSensitive)
}

// GetComments gets the comments for the given section or key.
// Use an empty key to get the section comments.
func (ini *Ini) GetComments(section, key string) []string {
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
func (ini *Ini) Set(section, key, value string) {
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
func (ini *Ini) SetComments(section, key string, comments ...string) {
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
func (ini *Ini) Sections() []string {
	sections := make([]string, len(ini.sections))
	for i, s := range ini.sections {
		sections[i] = s.Name
	}
	return sections
}

// Keys returns the list of keys for the given section.
func (ini *Ini) Keys(section string) []string {
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
func (ini *Ini) Del(section, key string) bool {
	// Remove the section.
	if key == "" {
		if section == "" {
			ini.global = nil
			return true
		}

		return ini.rmSection(section)
	}

	// Remove the key for the section.
	return ini.getSection(section).rmItem(key, ini.isCaseSensitive)
}
