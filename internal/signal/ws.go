package signal

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"

	"stream-relay/internal/sfu"
)

type Message struct {
	Type string `json:"type"`
	Data string `json:"data,omitempty"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request, room *sfu.Room) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	peerID := fmt.Sprintf("peer_%d", time.Now().UnixNano())
	peer := &sfu.Peer{
		ID:     peerID,
		Socket: c,
	}

	if err := sfu.InitPeer(peer, room); err != nil {
		log.Println("Peer init failed:", err)
		return
	}

	room.AddPeer(peer)
	defer room.RemovePeer(peer.ID)
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "offer":
			offer := webrtc.SessionDescription{
				Type: webrtc.SDPTypeOffer,
				SDP:  msg.Data,
			}

			if err := peer.PC.SetRemoteDescription(offer); err != nil {
				log.Println("SetRemoteDescription error:", err)
				continue
			}
			answer, err := peer.PC.CreateAnswer(nil)
			if err != nil {
				log.Println("CreateAnswer error:", err)
				continue
			}

			if err := peer.PC.SetLocalDescription(answer); err != nil {
				continue
			}

			resp := Message{
				Type: "answer",
				Data: peer.PC.LocalDescription().SDP,
			}
			if err := peer.SendJSON(resp); err != nil {
				log.Println("Write Answer error:", err)
			}
		case "answer":
			fmt.Println("Received Answer from Client")
			answer := webrtc.SessionDescription{
				Type: webrtc.SDPTypeAnswer,
				SDP:  msg.Data,
			}
			if err := peer.PC.SetRemoteDescription(answer); err != nil {
				log.Println("SetRemoteDescription (Answer) error:", err)
			}
		}
	}
}
