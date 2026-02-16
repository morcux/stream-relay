package main

import (
	"stream-relay/internal/signal"
)

func main() {
	server := signal.NewHTTPServer(":8080")
	server.ListenAndServe()
}
