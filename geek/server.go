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

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	defaultServiceName = "geek-cache"
	defaultAddr        = "127.0.0.1:7654"
)

type Server struct {
	pb.UnimplementedGroupCacheServer
	self       string     // self ip
	sname      string     // name of service
	status     bool       // true if the server is running
	mu         sync.Mutex // guards
	stopSignal chan error // signal to stop
}

type ServerOptions func(*Server)

func NewServer(self string, opts ...ServerOptions) (*Server, error) {
	if self == "" {
		self = defaultAddr
	} else if !validPeerAddr(self) {
		return nil, fmt.Errorf("invalid address: %v", self)
	}
	s := Server{
		self:  self,
		sname: defaultServiceName,
	}
	for _, opt := range opts {
		opt(&s)
	}
	return &s, nil
}

func (s *Server) ServiceName(sname string) ServerOptions {
	return func(s *Server) {
		s.sname = sname
	}
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
		err := registy.Register(s.sname, s.self, s.stopSignal)
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

func (s *Server) Stop() {
	s.mu.Lock()
	if !s.status {
		s.mu.Unlock()
		return
	}
	s.stopSignal <- nil
	s.status = false
	s.mu.Unlock()
}
