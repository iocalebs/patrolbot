package server

import (
	"net/http"

	"github.com/iocalebs/patrolbot/internal/server/routes/interactions"
)

func newMux() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/interactions", interactions.NewHandler())

	return mux
}
