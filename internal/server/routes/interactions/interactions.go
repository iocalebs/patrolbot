// Package interactions implements the /interactions route handler
package interactions

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/iocalebs/patrolbot/internal/config"
	"github.com/iocalebs/patrolbot/internal/discord"
	"github.com/iocalebs/patrolbot/internal/errdefs"
)

// NewHandler returns the [http.Handler] for the /interactions route.
func NewHandler(cfg config.Config) (http.Handler, error) {
	key, err := decodePublicKey(cfg)
	if err != nil {
		return nil, err
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		valid := discord.VerifySignature(r, key)
		if !valid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		var interaction discord.Interaction

		err := json.NewDecoder(r.Body).Decode(&interaction)
		if err != nil {
			http.Error(w, "Failed to decode interaction", http.StatusBadRequest) // TODO: How does this appear in Discord?
			return
		}

		switch interaction.Type {
		case discord.InteractionTypePing:
			pong := discord.InteractionResponse{
				Type: discord.InteractionCallbackTypePong,
			}

			w.Header().Set("Content-Type", "application/json")

			err = json.NewEncoder(w).Encode(pong)
			if err != nil {
				http.Error(w, "Failed to encode interaction response", http.StatusInternalServerError)
				return
			}
		}
	}), nil
}

func decodePublicKey(cfg config.Config) ([]byte, error) {
	if cfg.Discord.PublicKey == "" {
		return nil, errdefs.NewConfigError(".discord.publicKey not set")
	}

	key, err := hex.DecodeString(cfg.Discord.PublicKey)
	if err != nil {
		msg := fmt.Sprintf("failed to decode .discord.publicKey: %v", err)
		return nil, errdefs.NewConfigError(msg)
	}

	return key, nil
}
