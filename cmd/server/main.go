package main

import (
	"stream-relay/internal/sfu"
	"stream-relay/internal/signal"
)

func main() {
	room := sfu.NewRoom()
	server := signal.NewHTTPServer(":8080", room)
	server.ListenAndServe()
}
