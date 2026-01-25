package server

import (
	"fmt"
	"net/http"

	"github.com/iocalebs/patrolbot/internal/config"
	"github.com/iocalebs/patrolbot/internal/server/routes/interactions"
)

func mux(cfg config.Config) (http.Handler, error) {
	mux := http.NewServeMux()

	interactionsHandler, err := interactions.NewHandler(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize /interactions handler: %w", err)
	}

	mux.Handle("/interactions", interactionsHandler)

	return mux, nil
}
