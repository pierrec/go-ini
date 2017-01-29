package ini

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"unicode"
	"unicode/utf8"
)

var (
	errInvalidSectionName = errors.New("invalid section name")
	errInvalidKeyValue    = errors.New("string literal not terminated")
)

// ReadFrom populates Ini with the data read from the reader.
// Leading and trailing whitespaces for the key names are removed.
// Leading whitespaces for key values are removed.
// If multiple sections have the same name, by default, the last
// one is used. This can be overriden with the MergeSections option.
func (ini *Ini) ReadFrom(r io.Reader) (int64, error) {
	var (
		read int64
		s    = bufio.NewReader(r)
		// Current line number.
		lineNum = 0
		// Comments currently parsed.
		// They are valid for the next element (Section or Item) or global.
		comments []string
		// Current block.
		items   []*iniItem
		current *iniSection
	)
	for {
		// Parse the current line.
		lineNum++
		line, err := s.ReadBytes('\n')
		read += int64(len(line))
		if err != nil {
			if err != io.EOF {
				return read, err
			}
			// There is potentially data along the io.EOF error.
			// Ignore the error until there is no more data.
			if len(line) == 0 {
				ini.addItemsToSection(items, current)
				return read, nil
			}
		}
		// Remove trailing newline.
		line = stripNewline(line)
		// Ignore leading whitespace for the key name.
		line = bytes.TrimLeftFunc(line, unicode.IsSpace)

		if len(line) == 0 {
			// Empty line is ignored unless used to separate:
			// general section comments
			// blocks of kvp
			if n := len(ini.sections); n > 0 {
				ini.addItemsToSection(items, current)
				items = nil
			} else if ini.global == nil && len(comments) > 0 {
				ini.global = &iniSection{Comments: comments}
				comments = nil
				current = ini.global
			}
			continue
		}

		if line[0] == '[' {
			// Section.
			i := bytes.IndexByte(line, ']')
			if i < 0 {
				return read, fmt.Errorf("ini: %d: missing ]", lineNum)
			}
			name := string(line[1:i])
			if name == "" {
				return read, errInvalidSectionName
			}

			if !ini.mergeSections {
				// Remove any previous section with the same name.
				ini.rmSection(name)
			} else if section := ini.getSection(name); section != nil {
				// Sections are merged: the new section comments are ignored.
				comments = nil

				ini.addItemsToSection(items, current)
				items = nil

				current = section
				continue
			}

			section := &iniSection{
				Comments: comments,
				Name:     name,
			}
			comments = nil

			ini.addItemsToSection(items, current)
			items = nil

			ini.sections = append(ini.sections, section)
			current = section
			continue
		}

		if first, _ := utf8.DecodeRune(line); ini.comment == first {
			// Comment.
			comments = append(comments, string(line[1:]))
			continue
		}

		// Key/Value pair.
		i := bytes.IndexByte(line, '=')
		if i < 0 {
			return read, fmt.Errorf("ini: %d: missing =", lineNum)
		}
		// Ignore trailing whitespace for the key name.
		key := string(bytes.TrimRightFunc(line[:i], unicode.IsSpace))

		// Ignore leading whitespace for the value.
		valueBytes := bytes.TrimLeftFunc(line[i+1:], unicode.IsSpace)
		valueBytes, err = scanString(valueBytes)
		if err != nil {
			return read, fmt.Errorf("ini: %d: %v", lineNum, err)
		}
		value := string(valueBytes)

		// Deduplicate keys.
		for i, item := range items {
			if item.Key == key {
				n := len(items) - 1
				copy(items[i:], items[i+1:])
				items[n] = nil
				items = items[:n]
			}
		}

		item := &iniItem{
			Comments: comments,
			Key:      key,
			Value:    value,
		}
		comments = nil
		items = append(items, item)
	}
}

func (ini *Ini) addItemsToSection(items []*iniItem, section *iniSection) {
	if len(items) == 0 {
		return
	}

	if section == nil {
		ini.global = &iniSection{}
		section = ini.global
	}

	// Keys and values.
	section.Data = dedupItems(section.Data, items)
	// Blank line.
	section.Data = append(section.Data, nil)
}

// dedupItems only deduplicates items between slices, not within the slices.
func dedupItems(a, b []*iniItem) []*iniItem {
	for i := 0; i < len(a); i++ {
		itemA := a[i]
		if itemA == nil {
			continue
		}
		for _, itemB := range b {
			if itemB == nil || itemA.Key != itemB.Key {
				continue
			}
			copy(a[i:], a[i+1:])
			n := len(a) - 1
			a[n] = nil
			a = a[:n]
			i--
			break
		}
	}

	return append(a, b...)
}

// scanString scans a string and handles quoted ones.
func scanString(buf []byte) ([]byte, error) {
	n := len(buf)
	if n == 0 || n == 1 {
		return buf, nil
	}
	// Is the string quoted?
	quote := buf[0]
	if quote != '"' && quote != '\'' {
		// Not quoted.
		return buf, nil
	}

	// Quoted.
	var escapers []int
	idx := 1
	for buf[idx] != quote {
		if c := buf[idx]; c == '\\' {
			escapers = append(escapers, idx)
			idx++
		}
		idx++
		if idx == n {
			return nil, errInvalidKeyValue
		}
	}
	buf = buf[1:idx]

	if len(escapers) == 0 {
		return buf, nil
	}

	// Remove escapers.
	for i, pos := range escapers {
		copy(buf[pos-i-1:], buf[pos-i:])
	}

	return buf[:len(buf)-len(escapers)], nil
}

// buf may end with \n or \r\n.
func stripNewline(buf []byte) []byte {
	if n := len(buf); n > 0 {
		if n > 1 && buf[n-2] == '\r' {
			return buf[:n-2]
		}
		if buf[n-1] == '\n' {
			return buf[:n-1]
		}
	}
	return buf
}
