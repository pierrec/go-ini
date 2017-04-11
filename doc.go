/*
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
*/
package ini
