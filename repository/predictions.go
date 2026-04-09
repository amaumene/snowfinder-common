package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/amaumene/snowfinder_common/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PredictionRepository provides access to prediction-related tables.
type PredictionRepository struct {
	db      *pgxpool.Pool
	beginTx func(context.Context) (predictionTx, error)
}

type predictionTx interface {
	SendBatch(context.Context, *pgx.Batch) pgx.BatchResults
	Commit(context.Context) error
	Rollback(context.Context) error
}

// NewPredictionRepository creates a new prediction repository.
func NewPredictionRepository(db *pgxpool.Pool) *PredictionRepository {
	return &PredictionRepository{
		db: db,
		beginTx: func(ctx context.Context) (predictionTx, error) {
			return db.Begin(ctx)
		},
	}
}

// LoadPredictionConfig loads per-resort config from prediction_config table.
func (r *PredictionRepository) LoadPredictionConfig(ctx context.Context) (map[string]models.PredictorResortConfig, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	rows, err := r.db.Query(ctx, "SELECT resort_id, config_data FROM prediction_config")
	if err != nil {
		return nil, fmt.Errorf("query prediction_config: %w", err)
	}
	defer rows.Close()

	resorts := make(map[string]models.PredictorResortConfig)
	for rows.Next() {
		var resortID string
		var configData []byte
		if err := rows.Scan(&resortID, &configData); err != nil {
			return nil, fmt.Errorf("scan prediction_config: %w", err)
		}
		var cfg models.PredictorResortConfig
		if err := json.Unmarshal(configData, &cfg); err != nil {
			return nil, fmt.Errorf("unmarshal config for %s: %w", resortID, err)
		}
		resorts[resortID] = cfg
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate prediction_config rows: %w", err)
	}

	return resorts, nil
}

// LoadGlobalParams loads global predictor parameters.
func (r *PredictionRepository) LoadGlobalParams(ctx context.Context) (models.GlobalParams, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var paramsData []byte
	err := r.db.QueryRow(ctx,
		"SELECT params_data FROM prediction_global_params WHERE id = 1",
	).Scan(&paramsData)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.GlobalParams{}, nil
		}
		return models.GlobalParams{}, fmt.Errorf("query global_params: %w", err)
	}
	var params models.GlobalParams
	if err := json.Unmarshal(paramsData, &params); err != nil {
		return models.GlobalParams{}, fmt.Errorf("unmarshal global_params: %w", err)
	}
	return params, nil
}

// SavePredictions upserts all predictions using INSERT ON CONFLICT.
func (r *PredictionRepository) SavePredictions(ctx context.Context, predictions *models.PredictionData) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	batch := &pgx.Batch{}
	count, err := queuePredictionBatch(batch, predictions)
	if err != nil {
		return fmt.Errorf("prepare predictions batch: %w", err)
	}
	if count == 0 {
		return nil
	}
	if r.beginTx == nil {
		return fmt.Errorf("begin prediction transaction: repository not initialized")
	}

	tx, err := r.beginTx(ctx)
	if err != nil {
		return fmt.Errorf("begin prediction transaction: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(context.Background())
		}
	}()

	if err := executeBatchResults(tx.SendBatch(ctx, batch), count, "upsert prediction"); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit predictions: %w", err)
	}
	committed = true

	return nil
}

func queuePredictionBatch(batch *pgx.Batch, predictions *models.PredictionData) (int, error) {
	if predictions == nil {
		return 0, errors.New("nil prediction data")
	}

	count := 0
	for resortID, pred := range predictions.Resorts {
		predJSON, err := json.Marshal(pred)
		if err != nil {
			return 0, fmt.Errorf("marshal prediction for %s: %w", resortID, err)
		}
		batch.Queue(
			`INSERT INTO predictions (resort_id, prediction_data, generated_at)
			 VALUES ($1, $2, $3)
			 ON CONFLICT (resort_id) DO UPDATE
			 SET prediction_data = EXCLUDED.prediction_data,
			     generated_at = EXCLUDED.generated_at`,
			resortID, predJSON, predictions.GeneratedAt,
		)
		count++
	}

	return count, nil
}
