package protocol

import (
	"fmt"
	"io"
)

// WriteSentence writes a series of words followed by a terminating zero-length word.
func WriteSentence(w io.Writer, words []string) error {
	for _, word := range words {
		if err := WriteWord(w, word); err != nil {
			return err
		}
	}
	// Terminating empty word (0x00 byte)
	if _, err := w.Write([]byte{0}); err != nil {
		return fmt.Errorf("failed to write sentence terminator: %w", err)
	}
	return nil
}

// ReadSentence reads words until a terminating zero-length word is encountered.
func ReadSentence(r io.Reader) ([]string, error) {
	var words []string
	for {
		word, err := ReadWord(r)
		if err != nil {
			return nil, err
		}
		if word == "" {
			break
		}
		words = append(words, word)
	}
	return words, nil
}
