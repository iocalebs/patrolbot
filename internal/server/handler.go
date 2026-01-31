package server

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/iocalebs/patrolbot/internal/config"
	"github.com/iocalebs/patrolbot/internal/discord"
	"github.com/iocalebs/patrolbot/internal/errdefs"
	"github.com/iocalebs/patrolbot/internal/interactions"
)

func handler(cfg config.Config, logger *slog.Logger) (http.Handler, error) {
	mux := http.NewServeMux()

	interactionsHandler, err := interactionsHandler(cfg, logger)
	if err != nil {
		return nil, err
	}

	mux.Handle("/interactions", interactionsHandler)

	return mux, nil
}

func interactionsHandler(cfg config.Config, logger *slog.Logger) (http.Handler, error) {
	key, err := decodePublicKey(cfg)
	if err != nil {
		return nil, err
	}

	interactionHandler, err := interactions.NewHandler(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize interaction handler: %w", err)
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
			logger.ErrorContext(r.Context(), "Failed to decode interaction request", "err", err)
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		w.Header().Set("Content-Type", "application/json")

		response := interactionHandler.Respond(r.Context(), interaction)

		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			logger.ErrorContext(r.Context(), "Failed to encode interaction response", "err", err)
			w.WriteHeader(http.StatusInternalServerError)

			return
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
