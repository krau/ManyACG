package infra

import (
	"context"
	"slices"
	"sync"

	"github.com/krau/ManyACG/internal/common/httpclient"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/database"
	"github.com/krau/ManyACG/internal/infra/internalrpc"
	"github.com/krau/ManyACG/internal/infra/kvstor"
	"github.com/krau/ManyACG/internal/infra/source"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/samber/oops"
)

type Resources struct {
	KV             kvstor.KVStore
	Dispatcher     *internalrpc.Dispatcher
	InternalServer *internalrpc.Server
	Agent          *internalrpc.Agent
	SourceNamers   map[shared.SourceType]source.ArtworkNamer
	closers        []func() error
	closeOnce      sync.Once
	closeErr       error
}

func (r *Resources) Close() error {
	r.closeOnce.Do(func() {
		var errs []error
		for _, v := range slices.Backward(r.closers) {
			if err := v(); err != nil {
				errs = append(errs, err)
			}
		}
		r.closeErr = oops.Join(errs...)
	})
	return r.closeErr
}

func Init(ctx context.Context, cfg runtimecfg.Config) (*Resources, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	resources := &Resources{}
	mainNode := cfg.App.IsMain()
	readable := cfg.App.Capabilities.Has(runtimecfg.CapabilityRead)
	if mainNode {
		store, err := kvstor.Open(cfg.KVDB)
		if err != nil {
			return nil, err
		}
		resources.KV = store
		resources.closers = append(resources.closers, store.Close)
	}
	var localFetch source.ArtworkFetcher
	if mainNode || cfg.App.Capabilities.Has(runtimecfg.CapabilityFetch) {
		initSource(resources.KV)
		localFetch = source.LocalFetcher(source.Sources())
	}
	if readable || mainNode {
		resources.SourceNamers = sourceNamers(cfg.Source)
		if err := initStorage(ctx); err != nil {
			return nil, oops.Join(err, resources.Close())
		}
		database.Init(ctx, cfg)
		resources.closers = append(resources.closers, database.Default().Close)
	}
	hub := internalrpc.NewHub()
	resources.Dispatcher = internalrpc.NewDispatcher(hub, localFetch, httpclient.LocalFetcher(), resources.SourceNamers, cfg)
	if cfg.Internal.Gateway.Addr != "" {
		server, err := internalrpc.NewServer(cfg.Internal, hub)
		if err != nil {
			return nil, oops.Join(err, resources.Close())
		}
		resources.InternalServer = server
		resources.closers = append(resources.closers, server.Close)
	}
	if cfg.Internal.Agent.Gateway != "" {
		agent, err := internalrpc.NewAgent(cfg.Internal, cfg.App.Capabilities, localFetch)
		if err != nil {
			return nil, oops.Join(err, resources.Close())
		}
		resources.Agent = agent
	}
	return resources, nil
}
