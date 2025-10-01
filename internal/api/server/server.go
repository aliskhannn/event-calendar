package server

import "net/http"

// New creates and configures a new HTTP server instance.
// It initializes the server with the specified address and handler.
//
// Parameters:
//   - addr: The address the server will listen on (e.g., ":8080").
//   - handler: The HTTP handler to process incoming requests.
//
// Returns:
//   - A pointer to the configured http.Server instance.
func New(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:    addr,    // server listening address
		Handler: handler, // handler for processing HTTP requests
	}
}
