package repository

import (
	"context"
	"fmt"

	"github.com/amaumene/snowfinder-common/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ReaderRepository provides read-only database access.
type ReaderRepository struct {
	db *pgxpool.Pool
}

// NewReader creates a new read-only repository.
func NewReader(db *pgxpool.Pool) *ReaderRepository {
	return &ReaderRepository{db: db}
}

func (r *ReaderRepository) GetResortBySlug(ctx context.Context, slug string) (*models.Resort, error) {
	query := `
		SELECT id, slug, name, prefecture, region,
			   top_elevation_m, base_elevation_m, vertical_m,
			   num_courses, longest_course_km, steepest_course_deg,
			   last_updated
		FROM resorts
		WHERE slug = $1
	`

	var resort models.Resort
	err := r.db.QueryRow(ctx, query, slug).Scan(
		&resort.ID, &resort.Slug, &resort.Name, &resort.Prefecture, &resort.Region,
		&resort.TopElevationM, &resort.BaseElevationM, &resort.VerticalM,
		&resort.NumCourses, &resort.LongestCourseKM, &resort.SteepestCourseDeg,
		&resort.LastUpdated,
	)

	if err != nil {
		return nil, fmt.Errorf("get resort: %w", err)
	}

	return &resort, nil
}

func (r *ReaderRepository) GetResortByID(ctx context.Context, id string) (*models.Resort, error) {
	query := `
		SELECT id, slug, name, prefecture, region,
			   top_elevation_m, base_elevation_m, vertical_m,
			   num_courses, longest_course_km, steepest_course_deg,
			   last_updated
		FROM resorts
		WHERE id = $1
	`

	var resort models.Resort
	err := r.db.QueryRow(ctx, query, id).Scan(
		&resort.ID, &resort.Slug, &resort.Name, &resort.Prefecture, &resort.Region,
		&resort.TopElevationM, &resort.BaseElevationM, &resort.VerticalM,
		&resort.NumCourses, &resort.LongestCourseKM, &resort.SteepestCourseDeg,
		&resort.LastUpdated,
	)

	if err != nil {
		return nil, fmt.Errorf("get resort by id: %w", err)
	}

	return &resort, nil
}

// GetSnowiestResorts queries snowiest resorts for a date range with optional prefecture filter.
// If endDate is empty, it defaults to startDate + 6 days (week mode).
func (r *ReaderRepository) GetSnowiestResorts(ctx context.Context, startDate, endDate, prefecture string, limit int) ([]models.WeeklyResortStats, error) {
	args := []any{startDate}

	var cteFilter string
	if endDate == "" {
		// Week mode: use mmdd arithmetic on a single start date
		cteFilter = `mmdd(date) >= mmdd($1::date)
			  AND mmdd(date) <= mmdd($1::date + INTERVAL '6 days')`
	} else {
		// Date range mode: handle wrap-around (e.g. Dec–Jan)
		args = append(args, endDate)
		cteFilter = `CASE
					WHEN $1 <= $2 THEN
						mmdd(date) >= $1 AND mmdd(date) <= $2
					ELSE
						mmdd(date) >= $1 OR mmdd(date) <= $2
				END`
	}

	// Optional prefecture filter — appended as the next positional parameter
	prefectureClause := ""
	if prefecture != "" {
		args = append(args, prefecture)
		prefectureClause = fmt.Sprintf("AND r.prefecture = $%d", len(args))
	}

	// LIMIT is always the last parameter
	args = append(args, limit)
	limitParam := fmt.Sprintf("$%d", len(args))

	query := fmt.Sprintf(`
		WITH range_data AS (
			SELECT
				resort_id,
				EXTRACT(YEAR FROM date) as year,
				SUM(snowfall_cm) as total_snowfall
			FROM daily_snowfall
			WHERE %s
			GROUP BY resort_id, year
		),
		avg_range_data AS (
			SELECT
				resort_id,
				AVG(total_snowfall) as avg_snowfall,
				COUNT(*) as years_with_data
			FROM range_data
			GROUP BY resort_id
		)
		SELECT
			r.id,
			r.name,
			r.prefecture,
			ROUND(ard.avg_snowfall)::int as avg_snowfall,
			ard.years_with_data,
			r.top_elevation_m,
			r.base_elevation_m,
			r.vertical_m,
			r.num_courses,
			r.longest_course_km
		FROM avg_range_data ard
		JOIN resorts r ON r.id = ard.resort_id
		WHERE ard.years_with_data >= 1
		%s
		ORDER BY ard.avg_snowfall DESC
		LIMIT %s
	`, cteFilter, prefectureClause, limitParam)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query snowiest resorts: %w", err)
	}
	defer rows.Close()

	return scanWeeklyResortStats(rows)
}

// scanWeeklyResortStats scans pgx rows into a slice of WeeklyResortStats.
func scanWeeklyResortStats(rows pgx.Rows) ([]models.WeeklyResortStats, error) {
	results := []models.WeeklyResortStats{}
	for rows.Next() {
		var stat models.WeeklyResortStats
		if err := rows.Scan(&stat.ResortID, &stat.Name, &stat.Prefecture, &stat.TotalSnowfall, &stat.YearsWithData,
			&stat.TopElevationM, &stat.BaseElevationM, &stat.VerticalM, &stat.NumCourses, &stat.LongestCourseKM); err != nil {
			return nil, fmt.Errorf("scan result: %w", err)
		}
		results = append(results, stat)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}
	return results, nil
}

func (r *ReaderRepository) GetAllResortsWithPeaks(ctx context.Context) ([]models.ResortWithPeaks, error) {
	// Single JOIN query to fetch resorts and their peaks together
	query := `
		SELECT r.id, r.slug, r.name, r.prefecture, r.region,
			   r.top_elevation_m, r.base_elevation_m, r.vertical_m,
			   r.num_courses, r.longest_course_km, r.steepest_course_deg,
			   r.last_updated,
			   p.id, p.peak_rank, p.start_date, p.end_date, p.center_date,
			   p.avg_daily_snowfall, p.total_period_snowfall, p.prominence_score,
			   p.years_of_data, p.confidence_level, p.calculated_at
		FROM resorts r
		INNER JOIN resort_peak_periods p ON r.id = p.resort_id
		ORDER BY r.prefecture, r.name, p.peak_rank
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query resorts with peaks: %w", err)
	}
	defer rows.Close()

	resortMap := make(map[string]*models.ResortWithPeaks)
	var order []string

	for rows.Next() {
		var resort models.Resort
		var peak models.PeakPeriod
		if err := rows.Scan(
			&resort.ID, &resort.Slug, &resort.Name, &resort.Prefecture, &resort.Region,
			&resort.TopElevationM, &resort.BaseElevationM, &resort.VerticalM,
			&resort.NumCourses, &resort.LongestCourseKM, &resort.SteepestCourseDeg,
			&resort.LastUpdated,
			&peak.ID, &peak.PeakRank, &peak.StartDate, &peak.EndDate, &peak.CenterDate,
			&peak.AvgDailySnowfall, &peak.TotalPeriodSnowfall, &peak.ProminenceScore,
			&peak.YearsOfData, &peak.ConfidenceLevel, &peak.CalculatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan resort with peak: %w", err)
		}

		peak.ResortID = resort.ID

		if _, exists := resortMap[resort.ID]; !exists {
			resortMap[resort.ID] = &models.ResortWithPeaks{
				Resort: resort,
				Peaks:  []models.PeakPeriod{},
			}
			order = append(order, resort.ID)
		}
		resortMap[resort.ID].Peaks = append(resortMap[resort.ID].Peaks, peak)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	results := make([]models.ResortWithPeaks, 0, len(order))
	for _, id := range order {
		results = append(results, *resortMap[id])
	}

	return results, nil
}

func (r *ReaderRepository) GetPendingFailedScrapeAttempts(ctx context.Context) ([]models.FailedScrapeAttempt, error) {
	query := `
		SELECT id, resort_url, error_message, failed_at, retried, retried_at
		FROM failed_scrape_attempts
		WHERE retried = FALSE
		ORDER BY failed_at ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query failed scrape attempts: %w", err)
	}
	defer rows.Close()

	var attempts []models.FailedScrapeAttempt
	for rows.Next() {
		var a models.FailedScrapeAttempt
		if err := rows.Scan(&a.ID, &a.ResortURL, &a.ErrorMessage, &a.FailedAt, &a.Retried, &a.RetriedAt); err != nil {
			return nil, fmt.Errorf("scan failed scrape attempt: %w", err)
		}
		attempts = append(attempts, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	return attempts, nil
}

func (r *ReaderRepository) GetPeakPeriodsForResort(ctx context.Context, resortID string) ([]models.PeakPeriod, error) {
	query := `
		SELECT id, resort_id, peak_rank, start_date, end_date, center_date,
			   avg_daily_snowfall, total_period_snowfall, prominence_score,
			   years_of_data, confidence_level, calculated_at
		FROM resort_peak_periods
		WHERE resort_id = $1
		ORDER BY peak_rank
	`

	rows, err := r.db.Query(ctx, query, resortID)
	if err != nil {
		return nil, fmt.Errorf("query peak periods: %w", err)
	}
	defer rows.Close()

	var peaks []models.PeakPeriod
	for rows.Next() {
		var peak models.PeakPeriod
		if err := rows.Scan(
			&peak.ID, &peak.ResortID, &peak.PeakRank, &peak.StartDate, &peak.EndDate, &peak.CenterDate,
			&peak.AvgDailySnowfall, &peak.TotalPeriodSnowfall, &peak.ProminenceScore,
			&peak.YearsOfData, &peak.ConfidenceLevel, &peak.CalculatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan peak period: %w", err)
		}
		peaks = append(peaks, peak)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	return peaks, nil
}
