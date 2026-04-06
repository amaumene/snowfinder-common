package repository

import (
	"context"
	"fmt"

	"github.com/amaumene/snowfinder-common/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const batchChunkSize = 500

// WriterRepository provides full read-write database access.
type WriterRepository struct {
	*ReaderRepository
}

// NewWriter creates a new read-write repository.
func NewWriter(db *pgxpool.Pool) *WriterRepository {
	return &WriterRepository{
		ReaderRepository: NewReader(db),
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

	err := r.ReaderRepository.db.QueryRow(ctx, query,
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

	for start := 0; start < len(readings); start += batchChunkSize {
		end := start + batchChunkSize
		if end > len(readings) {
			end = len(readings)
		}
		chunk := readings[start:end]

		batch := &pgx.Batch{}
		for _, reading := range chunk {
			batch.Queue(query, reading.ResortID, reading.Date, reading.DepthCM, reading.Season)
		}
		results := r.ReaderRepository.db.SendBatch(ctx, batch)
		for range chunk {
			if _, err := results.Exec(); err != nil {
				results.Close()
				return fmt.Errorf("save reading: %w", err)
			}
		}
		results.Close()
	}

	return nil
}

func (r *WriterRepository) SaveFailedScrapeAttempt(ctx context.Context, resortURL, errorMessage string) error {
	query := `
		INSERT INTO failed_scrape_attempts (id, resort_url, error_message, failed_at, retried)
		VALUES (gen_random_uuid(), $1, $2, NOW(), FALSE)
	`

	if _, err := r.ReaderRepository.db.Exec(ctx, query, resortURL, errorMessage); err != nil {
		return fmt.Errorf("save failed scrape attempt: %w", err)
	}

	return nil
}

func (r *WriterRepository) MarkFailedAttemptRetried(ctx context.Context, id string) error {
	query := `
		UPDATE failed_scrape_attempts
		SET retried = TRUE, retried_at = NOW()
		WHERE id = $1
	`

	if _, err := r.ReaderRepository.db.Exec(ctx, query, id); err != nil {
		return fmt.Errorf("mark failed attempt retried: %w", err)
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

	for start := 0; start < len(snowfalls); start += batchChunkSize {
		end := start + batchChunkSize
		if end > len(snowfalls) {
			end = len(snowfalls)
		}
		chunk := snowfalls[start:end]

		batch := &pgx.Batch{}
		for _, sf := range chunk {
			batch.Queue(query, sf.ResortID, sf.Date, sf.SnowfallCM, sf.Season)
		}
		results := r.ReaderRepository.db.SendBatch(ctx, batch)
		for range chunk {
			if _, err := results.Exec(); err != nil {
				results.Close()
				return fmt.Errorf("save snowfall: %w", err)
			}
		}
		results.Close()
	}

	return nil
}
