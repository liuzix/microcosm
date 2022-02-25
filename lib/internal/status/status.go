package status

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pingcap/errors"
	"github.com/pingcap/tiflow/dm/pkg/log"
	"github.com/pingcap/tiflow/pkg/workerpool"
	"go.uber.org/atomic"
	"go.uber.org/zap"

	"github.com/hanfei1991/microcosm/lib/common"
	"github.com/hanfei1991/microcosm/lib/internal/metahelpers"
	"github.com/hanfei1991/microcosm/pkg/clock"
	derror "github.com/hanfei1991/microcosm/pkg/errors"
	"github.com/hanfei1991/microcosm/pkg/p2p"
)

type senderFSMState = int32

const (
	senderFSMIdle = senderFSMState(iota + 1)
	senderFSMPending
	senderFSMSending
)

// Sender is used for a worker to send its status to its master.
type Sender struct {
	workerID common.WorkerID

	workerMetaClient *metahelpers.WorkerMetadataClient
	messageSender    p2p.MessageSender
	masterClient     *masterClient

	// The following describes the FSM state transitions.
	//
	// [Idle] ==SendStatus==> [Pending] ==sendStatus(pool)==> [Sending]
	// [Sending] ==sent successfully==> [Idle]
	// [Sending] ==need to retry==> [Pending]
	fsmState           atomic.Int32
	lastUnsentStatus   *common.WorkerStatus
	lastUnsentStatusMu sync.RWMutex

	pool workerpool.AsyncPool

	errCh chan error
}

// NewStatusSender returns a new StatusSender.
// NOTE: the pool is owned by the caller.
func NewStatusSender(
	workerID common.WorkerID,
	masterClient *masterClient,
	workerMetaClient *metahelpers.WorkerMetadataClient,
	messageSender p2p.MessageSender,
	pool workerpool.AsyncPool,
) *Sender {
	return &Sender{
		workerID:         workerID,
		workerMetaClient: workerMetaClient,
		messageSender:    messageSender,
		masterClient:     masterClient,
		fsmState:         *atomic.NewInt32(senderFSMIdle),
		pool:             pool,
		errCh:            make(chan error, 1),
	}
}

// Tick should be called periodically to drive the logic internal to StatusSender.
func (s *Sender) Tick(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return errors.Trace(ctx.Err())
	case err := <-s.errCh:
		return errors.Trace(err)
	default:
	}

	if s.fsmState.Load() == senderFSMPending {
		if err := s.sendStatus(ctx); err != nil {
			return errors.Trace(err)
		}
	}

	return nil
}

func (s *Sender) getLastUnsentStatus() *common.WorkerStatus {
	s.lastUnsentStatusMu.Lock()
	defer s.lastUnsentStatusMu.Unlock()
	return s.lastUnsentStatus
}

func (s *Sender) setLastUnsentStatus(status *common.WorkerStatus) {
	s.lastUnsentStatusMu.Lock()
	defer s.lastUnsentStatusMu.Unlock()
	s.lastUnsentStatus = status
}

// SendStatus is used by the business logic in a worker to notify its master
// of a status change.
// This function is non-blocking and if any error occurred during or after network IO,
// the subsequent Tick will return an error.
func (s *Sender) SendStatus(ctx context.Context, status common.WorkerStatus) error {
	if s.fsmState.Load() != senderFSMIdle {
		return derror.ErrWorkerUpdateStatusTryAgain.GenWithStackByArgs()
	}
	s.setLastUnsentStatus(&status)
	if old := s.fsmState.Swap(senderFSMPending); old != senderFSMIdle {
		log.L().Panic("StatusSender: unexpected fsm state",
			zap.Int32("old-fsm-state", old))
	}
	if err := s.sendStatus(ctx); err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (s *Sender) sendStatus(ctx context.Context) error {
	err := s.pool.Go(ctx, func() {
		if !s.fsmState.CAS(senderFSMPending, senderFSMSending) {
			return
		}

		status := s.getLastUnsentStatus()
		if err := s.workerMetaClient.Store(ctx, s.workerID, status); err != nil {
			s.onError(err)
		}

		ok, err := s.messageSender.SendToNode(
			ctx,
			s.masterClient.MasterNode(),
			workerStatusUpdatedTopic(s.masterClient.MasterID(), s.masterClient.workerID),
			&workerStatusUpdatedMessage{Epoch: s.masterClient.Epoch()})
		if err != nil {
			s.onError(err)
		}

		if !ok {
			// handle retry
			if old := s.fsmState.Swap(senderFSMPending); old != senderFSMSending {
				log.L().Panic("StatusSender: unexpected fsm state",
					zap.Int32("old-fsm-state", old))
			}
			return
		}

		if old := s.fsmState.Swap(senderFSMIdle); old != senderFSMSending {
			log.L().Panic("StatusSender: unexpected fsm state",
				zap.Int32("old-fsm-state", old))
		}
		s.setLastUnsentStatus(nil)
	})
	return errors.Trace(err)
}

func (s *Sender) onError(err error) {
	select {
	case s.errCh <- err:
	default:
		log.L().Warn("error is dropped because errCh is full",
			zap.Error(err))
	}
}

func workerStatusUpdatedTopic(masterID common.MasterID, workerID common.WorkerID) string {
	return fmt.Sprintf("worker-status-updated-%s-%s", masterID, workerID)
}

type workerStatusUpdatedMessage struct {
	Epoch common.Epoch
}

// Receiver is used by a master to receive the latest status update from **a** worker.
type Receiver struct {
	workerID common.WorkerID

	workerMetaClient      *metahelpers.WorkerMetadataClient
	messageHandlerManager p2p.MessageHandlerManager

	statusMu    sync.RWMutex
	statusCache common.WorkerStatus

	hasPendingNotification atomic.Bool
	lastStatusUpdated      atomic.Time
	isLoading              atomic.Bool

	errCh chan error

	epoch common.Epoch

	pool workerpool.AsyncPool

	clock clock.Clock
}

// NewStatusReceiver returns a new StatusReceiver
// NOTE: the messageHandlerManager is NOT owned by the StatusReceiver,
// and it only uses it to register a handler. It is not responsible
// for checking errors.
// NOTE: the pool is owned and managed by the caller.
func NewStatusReceiver(
	workerID common.WorkerID,
	workerMetaClient *metahelpers.WorkerMetadataClient,
	messageHandlerManager p2p.MessageHandlerManager,
	epoch common.Epoch,
	pool workerpool.AsyncPool,
	clock clock.Clock,
) *Receiver {
	return &Receiver{
		workerID:              workerID,
		workerMetaClient:      workerMetaClient,
		messageHandlerManager: messageHandlerManager,
		epoch:                 epoch,
		pool:                  pool,
		errCh:                 make(chan error, 1),
		clock:                 clock,
	}
}

// Init should be called to initialize a StatusReceiver.
// NOTE: this function can be blocked by IO to the metastore.
func (r *Receiver) Init(ctx context.Context) error {
	topic := workerStatusUpdatedTopic(r.workerMetaClient.MasterID(), r.workerID)
	ok, err := r.messageHandlerManager.RegisterHandler(
		ctx,
		topic,
		&workerStatusUpdatedMessage{},
		func(sender p2p.NodeID, value p2p.MessageValue) error {
			log.L().Debug("Received workerStatusUpdatedMessage",
				zap.String("sender", sender),
				zap.Any("value", value))

			msg := value.(*workerStatusUpdatedMessage)
			if msg.Epoch != r.epoch {
				return nil
			}
			r.hasPendingNotification.Store(true)
			log.L().Info("notification stored")
			return nil
		})
	if err != nil {
		return errors.Trace(err)
	}
	if !ok {
		log.L().Panic("duplicate handlers", zap.String("topic", topic))
	}

	initStatus, err := r.workerMetaClient.Load(ctx, r.workerID)
	if err != nil {
		return errors.Trace(err)
	}

	r.statusMu.Lock()
	defer r.statusMu.Unlock()

	r.statusCache = *initStatus
	r.lastStatusUpdated.Store(r.clock.Now())

	return nil
}

// Status returns the latest status of the worker.
func (r *Receiver) Status() common.WorkerStatus {
	r.statusMu.RLock()
	defer r.statusMu.RUnlock()

	return r.statusCache
}

// Tick should be called periodically to drive the logic internal to StatusReceiver.
func (r *Receiver) Tick(ctx context.Context) error {
	if r.hasPendingNotification.Load() {
		log.L().Info("has pending notification")
	}
	// TODO make the time interval configurable
	needFetchStatus := r.hasPendingNotification.Swap(false) ||
		r.clock.Since(r.lastStatusUpdated.Load()) > time.Second*10

	if !needFetchStatus {
		return nil
	}

	if r.isLoading.Swap(true) {
		// A load is already in progress.
		return nil
	}

	err := r.pool.Go(ctx, func() {
		defer r.isLoading.Store(false)

		status, err := r.workerMetaClient.Load(ctx, r.workerID)
		if err != nil {
			r.onError(err)
			return
		}

		r.statusMu.Lock()
		defer r.statusMu.Unlock()

		r.statusCache = *status
		r.lastStatusUpdated.Store(r.clock.Now())
	})
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}

func (r *Receiver) Close(ctx context.Context) error {
	topic := workerStatusUpdatedTopic(r.workerMetaClient.MasterID(), r.workerID)
	ok, err := r.messageHandlerManager.UnregisterHandler(ctx, topic)
	if err != nil {
		return errors.Trace(err)
	}
	if !ok {
		log.L().Warn("message handler for topic does not exist",
			zap.String("topic", topic))
	}
	return nil
}

func (r *Receiver) onError(err error) {
	select {
	case r.errCh <- err:
	default:
		log.L().Warn("error is dropped because errCh is full",
			zap.Error(err))
	}
}
