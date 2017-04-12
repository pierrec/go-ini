package ini

// iniSection represents a named section for keys.
// It may have comments.
// The Section may contain identical keys.
type iniSection struct {
	Comments []string
	Name     string

	// Keys may be grouped together and separated by a blank line.
	// A blank line is represented by a nil *Item.
	Data []*iniItem
}

// flag indicates whether or not the search is case sensitive.
func (s *iniSection) get(key string, flag bool) *string {
	if item := s.getItem(key, flag); item != nil {
		return &item.Value
	}
	return nil
}

// flag indicates whether or not the search is case sensitive.
func (s *iniSection) getItem(key string, flag bool) *iniItem {
	if s == nil {
		return nil
	}
	key = ident(flag, key)

	for _, item := range s.Data {
		if item == nil {
			continue
		}
		if ident(flag, item.Key) == key {
			return item
		}
	}
	return nil
}

// flag indicates whether or not the search is case sensitive.
func (s *iniSection) rmItem(key string, flag bool) bool {
	if s == nil {
		return false
	}
	key = ident(flag, key)
	for i, item := range s.Data {
		if item == nil {
			continue
		}
		if ident(flag, item.Key) == key {
			n := len(s.Data) - 1
			skip := 1
			// Remove newline if the key was the last one in the block.
			if (i > 0 && s.Data[i-1] == nil && i < n && s.Data[i+1] == nil) ||
				(i == 0 && n > 0 && s.Data[1] == nil) {
				// Removed key is not the first key and previous and next keys are newlines.
				// Removed key is the first key and next key is a newline.
				skip++
				n--
			}
			copy(s.Data[i:], s.Data[i+skip:])
			s.Data[n] = nil
			s.Data = s.Data[:n]
			return true
		}
	}
	return false
}

// iniItem represents a key/value pair.
// It may have comments.
type iniItem struct {
	Comments []string
	Key      string
	Value    string
}
