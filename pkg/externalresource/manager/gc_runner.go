package manager

import (
	"context"
	"time"

	"github.com/pingcap/errors"
	"github.com/pingcap/log"
	"go.uber.org/zap"

	"github.com/hanfei1991/microcosm/pkg/clock"
	resModel "github.com/hanfei1991/microcosm/pkg/externalresource/resourcemeta/model"
	pkgOrm "github.com/hanfei1991/microcosm/pkg/orm"
)

type gcHandlerFunc = func(ctx context.Context, meta *resModel.ResourceMeta) error

// DefaultGCRunner implements GCRunner.
type DefaultGCRunner struct {
	client     pkgOrm.ResourceClient
	gcHandlers map[resModel.ResourceType]gcHandlerFunc
	notifyCh   chan struct{}

	clock clock.Clock
}

// NewGCRunner returns a new GCRunner.
func NewGCRunner(
	client pkgOrm.ResourceClient,
	gcHandlers map[resModel.ResourceType]gcHandlerFunc,
) *DefaultGCRunner {
	return &DefaultGCRunner{
		client:     client,
		gcHandlers: gcHandlers,
		notifyCh:   make(chan struct{}, 1),
		clock:      clock.New(),
	}
}

// Run runs the GCRunner. It blocks until ctx is canceled.
func (r *DefaultGCRunner) Run(ctx context.Context) error {
	// TODO this will result in DB queries every 10 seconds.
	// This is a very naive strategy, we will modify the
	// algorithm after doing enough system testing.
	ticker := r.clock.Ticker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return errors.Trace(ctx.Err())
		case <-ticker.C:
		case <-r.notifyCh:
		}

		timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		err := r.gcOnce(timeoutCtx)
		cancel()

		if err != nil {
			log.L().Warn("resource GC encountered error", zap.Error(err))
		}
	}
}

// Notify is used to ask GCRunner to GC the next resource immediately.
// It is used when we have just marked a resource as gc_pending.
func (r *DefaultGCRunner) Notify() {
	select {
	case r.notifyCh <- struct{}{}:
	default:
	}
}

func (r *DefaultGCRunner) gcOnce(
	ctx context.Context,
) error {
	res, err := r.client.GetOneResourceForGC(ctx)
	if pkgOrm.IsNotFoundError(err) {
		// It is expected that sometimes we have
		// nothing to GC.
		return nil
	}
	if err != nil {
		return err
	}

	log.Info("start gc'ing resource", zap.Any("resource", res))
	if !res.GCPending {
		log.L().Panic("unexpected gc_pending = false")
	}

	tp, _, err := resModel.ParseResourcePath(res.ID)
	if err != nil {
		return err
	}

	handler, exists := r.gcHandlers[tp]
	if !exists {
		log.L().Warn("no gc handler is found for given resource type",
			zap.Any("resource-id", res.ID))
	}

	if err := handler(ctx, res); err != nil {
		return err
	}

	result, err := r.client.DeleteResource(ctx, res.ID)
	if err != nil {
		// If deletion fails, we do not need to retry for now.
		log.L().Warn("Failed to delete resource after GC", zap.Any("resource", res))
		return nil
	}
	if result.RowsAffected() == 0 {
		log.L().Warn("Resource is deleted unexpectedly", zap.Any("resource", res))
	}

	return nil
}