package domain

import (
	"maps"
	"slices"
	"strconv"
	"strings"
)

var jsonPointerEscaper = strings.NewReplacer("~", "~0", "/", "~1")

// ValidateStorageText checks a decoded string before it is stored in PostgreSQL.
func ValidateStorageText(value, pointer string) error {
	if strings.ContainsRune(value, 0) {
		return Fail(422, "validation_failed", pointer, "NUL characters cannot be stored in this field")
	}
	return nil
}

// ValidateStorageJSONText checks strings and keys in a decoded JSON tree going
// into JSONB. Visit keys in order so a document with multiple violations gets
// a consistent first error.
func ValidateStorageJSONText(value any, pointer string) error {
	switch v := value.(type) {
	case string:
		return ValidateStorageText(v, pointer)
	case map[string]any:
		for _, key := range slices.Sorted(maps.Keys(v)) {
			// Use the containing object as the pointer when the key itself
			// is invalid so the response does not echo the invalid byte.
			if err := ValidateStorageText(key, pointer); err != nil {
				return err
			}
			if err := ValidateStorageJSONText(v[key], pointer+"/"+jsonPointerEscaper.Replace(key)); err != nil {
				return err
			}
		}
	case []any:
		for i, item := range v {
			if err := ValidateStorageJSONText(item, pointer+"/"+strconv.Itoa(i)); err != nil {
				return err
			}
		}
	}
	return nil
}
