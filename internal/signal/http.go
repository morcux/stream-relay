package signal

import (
	"encoding/json"
	"net/http"
	"stream-relay/internal/sfu"

	"github.com/pion/webrtc/v4"
)

func NewHTTPServer(addr string) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/sdp", handleSDP)

	fs := http.FileServer(http.Dir("web/public"))

	mux.Handle("/", fs)

	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}

func handleSDP(w http.ResponseWriter, r *http.Request) {
	var offer webrtc.SessionDescription
	if err := json.NewDecoder(r.Body).Decode(&offer); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	answer, err := sfu.SetupPeer(offer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(answer)
}
