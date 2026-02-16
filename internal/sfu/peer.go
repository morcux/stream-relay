package sfu

import (
	"fmt"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"
)

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
