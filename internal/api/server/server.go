package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/aliskhannn/distributed-grep/internal/api/handlers/process"
)

// Run starts an HTTP server with the provided process handler.
// The server listens on the specified port and exposes the "/process" endpoint.
//
// Parameters:
//   - processHandler: the handler responsible for processing grep requests.
//   - port: the TCP port to listen on.
//
// This function blocks until the server exits or encounters an error.
// On server failure, it prints the error and exits the program.
func Run(processHandler *process.Handler, port int) {
	// Register the /process endpoint.
	http.HandleFunc("/process", processHandler.Process)
	addr := ":" + strconv.Itoa(port)

	// Start the HTTP server and exit on failure.
	_, _ = fmt.Fprintf(os.Stdout, "server listening on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
