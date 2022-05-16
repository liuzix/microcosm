package watchers

import (
	"context"
	"sync"

	"github.com/pingcap/errors"
	"go.uber.org/atomic"
)

const watchChannelSize = 16

type Watcher[S, E any] interface {
	Watch(ctx context.Context) (snap S, eventCh <-chan E, cancel func(), retErr error)
}

type WatchHelper[S, E any] struct {
	mu       sync.RWMutex
	watchChs map[int]chan E
	nextID   int

	getSnap func(ctx context.Context) (S, error)
}

func NewWatchHelper[S, E any](getSnap func(ctx context.Context) (S, error)) *WatchHelper[S, E] {
	return &WatchHelper[S, E]{
		watchChs: make(map[int]chan E),
		getSnap:  getSnap,
	}
}

func (h *WatchHelper[S, E]) Watch(
	ctx context.Context,
) (snap S, eventCh <-chan E, cancel func(), retErr error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	wid := h.nextID
	h.nextID++
	ch := make(chan E, watchChannelSize)
	h.watchChs[wid] = ch

	defer func() {
		if retErr != nil {
			delete(h.watchChs, wid)
		}
	}()

	snap, retErr = h.getSnap(ctx)
	if retErr != nil {
		return
	}

	var (
		cancelOnce sync.Once
		canceled   atomic.Bool
	)
	cancel = func() {
		cancelOnce.Do(func() {
			if canceled.Swap(true) {
				// Already canceled
				return
			}

			h.mu.Lock()
			defer h.mu.Unlock()

			close(ch)
			delete(h.watchChs, wid)
		})
	}

	eventCh = ch
	return
}

func (h *WatchHelper[S, E]) DoEvent(ctx context.Context, action func(context.Context) (E, error)) error {
	h.mu.RLock()

	var cleanUpChList []int
	defer func() {
		h.mu.Lock()
		defer h.mu.Unlock()

		for _, wid := range cleanUpChList {
			ch, exists := h.watchChs[wid]
			if !exists {
				continue
			}
			close(ch)
			delete(h.watchChs, wid)
		}
	}()

	defer h.mu.RUnlock()

	event, err := action(ctx)
	if err != nil {
		return err
	}

	for wid, ch := range h.watchChs {
		select {
		case <-ctx.Done():
			cleanUpChList = append(cleanUpChList, wid)
			return errors.Trace(err)
		case ch <- event:
		}
	}

	return nil
}
