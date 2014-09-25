package validate

import (
	"encoding/json"
	"fmt"
)

type Entry struct {
	kind    entryKind
	message string
	line    int
}

func (e Entry) String() string {
	return fmt.Sprintf("line %d: %s: %s", e.line, e.kind, e.message)
}

func (e Entry) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"kind":    e.kind.String(),
		"message": e.message,
		"line":    e.line,
	})
}

type Report struct {
	entries []Entry
}

func (r *Report) Error(line int, message string) {
	r.entries = append(r.entries, Entry{entryError, message, line})
}

func (r *Report) Warning(line int, message string) {
	r.entries = append(r.entries, Entry{entryWarning, message, line})
}

func (r *Report) Entries() []Entry {
	return r.entries
}

type entryKind int

func (k entryKind) String() string {
	switch k {
	case entryError:
		return "error"
	case entryWarning:
		return "warning"
	default:
		panic(fmt.Sprintf("invalid kind %q", k))
	}
}

const (
	entryError entryKind = iota
	entryWarning
)
