package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/amaumene/snowfinder-common/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PredictionRepository provides access to prediction-related tables.
type PredictionRepository struct {
	db *pgxpool.Pool
}

// NewPredictionRepository creates a new prediction repository.
func NewPredictionRepository(db *pgxpool.Pool) *PredictionRepository {
	return &PredictionRepository{db: db}
}

// LoadPredictionConfig loads per-resort config from prediction_config table.
func (r *PredictionRepository) LoadPredictionConfig(ctx context.Context) (map[string]models.PredictorResortConfig, error) {
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
	return resorts, rows.Err()
}

// LoadGlobalParams loads global predictor parameters.
func (r *PredictionRepository) LoadGlobalParams(ctx context.Context) (models.GlobalParams, error) {
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

// SavePredictions atomically replaces all predictions (DELETE + INSERT in a transaction).
func (r *PredictionRepository) SavePredictions(ctx context.Context, predictions *models.PredictionData) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "DELETE FROM predictions"); err != nil {
		return fmt.Errorf("delete predictions: %w", err)
	}

	batch := &pgx.Batch{}
	for resortID, pred := range predictions.Resorts {
		predJSON, err := json.Marshal(pred)
		if err != nil {
			return fmt.Errorf("marshal prediction for %s: %w", resortID, err)
		}
		batch.Queue(
			"INSERT INTO predictions (resort_id, prediction_data, generated_at) VALUES ($1, $2, $3)",
			resortID, predJSON, predictions.GeneratedAt,
		)
	}

	results := tx.SendBatch(ctx, batch)
	for range predictions.Resorts {
		if _, err := results.Exec(); err != nil {
			results.Close()
			return fmt.Errorf("insert prediction: %w", err)
		}
	}
	results.Close()

	return tx.Commit(ctx)
}
