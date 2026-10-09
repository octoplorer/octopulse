package engine

import (
	"context"
	"errors"
	"testing"
)

func TestUncertainCommitReconcilesWithoutRepeatingWrite(t *testing.T) {
	calls := 0
	err := reconcileCommit(
		context.Background(),
		func() error {
			calls++
			return errors.New("connection lost after commit")
		},
		func() (bool, error) { return true, nil },
	)
	if err != nil || calls != 1 {
		t.Fatalf("uncertain committed round: %v, calls %d", err, calls)
	}
}
func TestPersistenceRetryPreservesObservationAndIsBounded(t *testing.T) {
	calls := 0
	err := reconcileCommit(context.Background(), func() error {
		calls++
		if calls < 3 {
			return errors.New("transient persistence error")
		}
		return nil
	}, func() (bool, error) { return false, nil })
	if err != nil || calls != 3 {
		t.Fatalf("retry: %v, calls %d", err, calls)
	}
	calls = 0
	err = reconcileCommit(
		context.Background(),
		func() error {
			calls++
			return ErrSuperseded
		},
		func() (bool, error) { return false, nil },
	)
	if !errors.Is(err, ErrSuperseded) || calls != 1 {
		t.Fatal("obsolete observation retried")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls = 0
	err = reconcileCommit(
		ctx,
		func() error {
			calls++
			return errors.New("offline")
		},
		func() (bool, error) { return false, nil },
	)
	if !errors.Is(err, context.Canceled) || calls > 1 {
		t.Fatal("cancelled commit kept retrying")
	}
}
