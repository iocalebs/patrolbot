// Package discordsign provides funcionality for generating request signature headers in testscript tests to simulate
// Discord interaction webhook requests.
// See: https://discord.com/developers/docs/interactions/overview#setting-up-an-endpoint-validating-security-request-headers
package discordsign

import (
	"crypto/ed25519"
	"encoding/hex"
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
// i.e. request body. The signature and timestamp are stored in env variables under the provided names.
func Cmd(name string) func(*testscript.TestScript, bool, []string) {
	return func(ts *testscript.TestScript, _ bool, args []string) {
		const (
			inPath int = iota
			signatureEnv
			timestampEnv
		)

		if len(args) != 3 {
			ts.Fatalf("usage: %s inPath signatureEnv timestampEnv", name)
		}

		body := ts.ReadFile(args[inPath])

		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		msg := append([]byte(timestamp), []byte(body)...)

		priv, ok := ts.Value(privateKey{}).(ed25519.PrivateKey)
		if !ok {
			ts.Fatalf("No private key in environment. Init() must be called in testscript Setup.")
		}

		sig := ed25519.Sign(priv, msg)
		ts.Setenv(args[signatureEnv], hex.EncodeToString(sig))
		ts.Setenv(args[timestampEnv], timestamp)
	}
}
