package clients

import (
	"context"
	pb "github.com/aibotsoft/gen/fortedpb"
	"github.com/aibotsoft/micro/config"
	"github.com/aibotsoft/micro/config_client"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
	"strconv"
)

type Clients map[string]pb.FortedClient

func NewClients(cfg *config.Config, log *zap.SugaredLogger, conf *config_client.ConfClient) Clients {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Service.GrpcTimeout)
	defer cancel()
	services, err := conf.GetServices(ctx)
	if err != nil {
		log.Panic(err)
	}
	clients := make(Clients)
	for _, s := range services {
		addr := net.JoinHostPort("", strconv.FormatInt(s.GrpcPort, 10))
		conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Errorw("dial to client error", "name", s.ServiceName, "addr", addr)
			continue
		}
		c := pb.NewFortedClient(conn)
		log.Infow("begin ping to server", "name", s.ServiceName, "addr", addr)
		_, err = c.Ping(ctx, &pb.PingRequest{})
		if err != nil {
			log.Errorw("server do not response to ping", "name", s.ServiceName, "addr", addr)
			continue
		}
		log.Infow("server responded", "name", s.ServiceName, "addr", addr)
		clients[s.FortedName] = c
	}
	return clients
}
