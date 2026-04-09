package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/amaumene/snowfinder_common/models"
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
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if resort.ID == "" {
		resort.ID = uuid.New().String()
	}

	persistedRecord, err := r.resolveResortRecord(ctx, resort)
	if err != nil {
		return fmt.Errorf("resolve resort identity: %w", err)
	}

	if persistedRecord.ID != "" {
		resort.ID = persistedRecord.ID
	}
	resort.Slug = persistedRecord.Slug

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

	err = r.ReaderRepository.db.QueryRow(ctx, query,
		resort.ID, resort.Slug, resort.Name, resort.Prefecture, resort.Region,
		resort.TopElevationM, resort.BaseElevationM, resort.VerticalM,
		resort.NumCourses, resort.LongestCourseKM, resort.SteepestCourseDeg,
	).Scan(&resort.ID)

	if err != nil {
		return fmt.Errorf("save resort: %w", err)
	}

	return nil
}

func (r *WriterRepository) resolveResortRecord(ctx context.Context, resort *models.Resort) (*resortIdentityRecord, error) {
	existingBySlug, err := r.getResortIdentityRecordBySlug(ctx, resort.Slug)
	if err != nil {
		return nil, err
	}

	scopedSlug := scopedResortSlug(resort.Slug, resort.Prefecture, resort.Region)
	var existingByScopedSlug *resortIdentityRecord
	if scopedSlug != resort.Slug {
		existingByScopedSlug, err = r.getResortIdentityRecordBySlug(ctx, scopedSlug)
		if err != nil {
			return nil, err
		}
	}

	return resolvePersistedResortRecord(resort, existingBySlug, existingByScopedSlug), nil
}

func (r *WriterRepository) getResortIdentityRecordBySlug(ctx context.Context, slug string) (*resortIdentityRecord, error) {
	query := `
		SELECT id, slug, prefecture, region
		FROM resorts
		WHERE slug = $1
	`

	var record resortIdentityRecord
	err := r.ReaderRepository.db.QueryRow(ctx, query, slug).Scan(
		&record.ID,
		&record.Slug,
		&record.Prefecture,
		&record.Region,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("get resort by slug %q: %w", slug, err)
	}

	return &record, nil
}

func (r *WriterRepository) SaveSnowDepthReadings(ctx context.Context, readings []models.SnowDepthReading) error {
	if len(readings) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	query := `
		INSERT INTO snow_depth_readings (resort_id, date, depth_cm)
		VALUES ($1, $2, $3)
		ON CONFLICT (resort_id, date) DO UPDATE SET
			depth_cm = EXCLUDED.depth_cm
	`

	for start := 0; start < len(readings); start += batchChunkSize {
		end := start + batchChunkSize
		if end > len(readings) {
			end = len(readings)
		}
		chunk := readings[start:end]

		batch := &pgx.Batch{}
		for _, reading := range chunk {
			batch.Queue(query, reading.ResortID, reading.Date, reading.DepthCM)
		}
		if err := executeBatchResults(r.ReaderRepository.db.SendBatch(ctx, batch), len(chunk), "save reading"); err != nil {
			return err
		}
	}

	return nil
}

func (r *WriterRepository) SaveFailedScrapeAttempt(ctx context.Context, resortURL, errorMessage string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

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
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	query := `
		UPDATE failed_scrape_attempts
		SET retried = TRUE, retried_at = NOW()
		WHERE id = $1
	`

	tag, err := r.ReaderRepository.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("mark failed attempt retried: %w", err)
	}
	if err := requireRowsAffected(tag, 1, "mark failed attempt retried"); err != nil {
		return err
	}

	return nil
}

func (r *WriterRepository) SaveDailySnowfall(ctx context.Context, snowfalls []models.DailySnowfall) error {
	if len(snowfalls) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	query := `
		INSERT INTO daily_snowfall (resort_id, date, snowfall_cm)
		VALUES ($1, $2, $3)
		ON CONFLICT (resort_id, date) DO UPDATE SET
			snowfall_cm = EXCLUDED.snowfall_cm
	`

	for start := 0; start < len(snowfalls); start += batchChunkSize {
		end := start + batchChunkSize
		if end > len(snowfalls) {
			end = len(snowfalls)
		}
		chunk := snowfalls[start:end]

		batch := &pgx.Batch{}
		for _, sf := range chunk {
			batch.Queue(query, sf.ResortID, sf.Date, sf.SnowfallCM)
		}
		if err := executeBatchResults(r.ReaderRepository.db.SendBatch(ctx, batch), len(chunk), "save snowfall"); err != nil {
			return err
		}
	}

	return nil
}
