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

	sched := newScheduler(jobs, specs, "tenant-default", nil)
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

// denyJobLocker menolak semua lock — mensimulasikan instance lain memegang lock.
type denyJobLocker struct {
	calls atomic.Int32
}

func (d *denyJobLocker) TryLock(context.Context, string) (func(), bool, error) {
	d.calls.Add(1)
	return func() {}, false, nil
}

func TestScheduler_SkipsJobWhenLockDenied(t *testing.T) {
	var runs atomic.Int32
	locker := &denyJobLocker{}

	sched := newScheduler(schedulerJobs{
		snapshot: func(ctx context.Context) error {
			runs.Add(1)
			return nil
		},
	}, schedulerSpecs{
		snapshot: "@every 30ms",
	}, "tenant-default", locker)

	sched.Start()
	defer sched.Stop()

	assert.Eventually(t, func() bool {
		return locker.calls.Load() >= 1
	}, 1*time.Second, 10*time.Millisecond)
	assert.Zero(t, runs.Load(), "job tidak boleh berjalan saat lock dipegang instance lain")
}
