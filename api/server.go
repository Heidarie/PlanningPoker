package handler

import (
	"net/http"

	"github.com/heidarie/cli_planning_poker/internal/server"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	// Handle requests using the server package
	server.HandleRequest(w, r)
}
