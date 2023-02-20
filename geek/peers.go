package geek

import (
	"fmt"
	"log"
	"sync"

	"github.com/Makonike/geek-cache/geek/consistenthash"
)

// PeerPicker must be implemented to locate the peer that owns a specific key
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter must be implemented by a peer
type PeerGetter interface {
	Get(group string, key string) ([]byte, error)
}

type ClientPicker struct {
	self        string // self ip
	serviceName string
	mu          sync.Mutex          // guards
	consHash    *consistenthash.Map // stores the list of peers, selected by specific key
	clients     map[string]*Client  // keyed by e.g. "10.0.0.2:8009"
}

func NewClientPicker(self string, opts ...PickerOptions) *ClientPicker {
	picker := ClientPicker{
		self:        self,
		serviceName: defaultServiceName,
	}
	for _, opt := range opts {
		opt(&picker)
	}
	return &picker
}

type PickerOptions func(*ClientPicker)

func PickerServiceName(serviceName string) PickerOptions {
	return func(picker *ClientPicker) {
		picker.serviceName = serviceName
	}
}

// add peer to cluster, create a new Client instance for every peer
func (s *ClientPicker) Set(hash consistenthash.Hash, peers ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.consHash = consistenthash.New(consistenthash.HashFunc(hash))
	s.consHash.Add(peers...)
	s.clients = make(map[string]*Client, len(peers))
	for _, peer := range peers {
		s.clients[peer], _ = NewClient(peer, s.serviceName)
	}
}

func (s *ClientPicker) SetSimply(peers ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.consHash = consistenthash.New()
	s.consHash.Add(peers...)
	s.clients = make(map[string]*Client, len(peers))
	for _, peer := range peers {
		s.clients[peer], _ = NewClient(peer, s.serviceName)
	}
}

// PickPeer pick a peer with the consistenthash algorithm
func (s *ClientPicker) PickPeer(key string) (PeerGetter, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if peer := s.consHash.Get(key); peer != "" && peer != s.self {
		s.Log("Pick peer %s", peer)
		return s.clients[peer], true
	}
	return nil, false
}

func (s *ClientPicker) SetWithReplicas(hash consistenthash.Hash, replicas int, peers ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.consHash == nil {
		s.consHash = consistenthash.New(consistenthash.Replicas(replicas), consistenthash.HashFunc(hash))
	}
	s.consHash.Add(peers...)
	if s.clients == nil {
		s.clients = make(map[string]*Client, len(peers))
	}
	for _, peer := range peers {
		s.clients[peer], _ = NewClient(peer, s.serviceName)
	}
}

func (s *ClientPicker) SetSinglePeer(hash consistenthash.Hash, replicas int, peer string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.consHash == nil {
		s.consHash = consistenthash.New(consistenthash.Replicas(replicas), consistenthash.HashFunc(hash))
	}
	s.consHash.Add(peer)
	if s.clients == nil {
		s.clients = make(map[string]*Client, 1)
	}
	s.clients[peer], _ = NewClient(peer, s.serviceName)
}

// Log info
func (s *ClientPicker) Log(format string, path ...interface{}) {
	log.Printf("[Server %s] %s", s.self, fmt.Sprintf(format, path...))
}
