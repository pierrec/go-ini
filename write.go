package ini

import (
	"fmt"
	"io"
)

// WriteTo writes the contents of Ini to the given Writer.
func (ini *INI) WriteTo(w io.Writer) (int64, error) {
	var written int64

	// Global section.
	m, err := ini.printSection(w, &ini.global)
	written += int64(m)
	if err != nil {
		return written, err
	}
	first := len(ini.global.Data) > 0

	for i, section := range ini.sections {
		if i > 0 || first {
			m, err := fmt.Fprintf(w, "\n")
			written += int64(m)
			if err != nil {
				return written, err
			}
		}
		m, err := ini.printSection(w, section)
		written += int64(m)
		if err != nil {
			return written, err
		}
	}

	return written, nil
}

func (ini *INI) printComments(w io.Writer, comments []string) (int, error) {
	var written int
	for _, s := range comments {
		n, err := fmt.Fprintf(w, "%s%s\n", ini.comment, s)
		written += n
		if err != nil {
			return written, err
		}
	}
	return written, nil
}

func (ini *INI) printSection(w io.Writer, section *iniSection) (int, error) {
	var written int

	n, err := ini.printComments(w, section.Comments)
	written += n
	if err != nil {
		return written, err
	}

	isGlobal := section.Name == ""
	if !isGlobal {
		n, err := fmt.Fprintf(w, "[%s]\n", section.Name)
		written += n
		if err != nil {
			return written, err
		}
	} else if n > 0 {
		// Newline separating the global section comments from its keys.
		n, err := fmt.Fprintf(w, "\n")
		written += n
		if err != nil {
			return written, err
		}
	}

	// The items may be separated by a single newline.
	// The items slice never starts with a newline.
	for items := section.Data; len(items) > 0; {
		// Block of items separated by a newline.
		var block []*iniItem
		for i, item := range items {
			if item == nil {
				block = items[:i]
				items = items[i+1:]
				break
			}
		}
		if block == nil {
			block = items
			items = nil
		}

		// Find the longest key.
		var n int
		for _, item := range block {
			k := item.Key
			if len(k) > n {
				n = len(k)
			}
		}
		kvFmt := fmt.Sprintf("%%-%ds = %%s\n", n)

		// Print all items with the equal sign aligned for all keys of this block.
		for _, item := range block {
			n, err := ini.printComments(w, item.Comments)
			written += n
			if err != nil {
				return written, err
			}
			n, err = fmt.Fprintf(w, kvFmt, item.Key, item.Value)
			written += n
			if err != nil {
				return written, err
			}
		}

		// Separate blocks with a newline unless it is the last one.
		if len(items) > 0 {
			n, err := fmt.Fprintf(w, "\n")
			written += n
			if err != nil {
				return written, err
			}
		}
	}

	return written, nil
}
