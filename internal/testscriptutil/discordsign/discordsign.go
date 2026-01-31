// Package discordsign provides funcionality for generating request signature headers in testscript tests to simulate
// Discord interaction webhook requests.
// See: https://discord.com/developers/docs/interactions/overview#setting-up-an-endpoint-validating-security-request-headers
package discordsign

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/rogpeppe/go-internal/testscript"
)

type privateKey struct{}

// Init creates an Ed25519 key pair for use in request signatures.
// It must be called in the testscript Setup function.
//
// The hex-encoded public key is stored in an env variable named PUBLIC_KEY.
func Init(env *testscript.Env) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		env.T().Fatal("error generating Ed25519 key pair: %v", err)
	}

	env.Setenv("PUBLIC_KEY", hex.EncodeToString(pub))

	env.Values[privateKey{}] = priv
}

// Cmd returns a [github.com/rogpeppe/go-internal/testscript] command that signs the contents of the provided file,
// i.e. request body. Signature and timestamp request headers are written to the specified file.
func Cmd(name string) func(*testscript.TestScript, bool, []string) {
	return func(ts *testscript.TestScript, _ bool, args []string) {
		const (
			inPath int = iota
			outPath
		)

		if len(args) != 2 {
			ts.Fatalf("usage: %s inPath outPath", name)
		}

		body := ts.ReadFile(args[inPath])
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		msg := append([]byte(timestamp), []byte(body)...)

		priv, ok := ts.Value(privateKey{}).(ed25519.PrivateKey)
		if !ok {
			ts.Fatalf("No private key in environment. Init() must be called in testscript Setup.")
		}

		sig := ed25519.Sign(priv, msg)
		headers := fmt.Sprintf(
			"X-Signature-Ed25519: %s\nX-Signature-Timestamp: %s\n",
			hex.EncodeToString(sig),
			timestamp,
		)
		headersFile := ts.MkAbs(args[1])

		err := os.WriteFile(headersFile, []byte(headers), 0600)
		if err != nil {
			ts.Fatalf("Error writing headers file: %v", err)
		}
	}
}
