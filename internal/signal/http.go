package signal

import (
	"net/http"
	"stream-relay/internal/sfu"
)

func NewHTTPServer(addr string, room *sfu.Room) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		HandleWebSocket(w, r, room)
	})

	fs := http.FileServer(http.Dir("web/public"))
	mux.Handle("/", fs)

	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}
