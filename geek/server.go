package geek

import (
	"context"
	"fmt"
	"geek-cache/geek/consistenthash"
	pb "geek-cache/geek/pb"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"google.golang.org/grpc"
)

const (
	defaultBasePath = "/_geek_cache/"
	defaultAddr     = "127.0.0.1:7654"
	defaultReplicas = 50
)

type Server struct {
	pb.UnimplementedGroupCacheServer
	self       string                 // self ip
	basePath   string                 // prefix path for communicating
	status     bool                   // true if the server is running
	mu         sync.Mutex             // guards
	peers      *consistenthash.Map    // stores the list of peers, selected by specific key
	clients    map[string]*httpGetter // keyed by e.g. "http://10.0.0.2:8009"
	stopSignal chan error             // signal to stop
}

func NewServer(self string) (*Server, error) {
	if self == "" {
		self = defaultAddr
	} else if !validPeerAddr(self) {
		return nil, fmt.Errorf("invalid address: %v", self)
	}
	return &Server{
		self:     self,
		basePath: defaultBasePath,
	}, nil
}

// Log info
func (s *Server) Log(format string, path ...interface{}) {
	log.Printf("[Server %s] %s", s.self, fmt.Sprintf(format, path...))
}

func (s *Server) Get(ctx context.Context, in *pb.Request) (*pb.Response, error) {
	group, key := in.GetGroup(), in.GetKey()
	out := &pb.Response{}
	log.Printf("[Geek-Cache %s] Recv RPC Request - (%s)/(%s)", s.self, group, key)

	if key == "" {
		return out, fmt.Errorf("key required")
	}
	g := GetGroup(group)
	if g == nil {
		return out, fmt.Errorf("group not found")
	}
	view, err := g.Get(key)
	if err != nil {
		return out, err
	}
	out.Value = view.ByteSLice()
	return out, nil
}

func (s *Server) Start() error {
	s.mu.Lock()
	if s.status {
		s.mu.Unlock()
		return fmt.Errorf("server already running")
	}
	s.status = true
	s.stopSignal = make(chan error)

	port := strings.Split(s.self, ":")[1]
	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %v", port, err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterGroupCacheServer(grpcServer, s)
	s.mu.Unlock()
	if err := grpcServer.Serve(l); s.status && err != nil {
		return fmt.Errorf("failed to serve on %s: %v", port, err)
	}
	return nil
}

// add peer to cluster, create the httpGetter function for every peer
func (s *Server) Set(peers ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.peers = consistenthash.New(defaultReplicas, nil)
	s.peers.Add(peers...)
	s.clients = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		s.clients[peer] = &httpGetter{
			baseURL: peer + s.basePath,
		}
	}
}

func (s *Server) Stop() {
	s.mu.Lock()
	if !s.status {
		s.mu.Unlock()
		return
	}
	s.stopSignal <- nil
	s.status = false
	s.clients = nil
	s.peers = nil
	s.mu.Unlock()
}

// -------------Client---------------

type httpGetter struct {
	baseURL string // the base URL of remote server
}

// resure implemented
var _ PeerGetter = (*httpGetter)(nil)

// Get send the url for getting specific group and key,
// and return the result
func (h *httpGetter) Get(group, key string) ([]byte, error) {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(group),
		url.QueryEscape(key))
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}
	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}
	return bytes, nil
}

// resure implemented
var _ PeerPicker = (*Server)(nil)

// PickPeer pick a peer with the consistenthash algorithm
func (p *Server) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.clients[peer], true
	}
	return nil, false
}
