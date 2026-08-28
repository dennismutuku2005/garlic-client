package protocol

import (
	"fmt"
	"io"
)

var zeroByte = []byte{0}

// WriteSentence writes a series of words followed by a terminating zero-length word.
func WriteSentence(w io.Writer, words []string) error {
	for _, word := range words {
		if err := WriteWord(w, word); err != nil {
			return err
		}
	}
	// Terminating empty word (0x00 byte)
	if _, err := w.Write(zeroByte); err != nil {
		return fmt.Errorf("failed to write sentence terminator: %w", err)
	}
	return nil
}

// WriteCommand writes a Command directly to the writer, avoiding slice allocations.
func WriteCommand(w io.Writer, cmd *Command) error {
	if err := WriteWord(w, cmd.Name); err != nil {
		return err
	}
	for _, arg := range cmd.Args {
		if err := WriteWord(w, arg); err != nil {
			return err
		}
	}
	if cmd.Tag != "" {
		if err := WriteWord(w, ".tag="+cmd.Tag); err != nil {
			return err
		}
	}
	// Terminating empty word (0x00 byte)
	if _, err := w.Write(zeroByte); err != nil {
		return fmt.Errorf("failed to write command sentence terminator: %w", err)
	}
	return nil
}

// ReadSentence reads words until a terminating zero-length word is encountered.
func ReadSentence(r io.Reader) ([]string, error) {
	words := make([]string, 0, 8)
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
