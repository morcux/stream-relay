package sfu

import (
	"fmt"
	"sync"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
)

type Room struct {
	peersLock sync.RWMutex
	peers     map[string]*Peer
}

func NewRoom() *Room {
	return &Room{
		peers: make(map[string]*Peer),
	}
}

func (r *Room) AddPeer(peer *Peer) {
	r.peersLock.Lock()
	defer r.peersLock.Unlock()

	r.peers[peer.ID] = peer

	for existingPeerID, existingPeer := range r.peers {
		if existingPeerID == peer.ID {
			continue
		}

		existingPeer.Mu.RLock()
		track := existingPeer.RemoteTrack
		existingPeer.Mu.RUnlock()

		if track != nil {
			fmt.Printf("Subscribing new peer %s to existing peer %s\n", peer.ID, existingPeerID)

			if err := peer.Subscribe(existingPeerID, track); err != nil {
				fmt.Printf("Failed to subscribe new peer to existing: %v\n", err)
			}
		}
	}
}
	
func (r *Room) Broadcast(pkt *rtp.Packet, senderID string) {
	r.peersLock.RLock()
	defer r.peersLock.RUnlock()

	for id, peer := range r.peers {
		if id == senderID {
			continue
		}

		peer.Mu.RLock()
		track, ok := peer.Senders[senderID]
		peer.Mu.RUnlock()

		if ok {
			clonedPkt := *pkt
			clonedPkt.Header = pkt.Header.Clone()

			if err := track.WriteRTP(&clonedPkt); err != nil {
			}
		}
	}
}
func (r *Room) RemovePeer(peerID string) {
	r.peersLock.Lock()
	defer r.peersLock.Unlock()
	delete(r.peers, peerID)
}

func (r *Room) SubscribeTo(publisherID string, track *webrtc.TrackRemote) {
	r.peersLock.RLock()
	defer r.peersLock.RUnlock()

	for peerID, peer := range r.peers {
		if peerID == publisherID {
			continue
		}

		if err := peer.Subscribe(publisherID, track); err != nil {
			fmt.Printf("Failed to subscribe peer %s to %s: %v\n", peerID, publisherID, err)
		}
	}
}
