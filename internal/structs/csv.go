package structs

import (
	"bytes"
	"encoding/csv"
	"strings"
)

func separators(seps []rune) (sliceSep, mapKeySep rune) {
	sliceSeparator := SliceSeparator
	if len(seps) > 0 {
		sliceSeparator = seps[0]
	}
	mapKeySeparator := MapKeySeparator
	if len(seps) > 1 {
		mapKeySeparator = seps[1]
	}
	return sliceSeparator, mapKeySeparator
}

func newcsvreadwriter(sep rune) *csvreadwriter {
	buf := bytes.NewBuffer(nil)
	return &csvreadwriter{sep: sep, buf: buf}
}

type csvreadwriter struct {
	sep rune
	buf *bytes.Buffer
	*csv.Reader
	*csv.Writer
}

// read converts the csv input string into a slice.
func (r *csvreadwriter) read(s string) ([]string, error) {
	if s == "" {
		return nil, nil
	}
	if !strings.ContainsRune(s, r.sep) {
		return []string{s}, nil
	}
	r.buf.Reset()
	if _, err := r.buf.WriteString(s); err != nil {
		return nil, err
	}
	if r.Reader == nil {
		rr := csv.NewReader(r.buf)
		rr.Comma = r.sep
		r.Reader = rr
	}
	return r.Reader.Read()
}

// write returns the input strings into a single string as a csv record.
func (r *csvreadwriter) write(s ...string) (string, error) {
	if len(s) == 0 {
		return "", nil
	}
	if len(s) == 1 {
		return s[0], nil
	}
	r.buf.Reset()
	if r.Writer == nil {
		w := csv.NewWriter(r.buf)
		w.Comma = r.sep
		r.Writer = w
	}
	if err := r.Writer.Write(s); err != nil {
		return "", err
	}
	r.Writer.Flush()

	bts := r.buf.Bytes()

	// Remove the trailing newline.
	return string(bts[:len(bts)-1]), nil
}
