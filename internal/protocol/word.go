package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
)

// WriteWord encodes a string word in RouterOS length-prefixed format using stack-allocated buffers.
func WriteWord(w io.Writer, word string) error {
	data := []byte(word)
	length := len(data)

	var lenBytes []byte
	var localBuf [5]byte

	if length < 0x80 {
		localBuf[0] = byte(length)
		lenBytes = localBuf[:1]
	} else if length < 0x4000 {
		binary.BigEndian.PutUint16(localBuf[:2], uint16(length|0x8000))
		lenBytes = localBuf[:2]
	} else if length < 0x200000 {
		binary.BigEndian.PutUint32(localBuf[:4], uint32(length|0xC00000))
		lenBytes = localBuf[1:4] // Extract the 3 lower bytes
	} else if length < 0x10000000 {
		binary.BigEndian.PutUint32(localBuf[:4], uint32(length|0xE0000000))
		lenBytes = localBuf[:4]
	} else {
		localBuf[0] = 0xF0
		binary.BigEndian.PutUint32(localBuf[1:5], uint32(length))
		lenBytes = localBuf[:5]
	}

	if _, err := w.Write(lenBytes); err != nil {
		return fmt.Errorf("failed to write word length: %w", err)
	}
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("failed to write word data: %w", err)
	}
	return nil
}

// ReadWord decodes a single length-prefixed word from the reader.
func ReadWord(r io.Reader) (string, error) {
	var b0 [1]byte
	if _, err := io.ReadFull(r, b0[:]); err != nil {
		return "", err
	}

	var length int

	if b0[0]&0x80 == 0 {
		length = int(b0[0])
	} else if b0[0]&0xC0 == 0x80 {
		var b1 [1]byte
		if _, err := io.ReadFull(r, b1[:]); err != nil {
			return "", fmt.Errorf("failed to read 2nd byte of length: %w", err)
		}
		length = int(b0[0]&0x3F)<<8 | int(b1[0])
	} else if b0[0]&0xE0 == 0xC0 {
		var b [2]byte
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return "", fmt.Errorf("failed to read bytes for 3-byte length: %w", err)
		}
		length = int(b0[0]&0x1F)<<16 | int(b[0])<<8 | int(b[1])
	} else if b0[0]&0xF0 == 0xE0 {
		var b [3]byte
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return "", fmt.Errorf("failed to read bytes for 4-byte length: %w", err)
		}
		length = int(b0[0]&0x0F)<<24 | int(b[0])<<16 | int(b[1])<<8 | int(b[2])
	} else if b0[0] == 0xF0 {
		var b [4]byte
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return "", fmt.Errorf("failed to read bytes for 5-byte length: %w", err)
		}
		length = int(binary.BigEndian.Uint32(b[:]))
	} else {
		return "", fmt.Errorf("invalid RouterOS control/length byte: 0x%02x", b0[0])
	}

	if length == 0 {
		return "", nil
	}

	data := make([]byte, length)
	if _, err := io.ReadFull(r, data); err != nil {
		return "", fmt.Errorf("failed to read word content of length %d: %w", length, err)
	}

	return string(data), nil
}
