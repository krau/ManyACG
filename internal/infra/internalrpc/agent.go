package internalrpc

import (
	"context"
	"errors"
	"io"
	"math/rand/v2"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/krau/ManyACG/internal/common/httpclient"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/internalrpc/internalpb"
	"github.com/krau/ManyACG/internal/infra/source"
	"github.com/krau/ManyACG/internal/shared/errs"
	"github.com/krau/ManyACG/pkg/log"
	"golang.org/x/crypto/ssh"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

const (
	chunkSize         = 1 << 20
	minReconnectDelay = time.Second
	maxReconnectDelay = 30 * time.Second
)

type Agent struct {
	cfg    runtimecfg.InternalConfig
	name   string
	caps   []internalpb.Capability
	fetch  source.ArtworkFetcher
	signer ssh.Signer
}

func NewAgent(cfg runtimecfg.InternalConfig, capabilities runtimecfg.Capabilities, fetch source.ArtworkFetcher) (*Agent, error) {
	signer, err := readNodeSigner(cfg.Agent.PrivateKey)
	if err != nil {
		return nil, err
	}
	var agentCaps []internalpb.Capability
	if capabilities.Has(runtimecfg.CapabilityFetch) {
		agentCaps = append(agentCaps, internalpb.Capability_CAPABILITY_FETCH)
	}
	if capabilities.Has(runtimecfg.CapabilityDownload) {
		agentCaps = append(agentCaps, internalpb.Capability_CAPABILITY_DOWNLOAD)
	}
	name := strings.TrimSpace(cfg.Agent.Name)
	if name == "" {
		name, _ = os.Hostname()
	}
	return &Agent{
		cfg:    cfg,
		name:   name,
		caps:   agentCaps,
		fetch:  fetch,
		signer: signer,
	}, nil
}

func (a *Agent) Run(ctx context.Context) {
	if len(a.caps) == 0 {
		return
	}
	delay := minReconnectDelay
	for ctx.Err() == nil {
		started := time.Now()
		if err := a.session(ctx); err != nil && ctx.Err() == nil {
			log.Warn("internal agent session ended", "gateway", a.cfg.Agent.Gateway, "name", a.name, "err", err)
		}
		if ctx.Err() != nil {
			return
		}
		if time.Since(started) > time.Minute {
			delay = minReconnectDelay
		}
		jitter := time.Duration(rand.Int64N(int64(delay)/2 + 1))
		select {
		case <-ctx.Done():
			return
		case <-time.After(delay + jitter):
		}
		if delay < maxReconnectDelay {
			delay *= 2
		}
	}
}

func (a *Agent) session(ctx context.Context) error {
	var transport credentials.TransportCredentials = insecure.NewCredentials()
	if a.cfg.Agent.SSH {
		transport = newSSHClientCredentials(a.cfg.Agent.Gateway, a.cfg.Agent.KnownHosts, a.signer)
	}
	options := []grpc.DialOption{
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithDisableRetry(),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMessageSize), grpc.MaxCallSendMsgSize(maxMessageSize)),
		grpc.WithTransportCredentials(transport),
	}
	conn, err := grpc.NewClient(a.cfg.Agent.Gateway, options...)
	if err != nil {
		return err
	}
	defer conn.Close()
	client := internalpb.NewGatewayClient(conn)
	stream, err := client.Attach(ctx)
	if err != nil {
		return err
	}
	register := &internalpb.Register{
		ProtocolVersion: protocolVersion,
		Name:            a.name,
		Capabilities:    a.caps,
		Concurrency:     uint32(a.cfg.Agent.Concurrency),
		PublicKey:       a.signer.PublicKey().Marshal(),
	}
	if err := stream.Send(&internalpb.AgentMessage{Payload: &internalpb.AgentMessage_Register{Register: register}}); err != nil {
		return err
	}
	first, err := stream.Recv()
	if err != nil {
		return err
	}
	if !a.cfg.Agent.SSH {
		proof, err := authenticationProof(a.signer, first.GetChallenge().GetNonce(), register)
		if err != nil {
			return err
		}
		if err := stream.Send(&internalpb.AgentMessage{Payload: &internalpb.AgentMessage_Proof{Proof: proof}}); err != nil {
			return err
		}
		first, err = stream.Recv()
		if err != nil {
			return err
		}
	}
	accepted := first.GetAccepted()
	if accepted == nil {
		return errors.New("internal gateway did not accept the session")
	}
	log.Info("internal agent registered",
		"gateway", a.cfg.Agent.Gateway,
		"name", a.name,
		"session", accepted.GetSessionId(),
		"concurrency", accepted.GetConcurrency(),
	)
	sessionCtx, cancelSession := context.WithCancel(ctx)
	var running sync.WaitGroup
	defer func() {
		cancelSession()
		running.Wait()
	}()
	var mu sync.Mutex
	cancels := make(map[string]context.CancelFunc)
	for {
		assignment, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		switch payload := assignment.Payload.(type) {
		case *internalpb.Assignment_Job:
			assignedJob := payload.Job
			jobCtx, cancelJob := context.WithCancel(sessionCtx)
			mu.Lock()
			cancels[assignedJob.GetJobId()] = cancelJob
			mu.Unlock()
			running.Go(func() {
				defer func() {
					mu.Lock()
					delete(cancels, assignedJob.GetJobId())
					mu.Unlock()
					cancelJob()
				}()
				a.execute(jobCtx, client, assignedJob)
			})
		case *internalpb.Assignment_Cancel:
			mu.Lock()
			cancelJob := cancels[payload.Cancel.GetJobId()]
			mu.Unlock()
			if cancelJob != nil {
				cancelJob()
			}
		}
	}
}

func (a *Agent) execute(ctx context.Context, client internalpb.GatewayClient, assignedJob *internalpb.Job) {
	if deadline := assignedJob.GetDeadlineUnixMs(); deadline > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, time.UnixMilli(deadline))
		defer cancel()
	}
	switch task := assignedJob.Task.(type) {
	case *internalpb.Job_FetchArtwork:
		if a.fetch == nil {
			a.completeFailure(ctx, client, assignedJob.GetJobId(), errs.ErrUnsupportedSource)
			return
		}
		requestedURL := task.FetchArtwork.GetSourceUrl()
		artwork, err := a.fetch.GetArtworkInfo(ctx, requestedURL)
		if err != nil {
			a.completeFailure(ctx, client, assignedJob.GetJobId(), err)
			return
		}
		if artwork != nil {
			artwork.SourceURL = requestedURL
		}
		if _, err := client.CompleteJob(ctx, &internalpb.JobCompletion{
			JobId:  assignedJob.GetJobId(),
			Result: &internalpb.JobCompletion_Artwork{Artwork: artworkToProto(artwork)},
		}); err != nil {
			log.Warn("internal agent failed to return artwork", "job", assignedJob.GetJobId(), "err", err)
		}
	case *internalpb.Job_DownloadOriginFile:
		limit := task.DownloadOriginFile.GetMaxBytes()
		if limit <= 0 || limit > a.cfg.Dispatch.MaxDownloadBytes {
			limit = a.cfg.Dispatch.MaxDownloadBytes
		}
		if err := a.pushDownload(ctx, client, assignedJob.GetJobId(), task.DownloadOriginFile.GetUrl(), limit); err != nil {
			a.completeFailure(ctx, client, assignedJob.GetJobId(), err)
		}
	}
}

func (a *Agent) completeFailure(ctx context.Context, client internalpb.GatewayClient, jobID string, cause error) {
	if ctx.Err() != nil {
		return
	}
	if _, err := client.CompleteJob(ctx, &internalpb.JobCompletion{
		JobId:  jobID,
		Result: &internalpb.JobCompletion_Failure{Failure: failureFromError(cause)},
	}); err != nil {
		log.Warn("internal agent failed to report job failure", "job", jobID, "err", err)
	}
}

func (a *Agent) pushDownload(ctx context.Context, client internalpb.GatewayClient, jobID, url string, maxBytes int64) error {
	file, err := httpclient.DownloadPublicOriginFile(ctx, url, maxBytes)
	if err != nil {
		return err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return err
	}
	streamCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	stream, err := client.PushDownload(streamCtx)
	if err != nil {
		return err
	}
	if err := stream.Send(&internalpb.DownloadChunk{JobId: jobID, TotalSize: info.Size()}); err != nil {
		return err
	}
	buffer := make([]byte, chunkSize)
	for {
		read, readErr := file.Read(buffer)
		if read > 0 {
			if err := stream.Send(&internalpb.DownloadChunk{Chunk: buffer[:read]}); err != nil {
				return err
			}
		}
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return readErr
		}
	}
	_, err = stream.CloseAndRecv()
	return err
}
