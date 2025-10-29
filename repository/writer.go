package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/amaumene/snowfinder-common/models"
)

// WriterRepository provides full read-write database access.
type WriterRepository struct {
	*ReaderRepository
	db *pgxpool.Pool
}

// NewWriter creates a new read-write repository.
func NewWriter(db *pgxpool.Pool) *WriterRepository {
	return &WriterRepository{
		ReaderRepository: NewReader(db),
		db:               db,
	}
}

func (r *WriterRepository) SaveResort(ctx context.Context, resort *models.Resort) error {
	if resort.ID == "" {
		resort.ID = uuid.New().String()
	}

	query := `
		INSERT INTO resorts (
			id, slug, name, prefecture, region,
			top_elevation_m, base_elevation_m, vertical_m,
			num_courses, longest_course_km, steepest_course_deg
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (slug) DO UPDATE SET
			name = EXCLUDED.name,
			prefecture = EXCLUDED.prefecture,
			region = EXCLUDED.region,
			top_elevation_m = EXCLUDED.top_elevation_m,
			base_elevation_m = EXCLUDED.base_elevation_m,
			vertical_m = EXCLUDED.vertical_m,
			num_courses = EXCLUDED.num_courses,
			longest_course_km = EXCLUDED.longest_course_km,
			steepest_course_deg = EXCLUDED.steepest_course_deg,
			last_updated = NOW()
		RETURNING id
	`

	err := r.db.QueryRow(ctx, query,
		resort.ID, resort.Slug, resort.Name, resort.Prefecture, resort.Region,
		resort.TopElevationM, resort.BaseElevationM, resort.VerticalM,
		resort.NumCourses, resort.LongestCourseKM, resort.SteepestCourseDeg,
	).Scan(&resort.ID)

	if err != nil {
		return fmt.Errorf("save resort: %w", err)
	}

	return nil
}

func (r *WriterRepository) SaveSnowDepthReadings(ctx context.Context, readings []models.SnowDepthReading) error {
	if len(readings) == 0 {
		return nil
	}

	query := `
		INSERT INTO snow_depth_readings (resort_id, date, depth_cm, season)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (resort_id, date) DO UPDATE SET
			depth_cm = EXCLUDED.depth_cm,
			season = EXCLUDED.season
	`

	batch := &pgx.Batch{}
	for _, reading := range readings {
		batch.Queue(query, reading.ResortID, reading.Date, reading.DepthCM, reading.Season)
	}

	results := r.db.SendBatch(ctx, batch)
	defer results.Close()

	for range readings {
		if _, err := results.Exec(); err != nil {
			return fmt.Errorf("save reading: %w", err)
		}
	}

	return nil
}

func (r *WriterRepository) SaveDailySnowfall(ctx context.Context, snowfalls []models.DailySnowfall) error {
	if len(snowfalls) == 0 {
		return nil
	}

	query := `
		INSERT INTO daily_snowfall (resort_id, date, snowfall_cm, season)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (resort_id, date) DO UPDATE SET
			snowfall_cm = EXCLUDED.snowfall_cm,
			season = EXCLUDED.season
	`

	batch := &pgx.Batch{}
	for _, sf := range snowfalls {
		batch.Queue(query, sf.ResortID, sf.Date, sf.SnowfallCM, sf.Season)
	}

	results := r.db.SendBatch(ctx, batch)
	defer results.Close()

	for range snowfalls {
		if _, err := results.Exec(); err != nil {
			return fmt.Errorf("save snowfall: %w", err)
		}
	}

	return nil
}
