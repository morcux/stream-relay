package sfu

import (
	"sync"

	"github.com/pion/rtp"
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

var GlobalRoom = NewRoom()

func (r *Room) AddPeer(peer *Peer) {
	r.peersLock.Lock()
	defer r.peersLock.Unlock()

	r.peers[peer.ID] = peer
}

func (r *Room) Broadcast(pkt *rtp.Packet, senderID string) {
	r.peersLock.RLock()
	defer r.peersLock.RUnlock()

	for id, peer := range r.peers {
		if id == senderID {
			continue
		}

		if peer.TrackLocal != nil {
			if err := peer.TrackLocal.WriteRTP(pkt); err != nil {
			}
		}
	}
}

func (r *Room) RemovePeer(peerID string) {
	r.peersLock.Lock()
	defer r.peersLock.Unlock()
	delete(r.peers, peerID)
}
