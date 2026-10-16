package internalrpc

import (
	"context"
	"errors"
	"io"

	"github.com/krau/ManyACG/internal/infra/internalrpc/internalpb"
	"github.com/krau/ManyACG/pkg/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const maxMessageSize = 8 << 20

type Gateway struct {
	internalpb.UnimplementedGatewayServer
	hub                *Hub
	authorizedKeysFile string
}

func NewGateway(hub *Hub, authorizedKeysFile string) *Gateway {
	return &Gateway{hub: hub, authorizedKeysFile: authorizedKeysFile}
}

func (g *Gateway) unaryAuth(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if _, _, err := authenticatedPeer(ctx); err != nil {
		return nil, err
	}
	return handler(ctx, req)
}

func (g *Gateway) streamAuth(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if info.FullMethod == internalpb.Gateway_Attach_FullMethodName {
		return handler(srv, stream)
	}
	if _, _, err := authenticatedPeer(stream.Context()); err != nil {
		return err
	}
	return handler(srv, stream)
}

func (g *Gateway) Attach(stream internalpb.Gateway_AttachServer) error {
	connection, err := beginAuthentication(stream.Context())
	if err != nil {
		return err
	}
	defer connection.clear()
	registration := g.authenticate(stream)
	if registration.err != nil {
		return registration.err
	}
	register := registration.register
	nodeID, connectionID := registration.nodeID, registration.connectionID
	session, err := g.hub.register(nodeID, connectionID, register)
	if err != nil {
		return status.Error(codes.AlreadyExists, err.Error())
	}
	defer g.hub.dropSession(session)
	log.Info("internal agent attached",
		"name", register.GetName(),
		"session", session.id,
		"node", nodeID,
		"concurrency", session.concurrency,
		"fetch", session.fetch,
		"download", session.download,
	)
	defer log.Info("internal agent detached", "name", register.GetName(), "session", session.id)
	connection.authorize(nodeID, connectionID)
	if err := stream.Send(&internalpb.Assignment{
		Payload: &internalpb.Assignment_Accepted{Accepted: &internalpb.SessionAccepted{
			SessionId:   session.id,
			Concurrency: session.concurrency,
		}},
	}); err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()
	errorsCh := make(chan error, 2)
	go func() {
		for {
			select {
			case assignment := <-session.send:
				if err := stream.Send(assignment); err != nil {
					errorsCh <- err
					return
				}
			case <-session.closed:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() {
		for {
			if _, err := stream.Recv(); err != nil {
				errorsCh <- err
				return
			}
		}
	}()
	select {
	case err := <-errorsCh:
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	case <-session.closed:
		// Replaced by a newer connection or the gateway is shutting down.
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (g *Gateway) CompleteJob(ctx context.Context, completion *internalpb.JobCompletion) (*internalpb.Ack, error) {
	_, connectionID, err := authenticatedPeer(ctx)
	if err != nil {
		return nil, err
	}
	if completion.GetJobId() == "" {
		return nil, status.Error(codes.InvalidArgument, "job_id is required")
	}
	if err := g.hub.complete(connectionID, completion); err != nil {
		return nil, jobStatus(err)
	}
	return &internalpb.Ack{Ok: true}, nil
}

func (g *Gateway) PushDownload(stream internalpb.Gateway_PushDownloadServer) (err error) {
	_, connectionID, err := authenticatedPeer(stream.Context())
	if err != nil {
		return err
	}
	first, err := stream.Recv()
	if err != nil {
		return err
	}
	writer, err := g.hub.beginPush(connectionID, first.GetJobId())
	if err != nil {
		return jobStatus(err)
	}
	defer func() { _ = writer.CloseWithError(err) }()
	expected := first.GetTotalSize()
	if expected < 0 || len(first.GetChunk()) != 0 {
		return status.Error(codes.InvalidArgument, "download must start with a non-negative size and no data")
	}
	written := int64(0)
	for {
		message, recvErr := stream.Recv()
		if errors.Is(recvErr, io.EOF) {
			break
		}
		if recvErr != nil {
			return recvErr
		}
		if message.GetJobId() != "" || message.GetTotalSize() != 0 {
			return status.Error(codes.InvalidArgument, "download metadata is only allowed in the first message")
		}
		chunk := message.GetChunk()
		if len(chunk) > chunkSize {
			return status.Error(codes.ResourceExhausted, "download chunk exceeds the byte limit")
		}
		if int64(len(chunk)) > expected-written {
			return status.Error(codes.InvalidArgument, "download exceeds its declared size")
		}
		if len(chunk) == 0 {
			continue
		}
		if _, err := writer.Write(chunk); err != nil {
			return status.Error(codes.Canceled, "download receiver closed")
		}
		written += int64(len(chunk))
	}
	if written != expected {
		return status.Errorf(codes.InvalidArgument, "download stream ended after %d of %d bytes", written, expected)
	}
	return stream.SendAndClose(&internalpb.Ack{Ok: true})
}

func jobStatus(err error) error {
	switch {
	case errors.Is(err, errUnknownJob):
		return status.Error(codes.NotFound, "unknown job for this connection")
	case errors.Is(err, errJobCompleted):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, errInvalidJobResult):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "internal job result failed")
	}
}
