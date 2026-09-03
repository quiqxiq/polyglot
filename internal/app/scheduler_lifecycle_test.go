package app

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScheduler_Lifecycle(t *testing.T) {
	var snapshotRuns atomic.Int32

	jobs := schedulerJobs{
		snapshot: func(ctx context.Context) error {
			snapshotRuns.Add(1)
			return nil
		},
	}
	specs := schedulerSpecs{
		snapshot: "@every 50ms",
	}

	sched := newScheduler(jobs, specs, "tenant-default")
	require.NotNil(t, sched)

	sched.Start()

	// Wait for at least one run
	assert.Eventually(t, func() bool {
		return snapshotRuns.Load() >= 1
	}, 1*time.Second, 20*time.Millisecond)

	sched.Stop()
	currentCount := snapshotRuns.Load()

	// After stop, count should not increase significantly
	time.Sleep(120 * time.Millisecond)
	assert.LessOrEqual(t, snapshotRuns.Load(), currentCount+1)
}
