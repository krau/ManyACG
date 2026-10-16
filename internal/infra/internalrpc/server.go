package internalrpc

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/internalrpc/internalpb"
	"github.com/krau/ManyACG/pkg/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const stopGrace = 5 * time.Second

var errServerStopped = errors.New("internal RPC server stopped")

type Server struct {
	addr string
	grpc *grpc.Server
	hub  *Hub
}

func NewServer(cfg runtimecfg.InternalConfig, hub *Hub) (*Server, error) {
	_, err := readAuthorizedKeys(cfg.Gateway.AuthorizedKeys)
	if err != nil {
		return nil, err
	}
	gateway := NewGateway(hub, cfg.Gateway.AuthorizedKeys)
	options := []grpc.ServerOption{
		grpc.StatsHandler(connectionAuthHandler{}),
		grpc.MaxRecvMsgSize(maxMessageSize),
		grpc.MaxSendMsgSize(maxMessageSize),
		grpc.ChainUnaryInterceptor(gateway.unaryAuth),
		grpc.ChainStreamInterceptor(gateway.streamAuth),
		grpc.KeepaliveParams(keepalive.ServerParameters{Time: 60 * time.Second, Timeout: 20 * time.Second}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{MinTime: 15 * time.Second, PermitWithoutStream: true}),
	}
	if cfg.Gateway.SSH {
		transport, err := newSSHServerCredentials(cfg.Gateway.HostKey, cfg.Gateway.AuthorizedKeys)
		if err != nil {
			return nil, err
		}
		options = append(options, grpc.Creds(transport))
	}
	server := grpc.NewServer(options...)
	internalpb.RegisterGatewayServer(server, gateway)
	return &Server{addr: cfg.Gateway.Addr, grpc: server, hub: hub}, nil
}

func (s *Server) Run(ctx context.Context) error {
	listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", s.addr)
	if err != nil {
		return err
	}
	defer listener.Close()
	stopped := make(chan struct{})
	stop := context.AfterFunc(ctx, func() {
		s.stopGracefully()
		close(stopped)
	})
	defer func() {
		if !stop() {
			<-stopped
		}
		s.grpc.Stop()
	}()
	log.Info("Starting internal RPC server", "addr", listener.Addr().String())
	err = s.grpc.Serve(listener)
	if errors.Is(err, grpc.ErrServerStopped) {
		return nil
	}
	return err
}

func (s *Server) Close() error {
	s.grpc.Stop()
	return nil
}

func (s *Server) stopGracefully() {
	s.hub.closeAll(errServerStopped)
	stopped := make(chan struct{})
	go func() {
		s.grpc.GracefulStop()
		close(stopped)
	}()
	timer := time.NewTimer(stopGrace)
	defer timer.Stop()
	select {
	case <-stopped:
	case <-timer.C:
		s.grpc.Stop()
		<-stopped
	}
}
