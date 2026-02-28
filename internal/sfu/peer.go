package sfu

import (
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"
)

type Peer struct {
	ID          string
	PC          *webrtc.PeerConnection
	Socket      *websocket.Conn
	SocketMu    sync.Mutex
	Mu          sync.RWMutex
	Senders     map[string]*webrtc.TrackLocalStaticRTP
	RemoteTrack *webrtc.TrackRemote
}

func (p *Peer) SendJSON(v interface{}) error {
	p.SocketMu.Lock()
	defer p.SocketMu.Unlock()
	return p.Socket.WriteJSON(v)
}

func InitPeer(p *Peer, room *Room) error {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	}

	p.Senders = make(map[string]*webrtc.TrackLocalStaticRTP)

	pc, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return err
	}
	p.PC = pc

	p.PC.OnTrack(func(remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		p.OnTrack(remoteTrack, room)
	})

	return nil
}
func (p *Peer) OnTrack(remoteTrack *webrtc.TrackRemote, room *Room) {
	fmt.Printf("Track received from %s: Codec: %s\n", p.ID, remoteTrack.Codec().MimeType)
	p.Mu.Lock()
	p.RemoteTrack = remoteTrack
	p.Mu.Unlock()
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			p.PC.WriteRTCP([]rtcp.Packet{
				&rtcp.PictureLossIndication{MediaSSRC: uint32(remoteTrack.SSRC())},
			})
		}
	}()

	room.SubscribeTo(p.ID, remoteTrack)

	for {
		pkt, _, err := remoteTrack.ReadRTP()
		if err != nil {
			return
		}

		pkt.Header.PayloadType = uint8(remoteTrack.PayloadType())

		room.Broadcast(pkt, p.ID)
	}
}
func (p *Peer) Subscribe(remotePeerID string, remoteTrack *webrtc.TrackRemote) error {
	p.Mu.Lock()
	defer p.Mu.Unlock()

	outputTrack, err := webrtc.NewTrackLocalStaticRTP(
		remoteTrack.Codec().RTPCodecCapability,
		"video_"+remotePeerID,
		"pion_stream_"+remotePeerID,
	)
	if err != nil {
		return err
	}

	rtpSender, err := p.PC.AddTrack(outputTrack)
	if err != nil {
		return err
	}

	p.Senders[remotePeerID] = outputTrack

	go func() {
		buf := make([]byte, 1500)
		for {
			if _, _, err := rtpSender.Read(buf); err != nil {
				return
			}
		}
	}()

	go p.Negotiate()

	return nil
}

func (p *Peer) Negotiate() {
	p.Mu.Lock()
	defer p.Mu.Unlock()

	offer, err := p.PC.CreateOffer(nil)
	if err != nil {
		fmt.Println("Negotiate CreateOffer error:", err)
		return
	}

	if err := p.PC.SetLocalDescription(offer); err != nil {
		fmt.Println("Negotiate SetLocalDescription error:", err)
		return
	}

	gatherComplete := webrtc.GatheringCompletePromise(p.PC)
	<-gatherComplete

	offerMsg := map[string]string{
		"type": "offer",
		"data": p.PC.LocalDescription().SDP,
	}

	if err := p.SendJSON(offerMsg); err != nil {
		fmt.Println("Negotiate WriteJSON error:", err)
	}
}
