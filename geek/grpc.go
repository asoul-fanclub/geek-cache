package geek

import (
	"context"
	pb "geek-cache/geek/pb"

	"google.golang.org/grpc"
)

type grpcGetter struct {
	addr string
}

func (g *grpcGetter) Get(in *pb.Request, out *pb.Response) error {
	c, err := grpc.Dial(g.addr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	client := pb.NewGroupCacheClient(c)
	resp, err := client.Get(context.Background(), in)
	out.Value = resp.Value
	return err
}

// var _ PeerGetter = (*grpcGetter)(nil)
