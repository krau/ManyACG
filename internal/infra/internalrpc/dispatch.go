package internalrpc

import (
	"context"
	"errors"
	"io"
	"math"
	"os"
	"time"

	"github.com/krau/ManyACG/internal/common/httpclient"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/internalrpc/internalpb"
	"github.com/krau/ManyACG/internal/infra/source"
	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/log"
)

type Dispatcher struct {
	hub             *Hub
	localFetch      source.ArtworkFetcher
	localDownload   httpclient.Fetcher
	fetchTimeout    time.Duration
	downloadTimeout time.Duration
	namers          map[shared.SourceType]source.ArtworkNamer
}

func NewDispatcher(
	hub *Hub,
	localFetch source.ArtworkFetcher,
	localDownload httpclient.Fetcher,
	namers map[shared.SourceType]source.ArtworkNamer,
	cfg runtimecfg.Config,
) *Dispatcher {
	return &Dispatcher{
		hub:             hub,
		localFetch:      localFetch,
		localDownload:   localDownload,
		fetchTimeout:    time.Duration(cfg.Internal.Dispatch.FetchTimeout) * time.Second,
		downloadTimeout: time.Duration(cfg.Internal.Dispatch.DownloadTimeout) * time.Second,
		namers:          namers,
	}
}

func (d *Dispatcher) GetArtworkInfo(ctx context.Context, sourceURL string) (*dto.FetchedArtwork, error) {
	artwork, err := d.fetchFromAgent(ctx, sourceURL)
	if err == nil {
		return artwork, nil
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if !errors.Is(err, errNoAgent) {
		log.Warn("internal agent artwork fetch failed, falling back to local sources", "source_url", sourceURL, "err", err)
	}
	if d.localFetch == nil {
		return nil, err
	}
	return d.localFetch.GetArtworkInfo(ctx, sourceURL)
}

func (d *Dispatcher) FetchToFile(ctx context.Context, url, path string) error {
	err := d.downloadFromAgent(ctx, url, path)
	if err == nil {
		return nil
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if !errors.Is(err, errNoAgent) {
		log.Warn("internal agent download failed, falling back to local download", "url", url, "err", err)
		if removeErr := os.Remove(path); removeErr != nil && !os.IsNotExist(removeErr) {
			log.Warn("failed to remove partial download", "path", path, "err", removeErr)
		}
	}
	if d.localDownload == nil {
		return err
	}
	return d.localDownload.FetchToFile(ctx, url, path)
}

func (d *Dispatcher) fetchFromAgent(ctx context.Context, sourceURL string) (*dto.FetchedArtwork, error) {
	dispatchJob, err := d.hub.dispatch(internalpb.Capability_CAPABILITY_FETCH)
	if err != nil {
		return nil, err
	}
	defer d.hub.release(dispatchJob)
	jobCtx, cancel := context.WithTimeout(ctx, d.fetchTimeout)
	defer cancel()
	assignment := &internalpb.Assignment{
		Payload: &internalpb.Assignment_Job{Job: &internalpb.Job{
			JobId:     dispatchJob.id,
			TimeoutMs: timeoutMillis(d.fetchTimeout),
			Task: &internalpb.Job_FetchArtwork{FetchArtwork: &internalpb.FetchArtwork{
				SourceUrl: sourceURL,
			}},
		}},
	}
	if err := dispatchJob.session.sendAssignment(jobCtx, assignment); err != nil {
		return nil, err
	}
	select {
	case completion := <-dispatchJob.result:
		if failure := completion.GetFailure(); failure != nil {
			return nil, failureToError(failure)
		}
		return validateArtworkResult(sourceURL, completion.GetArtwork(), d.namers)
	case <-jobCtx.Done():
		dispatchJob.cancel()
		return nil, jobCtx.Err()
	}
}

func (d *Dispatcher) downloadFromAgent(ctx context.Context, url, path string) (err error) {
	dispatchJob, err := d.hub.dispatch(internalpb.Capability_CAPABILITY_DOWNLOAD)
	if err != nil {
		return err
	}
	defer d.hub.release(dispatchJob)
	jobCtx, cancel := context.WithTimeout(ctx, d.downloadTimeout)
	defer cancel()
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := file.Close(); err == nil {
			err = closeErr
		}
		if err != nil {
			dispatchJob.cancel()
		}
	}()
	stopAbort := context.AfterFunc(jobCtx, func() { dispatchJob.abort(jobCtx.Err()) })
	defer stopAbort()
	assignment := &internalpb.Assignment{
		Payload: &internalpb.Assignment_Job{Job: &internalpb.Job{
			JobId:     dispatchJob.id,
			TimeoutMs: timeoutMillis(d.downloadTimeout),
			Task: &internalpb.Job_DownloadOriginFile{DownloadOriginFile: &internalpb.DownloadOriginFile{
				Url: url,
			}},
		}},
	}
	if err := dispatchJob.session.sendAssignment(jobCtx, assignment); err != nil {
		return err
	}
	_, err = io.Copy(file, dispatchJob.reader)
	return err
}

func timeoutMillis(timeout time.Duration) uint32 {
	if timeout <= 0 {
		return 0
	}
	if millis := timeout.Milliseconds(); millis > math.MaxUint32 {
		return math.MaxUint32
	}
	return uint32(timeout.Milliseconds())
}
