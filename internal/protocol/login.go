package protocol

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
)

// Login executes the RouterOS API login handshake.
// It auto-detects between the modern plain-text login (v7, v6.43+) and legacy MD5 challenge-response.
func Login(rw io.ReadWriter, username, password string) error {
	// 1. Attempt modern login (sending username and password directly)
	loginCmd := NewCommand("/login").Arg("name", username).Arg("password", password)
	if err := WriteSentence(rw, loginCmd.Sentence()); err != nil {
		return fmt.Errorf("failed to send initial login command: %w", err)
	}

	replyWords, err := ReadSentence(rw)
	if err != nil {
		return fmt.Errorf("failed to read login response: %w", err)
	}

	reply, err := ParseReply(replyWords)
	if err != nil {
		return fmt.Errorf("failed to parse login response: %w", err)
	}

	if err := reply.AsError(); err != nil {
		return err
	}

	if reply.Type == "!done" {
		challenge, hasChallenge := reply.Map["ret"]
		if hasChallenge && challenge != "" {
			// Legacy MD5 challenge-response is required
			md5Response, err := calculateMD5Response(password, challenge)
			if err != nil {
				return fmt.Errorf("failed to calculate MD5 challenge response: %w", err)
			}

			legacyLoginCmd := NewCommand("/login").Arg("name", username).Arg("response", md5Response)
			if err := WriteSentence(rw, legacyLoginCmd.Sentence()); err != nil {
				return fmt.Errorf("failed to send legacy login response: %w", err)
			}

			legacyReplyWords, err := ReadSentence(rw)
			if err != nil {
				return fmt.Errorf("failed to read legacy login response: %w", err)
			}

			legacyReply, err := ParseReply(legacyReplyWords)
			if err != nil {
				return fmt.Errorf("failed to parse legacy login response: %w", err)
			}

			if err := legacyReply.AsError(); err != nil {
				return err
			}

			if legacyReply.Type != "!done" {
				return fmt.Errorf("unexpected legacy login reply type: %s", legacyReply.Type)
			}
		}
	} else {
		return fmt.Errorf("unexpected initial login reply type: %s", reply.Type)
	}

	return nil
}

// calculateMD5Response computes: "00" + MD5(0x00 + password + challenge_bytes)
func calculateMD5Response(password, challengeHex string) (string, error) {
	challengeBytes, err := hex.DecodeString(challengeHex)
	if err != nil {
		return "", fmt.Errorf("invalid challenge hex: %w", err)
	}

	hasher := md5.New()
	hasher.Write([]byte{0x00})
	hasher.Write([]byte(password))
	hasher.Write(challengeBytes)

	return "00" + hex.EncodeToString(hasher.Sum(nil)), nil
}
