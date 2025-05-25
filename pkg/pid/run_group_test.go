package pid_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
	"github.com/zhulik/pal/pkg/pid"
)

var (
	errTest = errors.New("test error")
)

func job(d time.Duration, err error) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		timer := time.NewTimer(d)
		defer timer.Stop()

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timer.C:
				return err
			}
		}
	}
}

func TestRunGroup_Success(t *testing.T) {
	t.Parallel()

	t.Run("when no jobs started", func(t *testing.T) {
		t.Parallel()

		rg := pid.NewRunGroup()

		err := rg.Wait()

		require.NoError(t, err)
	})

	t.Run("when some jobs complete before the others", func(t *testing.T) {
		t.Parallel()

		rg := pid.NewRunGroup()
		rg.Go(t.Context(), true, job(100*time.Millisecond, nil))
		rg.Go(t.Context(), true, job(100*time.Millisecond, nil))
		rg.Go(t.Context(), true, job(200*time.Millisecond, nil))

		start := time.Now()

		err := rg.Wait()

		require.NoError(t, err)
		assert.InDelta(t, 205.0, time.Since(start).Milliseconds(), 5.0)
	})

	t.Run("when all jobs complete before wait is called", func(t *testing.T) {
		t.Parallel()

		rg := pid.NewRunGroup()
		rg.Go(t.Context(), true, job(100*time.Millisecond, nil))
		rg.Go(t.Context(), true, job(100*time.Millisecond, nil))
		rg.Go(t.Context(), true, job(100*time.Millisecond, nil))

		time.Sleep(200 * time.Millisecond)

		start := time.Now()

		go rg.Wait() // nolint:errcheck
		err := rg.Wait()

		require.NoError(t, err)
		require.Less(t, time.Since(start), 10*time.Millisecond)
	})

	t.Run("when all jobs aren non-wait jobs", func(t *testing.T) {
		t.Parallel()

		rg := pid.NewRunGroup()
		rg.Go(t.Context(), false, job(100*time.Millisecond, nil))
		rg.Go(t.Context(), false, job(100*time.Millisecond, nil))
		rg.Go(t.Context(), false, job(100*time.Millisecond, nil))

		start := time.Now()

		err := rg.Wait()

		require.ErrorIs(t, err, context.Canceled)
		require.Less(t, time.Since(start), 10*time.Millisecond)
	})

	t.Run("when all wait jobs complete", func(t *testing.T) {
		t.Parallel()

		rg := pid.NewRunGroup()
		rg.Go(t.Context(), false, job(500*time.Millisecond, nil))
		rg.Go(t.Context(), true, job(100*time.Millisecond, nil))
		rg.Go(t.Context(), true, job(100*time.Millisecond, nil))
		rg.Go(t.Context(), true, job(100*time.Millisecond, nil))

		start := time.Now()

		err := rg.Wait()

		require.ErrorIs(t, err, context.Canceled)
		assert.InDelta(t, 105.0, time.Since(start).Milliseconds(), 5.0)
	})

	t.Run("when context passed to Go is canceled", func(t *testing.T) {
		t.Parallel()

		rg := pid.NewRunGroup()

		ctx, cancel := context.WithCancel(t.Context())

		rg.Go(ctx, true, job(100*time.Millisecond, nil))
		rg.Go(ctx, true, job(100*time.Millisecond, nil))
		rg.Go(ctx, true, job(100*time.Millisecond, nil))

		cancel()

		start := time.Now()

		err := rg.Wait()

		require.ErrorIs(t, err, context.Canceled)
		require.Less(t, time.Since(start), 10*time.Millisecond)
	})
}

func TestRunGroup_Error(t *testing.T) {
	t.Parallel()

	t.Run("when a job returns an error", func(t *testing.T) {
		t.Parallel()

		rg := pid.NewRunGroup()
		rg.Go(t.Context(), true, job(100*time.Millisecond, errTest))
		rg.Go(t.Context(), true, job(20000*time.Millisecond, nil))

		start := time.Now()

		err := rg.Wait()

		require.ErrorIs(t, err, errTest)
		assert.InDelta(t, 105.0, time.Since(start).Milliseconds(), 5.0)
	})
}
