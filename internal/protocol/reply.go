package protocol

import (
	"fmt"
	"io"
	"strings"
)

// Reply represents a parsed sentence returned by the RouterOS API.
type Reply struct {
	Type string            // "!re", "!done", "!trap", "!fatal"
	Map  map[string]string // key-value attributes (e.g. from "=name=ether1")
	Tag  string            // value from the ".tag=" attribute
}

// ParseReply parses a RouterOS sentence into a Reply object.
func ParseReply(sentence []string) (*Reply, error) {
	if len(sentence) == 0 {
		return nil, fmt.Errorf("empty sentence received")
	}

	rep := &Reply{
		Type: sentence[0],
	}

	// Validate reply type
	switch rep.Type {
	case "!re", "!done", "!trap", "!fatal":
		// valid types
	default:
		return nil, fmt.Errorf("invalid reply type: %q", rep.Type)
	}

	for _, word := range sentence[1:] {
		if strings.HasPrefix(word, ".tag=") {
			rep.Tag = word[5:]
		} else if strings.HasPrefix(word, "=") {
			if rep.Map == nil {
				rep.Map = make(map[string]string)
			}
			// Format is =key=value (or =key=)
			// Find the index of the second '=' which separates key and value
			valIdx := strings.Index(word[1:], "=")
			if valIdx >= 0 {
				key := word[1 : valIdx+1]
				val := word[valIdx+2:]
				rep.Map[key] = val
			} else {
				// No second '=' found, treat the whole thing as a key with empty value
				rep.Map[word[1:]] = ""
			}
		}
	}

	return rep, nil
}

// ReadReply reads a RouterOS reply sentence directly from the reader and parses it on the fly.
// This completely avoids allocating temporary string slices for sentences.
func ReadReply(r io.Reader) (*Reply, error) {
	typeWord, err := ReadWord(r)
	if err != nil {
		return nil, err
	}
	if typeWord == "" {
		return nil, fmt.Errorf("unexpected empty reply sentence")
	}

	rep := &Reply{
		Type: typeWord,
	}

	// Validate reply type
	switch rep.Type {
	case "!re", "!done", "!trap", "!fatal":
		// valid types
	default:
		return nil, fmt.Errorf("invalid reply type: %q", rep.Type)
	}

	for {
		word, err := ReadWord(r)
		if err != nil {
			return nil, err
		}
		if word == "" {
			break
		}

		if strings.HasPrefix(word, ".tag=") {
			rep.Tag = word[5:]
		} else if strings.HasPrefix(word, "=") {
			if rep.Map == nil {
				rep.Map = make(map[string]string)
			}
			valIdx := strings.Index(word[1:], "=")
			if valIdx >= 0 {
				key := word[1 : valIdx+1]
				val := word[valIdx+2:]
				rep.Map[key] = val
			} else {
				rep.Map[word[1:]] = ""
			}
		}
	}

	return rep, nil
}

// AsError converts a Trap or Fatal reply to a Go error if applicable.
// Returns nil if the reply is not an error type.
func (r *Reply) AsError() error {
	switch r.Type {
	case "!trap":
		return &TrapError{
			Category: r.Map["category"],
			Message:  r.Map["message"],
		}
	case "!fatal":
		return &FatalError{
			Message: r.Map["message"],
		}
	}
	return nil
}
