package internalrpc

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/krau/ManyACG/internal/infra/internalrpc/internalpb"
)

const (
	protocolVersion = 3
	assignQueueSize = 32
)

var (
	errNoAgent           = errors.New("no idle internal agent available")
	errSessionClosed     = errors.New("internal agent session closed")
	errSessionReplaced   = errors.New("internal agent session replaced by a newer connection")
	errUnknownJob        = errors.New("unknown internal job")
	errJobCompleted      = errors.New("internal job already has a result")
	errInvalidJobResult  = errors.New("invalid internal job result")
	errInvalidJobRequest = errors.New("internal job request is invalid")
)

type session struct {
	id           string
	nodeID       string
	connectionID string
	fetch        bool
	download     bool
	concurrency  uint32
	jobs         map[string]*job
	send         chan *internalpb.Assignment
	closed       chan struct{}
	closeOnce    sync.Once
}

func (s *session) supports(capability internalpb.Capability) bool {
	switch capability {
	case internalpb.Capability_CAPABILITY_FETCH:
		return s.fetch
	case internalpb.Capability_CAPABILITY_DOWNLOAD:
		return s.download
	default:
		return false
	}
}

func (s *session) close() {
	s.closeOnce.Do(func() { close(s.closed) })
}

func (s *session) sendAssignment(ctx context.Context, assignment *internalpb.Assignment) error {
	select {
	case s.send <- assignment:
		return nil
	case <-s.closed:
		return errSessionClosed
	case <-ctx.Done():
		return ctx.Err()
	}
}

type job struct {
	id        string
	session   *session
	result    chan *internalpb.JobCompletion
	reader    *io.PipeReader
	writer    *io.PipeWriter
	pushing   bool
	completed bool
}

func (j *job) abort(err error) {
	if j.writer != nil {
		_ = j.writer.CloseWithError(err)
		return
	}
	select {
	case j.result <- &internalpb.JobCompletion{
		JobId:  j.id,
		Result: &internalpb.JobCompletion_Failure{Failure: failureFromError(err)},
	}:
	default:
	}
}

func (j *job) cancel() {
	assignment := &internalpb.Assignment{
		Payload: &internalpb.Assignment_Cancel{Cancel: &internalpb.CancelJob{JobId: j.id}},
	}
	select {
	case j.session.send <- assignment:
	default:
	}
}

type Hub struct {
	mu       sync.Mutex
	sessions []*session
	jobs     map[string]*job
	next     uint64
}

func NewHub() *Hub {
	return &Hub{jobs: make(map[string]*job)}
}

func (h *Hub) register(nodeID, connectionID string, reg *internalpb.Register) (*session, error) {
	concurrency := reg.GetConcurrency()
	if concurrency == 0 {
		concurrency = 1
	}
	attached := &session{
		id:           newID(),
		nodeID:       nodeID,
		connectionID: connectionID,
		concurrency:  concurrency,
		jobs:         make(map[string]*job),
		send:         make(chan *internalpb.Assignment, assignQueueSize),
		closed:       make(chan struct{}),
	}
	for _, capability := range reg.GetCapabilities() {
		switch capability {
		case internalpb.Capability_CAPABILITY_FETCH:
			attached.fetch = true
		case internalpb.Capability_CAPABILITY_DOWNLOAD:
			attached.download = true
		}
	}
	h.mu.Lock()
	var replaced []*session
	kept := h.sessions[:0]
	for _, existing := range h.sessions {
		if existing.nodeID == nodeID || existing.connectionID == connectionID {
			replaced = append(replaced, existing)
			continue
		}
		kept = append(kept, existing)
	}
	h.sessions = append(kept, attached)
	var orphans []*job
	for _, old := range replaced {
		for id, orphan := range old.jobs {
			delete(h.jobs, id)
			orphans = append(orphans, orphan)
		}
		old.jobs = nil
	}
	h.mu.Unlock()
	for _, old := range replaced {
		old.close()
	}
	for _, orphan := range orphans {
		orphan.abort(errSessionReplaced)
	}
	return attached, nil
}

func (h *Hub) closeAll(reason error) {
	h.mu.Lock()
	sessions := h.sessions
	h.sessions = nil
	var orphans []*job
	for _, session := range sessions {
		for id, orphan := range session.jobs {
			delete(h.jobs, id)
			orphans = append(orphans, orphan)
		}
		session.jobs = nil
	}
	h.mu.Unlock()
	for _, session := range sessions {
		session.close()
	}
	for _, orphan := range orphans {
		orphan.abort(reason)
	}
}

func (h *Hub) dropSession(session *session) {
	h.mu.Lock()
	for i, existing := range h.sessions {
		if existing == session {
			h.sessions = append(h.sessions[:i], h.sessions[i+1:]...)
			break
		}
	}
	orphans := session.jobs
	for id := range orphans {
		delete(h.jobs, id)
	}
	session.jobs = nil
	h.mu.Unlock()
	session.close()
	for _, orphan := range orphans {
		orphan.abort(errSessionClosed)
	}
}

func (h *Hub) dispatch(capability internalpb.Capability) (*job, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	count := uint64(len(h.sessions))
	for i := range count {
		session := h.sessions[(h.next+i)%count]
		if !session.supports(capability) || len(session.jobs) >= int(session.concurrency) {
			continue
		}
		h.next += i + 1
		dispatchJob := &job{id: newID(), session: session}
		if capability == internalpb.Capability_CAPABILITY_DOWNLOAD {
			dispatchJob.reader, dispatchJob.writer = io.Pipe()
		} else {
			dispatchJob.result = make(chan *internalpb.JobCompletion, 1)
		}
		h.jobs[dispatchJob.id] = dispatchJob
		session.jobs[dispatchJob.id] = dispatchJob
		return dispatchJob, nil
	}
	return nil, errNoAgent
}

func (h *Hub) release(dispatchJob *job) {
	h.mu.Lock()
	delete(h.jobs, dispatchJob.id)
	delete(dispatchJob.session.jobs, dispatchJob.id)
	h.mu.Unlock()
	if dispatchJob.reader != nil {
		_ = dispatchJob.reader.Close()
		_ = dispatchJob.writer.Close()
	}
}

func (h *Hub) complete(connectionID string, completion *internalpb.JobCompletion) error {
	h.mu.Lock()
	dispatchJob := h.jobs[completion.GetJobId()]
	if dispatchJob == nil || dispatchJob.session.connectionID != connectionID {
		h.mu.Unlock()
		return errUnknownJob
	}
	if dispatchJob.completed {
		h.mu.Unlock()
		return errJobCompleted
	}
	failure := completion.GetFailure()
	validFailure := failure != nil && failure.GetCode() >= internalpb.FailureCode_FAILURE_CODE_UNSUPPORTED_SOURCE &&
		failure.GetCode() <= internalpb.FailureCode_FAILURE_CODE_INTERNAL && len(failure.GetMessage()) <= maxFailureMessageBytes
	validResult := validFailure || dispatchJob.writer == nil && completion.GetArtwork() != nil
	dispatchJob.completed = true
	h.mu.Unlock()
	if !validResult {
		err := fmt.Errorf("%w: result does not match the assigned task", errInvalidJobResult)
		dispatchJob.abort(err)
		return err
	}
	if dispatchJob.writer != nil {
		_ = dispatchJob.writer.CloseWithError(failureToError(failure))
		return nil
	}
	select {
	case dispatchJob.result <- completion:
	default:
	}
	return nil
}

func (h *Hub) beginPush(connectionID, jobID string) (*io.PipeWriter, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	dispatchJob := h.jobs[jobID]
	if dispatchJob == nil || dispatchJob.writer == nil || dispatchJob.session.connectionID != connectionID {
		return nil, errUnknownJob
	}
	if dispatchJob.pushing || dispatchJob.completed {
		return nil, errJobCompleted
	}
	dispatchJob.pushing = true
	return dispatchJob.writer, nil
}

func newID() string {
	var buffer [16]byte
	_, _ = rand.Read(buffer[:])
	return hex.EncodeToString(buffer[:])
}
