package repository

import (
	"context"
	"errors"
	"math"
	"testing"

	"github.com/amaumene/snowfinder_common/models"
	"github.com/jackc/pgx/v5"
)

type fakePredictionTx struct {
	batchResults pgx.BatchResults
	committed    bool
	rolledBack   bool
	commitErr    error
}

func (f *fakePredictionTx) SendBatch(_ context.Context, _ *pgx.Batch) pgx.BatchResults {
	return f.batchResults
}

func (f *fakePredictionTx) Commit(_ context.Context) error {
	f.committed = true
	return f.commitErr
}

func (f *fakePredictionTx) Rollback(_ context.Context) error {
	f.rolledBack = true
	return nil
}

func TestPredictionRepositorySavePredictions_CommitsTransaction(t *testing.T) {
	t.Parallel()

	results := &fakeBatchResults{}
	tx := &fakePredictionTx{batchResults: results}
	repo := &PredictionRepository{
		beginTx: func(context.Context) (predictionTx, error) {
			return tx, nil
		},
	}

	predictions := &models.PredictionData{
		GeneratedAt: "2026-04-09T12:00:00Z",
		Resorts: map[string]models.Prediction{
			"resort-1": {Name: "One"},
			"resort-2": {Name: "Two"},
		},
	}

	if err := repo.SavePredictions(context.Background(), predictions); err != nil {
		t.Fatalf("SavePredictions() error = %v", err)
	}
	if !tx.committed {
		t.Fatal("expected transaction commit")
	}
	if tx.rolledBack {
		t.Fatal("did not expect rollback on success")
	}
	if results.execCalls != len(predictions.Resorts) {
		t.Fatalf("executed %d statements, want %d", results.execCalls, len(predictions.Resorts))
	}
}

func TestPredictionRepositorySavePredictions_RollsBackOnBatchError(t *testing.T) {
	t.Parallel()

	execErr := errors.New("batch failed")
	results := &fakeBatchResults{
		failOnExec: 2,
		execErr:    execErr,
	}
	tx := &fakePredictionTx{batchResults: results}
	repo := &PredictionRepository{
		beginTx: func(context.Context) (predictionTx, error) {
			return tx, nil
		},
	}

	predictions := &models.PredictionData{
		GeneratedAt: "2026-04-09T12:00:00Z",
		Resorts: map[string]models.Prediction{
			"resort-1": {Name: "One"},
			"resort-2": {Name: "Two"},
		},
	}

	err := repo.SavePredictions(context.Background(), predictions)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, execErr) {
		t.Fatalf("expected wrapped exec error, got %v", err)
	}
	if tx.committed {
		t.Fatal("did not expect commit after batch failure")
	}
	if !tx.rolledBack {
		t.Fatal("expected rollback after batch failure")
	}
}

func TestPredictionRepositorySavePredictions_RejectsInvalidJSONBeforeStartingTransaction(t *testing.T) {
	t.Parallel()

	beginCalled := false
	repo := &PredictionRepository{
		beginTx: func(context.Context) (predictionTx, error) {
			beginCalled = true
			return nil, nil
		},
	}

	predictions := &models.PredictionData{
		GeneratedAt: "2026-04-09T12:00:00Z",
		Resorts: map[string]models.Prediction{
			"resort-1": {
				HourlySnowfall: []float64{math.NaN()},
			},
		},
	}

	err := repo.SavePredictions(context.Background(), predictions)
	if err == nil {
		t.Fatal("expected error")
	}
	if beginCalled {
		t.Fatal("did not expect transaction to start when marshaling fails")
	}
}
