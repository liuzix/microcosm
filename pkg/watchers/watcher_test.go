package watchers

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
)

// Asserts that WatchHelper implements Watcher
var _ Watcher[int, string] = &WatchHelper[int, string]{}

type numberWatcher struct {
	h    *WatchHelper[int, int]
	last atomic.Int64
}

func newNumberWatcher() *numberWatcher {
	ret := &numberWatcher{}
	ret.h = NewWatchHelper[int, int](func(ctx context.Context) (int, error) {
		return int(ret.last.Load()), nil
	})
	return ret
}

func (w *numberWatcher) TriggerUpdate() error {
	return w.h.DoEvent(context.Background(), func(ctx context.Context) (int, error) {
		return int(w.last.Add(1)), nil
	})
}

const (
	numUpdater    = 10
	numSubscriber = 10
	updateTimes   = 10000
)

func TestNumberWatcher(t *testing.T) {
	t.Parallel()

	watcher := newNumberWatcher()

	var wg sync.WaitGroup
	for i := 0; i < numUpdater; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < updateTimes; j++ {
				err := watcher.TriggerUpdate()
				require.NoError(t, err)
			}
		}()
	}

	for i := 0; i < numSubscriber; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			last, ch, cancel, err := watcher.h.Watch(context.Background())
			require.NoError(t, err)

			defer cancel()

			for ev := range ch {
				require.Equal(t, last+1, ev)
				last = ev
				if ev == numUpdater*updateTimes-1 {
					return
				}
			}
		}()
	}

	wg.Wait()
}
