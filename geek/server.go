package geek

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"

	pb "github.com/Makonike/geek-cache/geek/pb"
	registy "github.com/Makonike/geek-cache/geek/registry"

	"github.com/Makonike/geek-cache/geek/consistenthash"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	defaultBasePath = "/_geek_cache/"
	defaultAddr     = "127.0.0.1:7654"
	defaultReplicas = 150
)

type Server struct {
	pb.UnimplementedGroupCacheServer
	self       string              // self ip
	status     bool                // true if the server is running
	mu         sync.Mutex          // guards
	consHash   *consistenthash.Map // stores the list of peers, selected by specific key
	clients    map[string]*Client  // keyed by e.g. "10.0.0.2:8009"
	stopSignal chan error          // signal to stop
}

func NewServer(self string) (*Server, error) {
	if self == "" {
		self = defaultAddr
	} else if !validPeerAddr(self) {
		return nil, fmt.Errorf("invalid address: %v", self)
	}
	return &Server{
		self: self,
	}, nil
}

func (s *Server) Self() string {
	return s.self
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
	// 启动 reflection 反射服务
	reflection.Register(grpcServer)
	go func() {
		err := registy.Register("geek-cache", s.self, s.stopSignal)
		if err != nil {
			log.Fatalf(err.Error())
		}
		close(s.stopSignal)
		err = l.Close()
		if err != nil {
			log.Fatalf(err.Error())
		}
		log.Printf("[%s] Revoke service and close tcp socket ok", s.self)
	}()

	s.mu.Unlock()
	if err := grpcServer.Serve(l); s.status && err != nil {
		return fmt.Errorf("failed to serve on %s: %v", port, err)
	}
	return nil
}

// add peer to cluster, create a new Client instance for every peer
func (s *Server) Set(peers ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.consHash = consistenthash.New(defaultReplicas, nil)
	s.consHash.Add(peers...)
	s.clients = make(map[string]*Client, len(peers))
	for _, peer := range peers {
		s.clients[peer], _ = NewClient(peer)
	}
}

func (s *Server) SetWithReplicas(replicas int, peers ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.consHash = consistenthash.New(replicas, nil)
	s.consHash.Add(peers...)
	s.clients = make(map[string]*Client, len(peers))
	for _, peer := range peers {
		s.clients[peer], _ = NewClient(peer)
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
	s.consHash = nil
	s.mu.Unlock()
}

// PickPeer pick a peer with the consistenthash algorithm
func (s *Server) PickPeer(key string) (PeerGetter, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if peer := s.consHash.Get(key); peer != "" && peer != s.self {
		s.Log("Pick peer %s", peer)
		return s.clients[peer], true
	}
	return nil, false
}

// resure implemented
var _ PeerPicker = (*Server)(nil)
