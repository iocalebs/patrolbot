package discord

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"io"
	"net/http"
)

// VerifySignature verifies the signature headers Discord sends with webhook requests and returns true if valid.
// https://discord.com/developers/docs/interactions/overview#setting-up-an-endpoint-validating-security-request-headers
func VerifySignature(r *http.Request, publicKey []byte) bool {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return false
	}

	// Reset body so it can be read again
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	sigHex := r.Header.Get("X-Signature-Ed25519")

	timestamp := r.Header.Get("X-Signature-Timestamp")
	if sigHex == "" || timestamp == "" {
		return false
	}

	sig, err := hex.DecodeString(sigHex)
	if err != nil {
		return false
	}

	message := append([]byte(timestamp), body...)

	return ed25519.Verify(publicKey, message, sig)
}
