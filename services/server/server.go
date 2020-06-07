package server

import (
	"context"
	pb "github.com/aibotsoft/gen/fortedpb"
	"github.com/aibotsoft/micro/config"
	"github.com/aibotsoft/middle-service/services/handler"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
)

type Server struct {
	cfg     *config.Config
	log     *zap.SugaredLogger
	gs      *grpc.Server
	handler *handler.Handler
	pb.UnimplementedFortedServer
}

func New(cfg *config.Config, log *zap.SugaredLogger, handler *handler.Handler) *Server {
	return &Server{cfg: cfg, log: log, handler: handler, gs: grpc.NewServer()}
}
func (s *Server) Close() {
	s.log.Debug("begin gRPC server gracefulStop")
	s.gs.GracefulStop()
	s.handler.Close()
	s.log.Debug("end gRPC server gracefulStop")
}
func (s *Server) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	return &pb.PingResponse{}, nil
}
func (s *Server) Serve() error {
	addr, err := s.handler.Conf.GetGrpcAddr(context.Background(), s.cfg.Service.Name)
	if err != nil {
		return err
	}
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return errors.Wrap(err, "net.Listen error")
	}
	pb.RegisterFortedServer(s.gs, s)
	s.log.Infow("gRPC server listens", "name", s.cfg.Service.Name, "addr", addr)
	return s.gs.Serve(lis)
}
