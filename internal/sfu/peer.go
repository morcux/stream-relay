package sfu

import (
	"fmt"
	"io"
	"time"

	"github.com/gorilla/websocket"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"
)

type Peer struct {
	ID         string
	PC         *webrtc.PeerConnection
	Socket     *websocket.Conn
	TrackLocal *webrtc.TrackLocalStaticRTP
}

func InitPeer(p *Peer, room *Room) error {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	}

	pc, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return err
	}
	p.PC = pc

	localTrack, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8},
		"video_"+p.ID, "pion_stream",
	)
	if err != nil {
		return err
	}
	p.TrackLocal = localTrack

	rtpSender, err := p.PC.AddTrack(p.TrackLocal)
	if err != nil {
		return err
	}

	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, err := rtpSender.Read(rtcpBuf); err != nil {
				return
			}
		}
	}()

	p.PC.OnTrack(func(remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		p.OnTrack(remoteTrack, room)
	})

	return nil
}
func SetupPeer(offer webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	}

	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return nil, err
	}

	localTrack, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8},
		"video_echo", "pion_stream",
	)
	if err != nil {
		return nil, err
	}

	rtpSender, err := peerConnection.AddTrack(localTrack)
	if err != nil {
		return nil, err
	}

	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, err := rtpSender.Read(rtcpBuf); err != nil {
				return
			}
		}
	}()

	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		fmt.Printf("Peer Connection State: %s\n", s.String())
	})

	peerConnection.OnTrack(func(remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		fmt.Printf("Track: %s, Codec: %s\n", remoteTrack.ID(), remoteTrack.Codec().MimeType)
		go func() {
			ticker := time.NewTicker(3 * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				if err := peerConnection.WriteRTCP([]rtcp.Packet{
					&rtcp.PictureLossIndication{MediaSSRC: uint32(remoteTrack.SSRC())},
				}); err != nil {
					return
				}
			}
		}()

		for {
			packet, _, err := remoteTrack.ReadRTP()
			if err != nil {
				return
			}
			if err := localTrack.WriteRTP(packet); err != nil {
				return
			}
		}
	})

	if err := peerConnection.SetRemoteDescription(offer); err != nil {
		return nil, err
	}

	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return nil, err
	}

	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	if err := peerConnection.SetLocalDescription(answer); err != nil {
		return nil, err
	}
	<-gatherComplete

	return peerConnection.LocalDescription(), nil
}
func (p *Peer) OnTrack(remoteTrack *webrtc.TrackRemote, room *Room) {
	fmt.Printf("Track received from %s: Codec: %s\n", p.ID, remoteTrack.Codec().MimeType)

	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := p.PC.WriteRTCP([]rtcp.Packet{
				&rtcp.PictureLossIndication{MediaSSRC: uint32(remoteTrack.SSRC())},
			}); err != nil {
				return
			}
		}
	}()

	for {
		pkt, _, err := remoteTrack.ReadRTP()
		if err != nil {
			if err == io.EOF {
				fmt.Printf("Peer %s closed track\n", p.ID)
			} else {
				fmt.Printf("Read error from %s: %v\n", p.ID, err)
			}
			return
		}

		pkt.Header.PayloadType = uint8(remoteTrack.PayloadType())

		room.Broadcast(pkt, p.ID)
	}
}
