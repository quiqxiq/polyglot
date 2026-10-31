package routerjob_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainRouterJob "github.com/quixiq/polyglot/internal/domain/routerjob"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	uc "github.com/quixiq/polyglot/internal/usecase/routerjob"
)

var fixedNow = time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)

func seedJob(repo *mocktest.FakeRouterJobRepo, id string, jobType domainRouterJob.Type, attempts, maxAttempts int) {
	repo.Seed(domainRouterJob.Job{
		ID: id, TenantID: "tenant-default", DeviceID: "dev-1",
		JobType: jobType, Payload: []byte(`{"subscription_id":"sub-1"}`),
		IdempotencyKey: "key-" + id, Status: domainRouterJob.StatusQueued,
		Attempts: attempts, MaxAttempts: maxAttempts, NextRunAt: fixedNow.Add(-time.Minute),
		CreatedAt: fixedNow.Add(-time.Hour),
	})
}

func TestWorker_RunMarksSucceeded(t *testing.T) {
	repo := mocktest.NewFakeRouterJobRepo()
	seedJob(repo, "job-ok", domainRouterJob.TypeProvisionPPPoE, 0, 5)
	var handled int
	worker := uc.NewWorker(repo, map[domainRouterJob.Type]uc.Handler{
		domainRouterJob.TypeProvisionPPPoE: func(_ context.Context, _ domainRouterJob.Job) error {
			handled++
			return nil
		},
	}, nil).WithClock(func() time.Time { return fixedNow })

	res, err := worker.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, res.Processed)
	assert.Equal(t, 1, res.Succeeded)
	assert.Equal(t, 1, handled)

	job, err := repo.FindByID(context.Background(), "job-ok")
	require.NoError(t, err)
	assert.Equal(t, domainRouterJob.StatusSucceeded, job.Status)
}

func TestWorker_FailureSchedulesBackoff(t *testing.T) {
	repo := mocktest.NewFakeRouterJobRepo()
	seedJob(repo, "job-retry", domainRouterJob.TypeProvisionPPPoE, 0, 5)
	worker := uc.NewWorker(repo, map[domainRouterJob.Type]uc.Handler{
		domainRouterJob.TypeProvisionPPPoE: func(_ context.Context, _ domainRouterJob.Job) error {
			return errors.New("router unreachable")
		},
	}, nil).WithClock(func() time.Time { return fixedNow })

	res, err := worker.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, res.Retried)

	job, err := repo.FindByID(context.Background(), "job-retry")
	require.NoError(t, err)
	assert.Equal(t, domainRouterJob.StatusRetryWait, job.Status)
	assert.Equal(t, "router unreachable", job.LastError)
	assert.True(t, job.NextRunAt.Equal(fixedNow.Add(30*time.Second)), "attempt pertama backoff 30s")
}

func TestWorker_MarksDeadAfterMaxAttempts(t *testing.T) {
	repo := mocktest.NewFakeRouterJobRepo()
	seedJob(repo, "job-dead", domainRouterJob.TypeProvisionPPPoE, 4, 5) // attempts 4 → claim jadi 5
	worker := uc.NewWorker(repo, map[domainRouterJob.Type]uc.Handler{
		domainRouterJob.TypeProvisionPPPoE: func(_ context.Context, _ domainRouterJob.Job) error {
			return errors.New("still failing")
		},
	}, nil).WithClock(func() time.Time { return fixedNow })

	res, err := worker.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, res.Dead)

	job, err := repo.FindByID(context.Background(), "job-dead")
	require.NoError(t, err)
	assert.Equal(t, domainRouterJob.StatusDead, job.Status)
	assert.Equal(t, "still failing", job.LastError)
}

func TestWorker_ErrorBudgetStopsBatch(t *testing.T) {
	repo := mocktest.NewFakeRouterJobRepo()
	for i := 0; i < 8; i++ {
		seedJob(repo, "job-"+string(rune('a'+i)), domainRouterJob.TypeProvisionPPPoE, 0, 5)
	}
	settings := mocktest.NewFakeSettingReader(map[string]string{"isp.router_job_error_budget": "5"})
	worker := uc.NewWorker(repo, map[domainRouterJob.Type]uc.Handler{
		domainRouterJob.TypeProvisionPPPoE: func(_ context.Context, _ domainRouterJob.Job) error {
			return errors.New("fail")
		},
	}, settings).WithClock(func() time.Time { return fixedNow })

	res, err := worker.Run(context.Background())
	require.NoError(t, err)
	assert.True(t, res.StoppedEarly)
	assert.Equal(t, 5, res.Processed, "batch berhenti setelah 5 error beruntun")
}

func TestWorker_BatchSizeFromSettings(t *testing.T) {
	repo := mocktest.NewFakeRouterJobRepo()
	for i := 0; i < 5; i++ {
		seedJob(repo, "job-"+string(rune('a'+i)), domainRouterJob.TypeProvisionPPPoE, 0, 5)
	}
	settings := mocktest.NewFakeSettingReader(map[string]string{"isp.router_job_batch_size": "2"})
	worker := uc.NewWorker(repo, map[domainRouterJob.Type]uc.Handler{
		domainRouterJob.TypeProvisionPPPoE: func(_ context.Context, _ domainRouterJob.Job) error { return nil },
	}, settings).WithClock(func() time.Time { return fixedNow })

	res, err := worker.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 2, res.Processed)
}

func TestWorker_EnqueueProvisionIdempotentWithinMinute(t *testing.T) {
	repo := mocktest.NewFakeRouterJobRepo()
	worker := uc.NewWorker(repo, nil, nil).WithClock(func() time.Time { return fixedNow })

	first, err := worker.EnqueueProvision(context.Background(), "tenant-default", "dev-1", "sub-1", "PPPOE")
	require.NoError(t, err)
	second, err := worker.EnqueueProvision(context.Background(), "tenant-default", "dev-1", "sub-1", "PPPOE")
	require.NoError(t, err)
	assert.Equal(t, first.ID, second.ID, "klik ganda dalam menit sama tidak menggandakan job")
	assert.Equal(t, domainRouterJob.TypeProvisionPPPoE, first.JobType)
}

func TestBackoff(t *testing.T) {
	assert.Equal(t, 30*time.Second, domainRouterJob.Backoff(1))
	assert.Equal(t, time.Minute, domainRouterJob.Backoff(2))
	assert.Equal(t, 2*time.Minute, domainRouterJob.Backoff(3))
	assert.Equal(t, 30*time.Minute, domainRouterJob.Backoff(10))
}
