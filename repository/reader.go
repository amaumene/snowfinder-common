package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/amaumene/snowfinder-common/models"
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

func (r *ReaderRepository) GetSnowiestResortsForWeek(ctx context.Context, weekStart string, limit int) ([]models.WeeklyResortStats, error) {
	query := `
		WITH weekly_data AS (
			SELECT
				resort_id,
				EXTRACT(YEAR FROM date) as year,
				SUM(snowfall_cm) as total_snowfall
			FROM daily_snowfall
			WHERE EXTRACT(MONTH FROM date) = EXTRACT(MONTH FROM $1::date)
			  AND EXTRACT(DAY FROM date) BETWEEN EXTRACT(DAY FROM $1::date)
			    AND EXTRACT(DAY FROM $1::date) + 6
			GROUP BY resort_id, year
		),
		avg_weekly_data AS (
			SELECT
				resort_id,
				AVG(total_snowfall) as avg_snowfall,
				COUNT(*) as years_with_data
			FROM weekly_data
			GROUP BY resort_id
		)
		SELECT
			r.id,
			r.name,
			r.prefecture,
			ROUND(awd.avg_snowfall)::int as avg_snowfall
		FROM avg_weekly_data awd
		JOIN resorts r ON r.id = awd.resort_id
		WHERE awd.years_with_data >= 1
		ORDER BY awd.avg_snowfall DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, weekStart, limit)
	if err != nil {
		return nil, fmt.Errorf("query snowiest resorts: %w", err)
	}
	defer rows.Close()

	var results []models.WeeklyResortStats
	for rows.Next() {
		var stat models.WeeklyResortStats
		if err := rows.Scan(&stat.ResortID, &stat.Name, &stat.Prefecture, &stat.TotalSnowfall,
			&stat.TopElevationM, &stat.BaseElevationM, &stat.VerticalM, &stat.NumCourses, &stat.LongestCourseKM); err != nil {
			return nil, fmt.Errorf("scan result: %w", err)
		}
		results = append(results, stat)
	}

	return results, nil
}

func (r *ReaderRepository) GetSnowiestResortsForWeekByPrefecture(ctx context.Context, weekStart, prefecture string, limit int) ([]models.WeeklyResortStats, error) {
	query := `
		WITH weekly_data AS (
			SELECT
				resort_id,
				EXTRACT(YEAR FROM date) as year,
				SUM(snowfall_cm) as total_snowfall
			FROM daily_snowfall
			WHERE EXTRACT(MONTH FROM date) = EXTRACT(MONTH FROM $1::date)
			  AND EXTRACT(DAY FROM date) BETWEEN EXTRACT(DAY FROM $1::date)
			    AND EXTRACT(DAY FROM $1::date) + 6
			GROUP BY resort_id, year
		),
		avg_weekly_data AS (
			SELECT
				resort_id,
				AVG(total_snowfall) as avg_snowfall,
				COUNT(*) as years_with_data
			FROM weekly_data
			GROUP BY resort_id
		)
		SELECT
			r.id,
			r.name,
			r.prefecture,
			ROUND(awd.avg_snowfall)::int as avg_snowfall
		FROM avg_weekly_data awd
		JOIN resorts r ON r.id = awd.resort_id
		WHERE awd.years_with_data >= 1
		  AND r.prefecture = $2
		ORDER BY awd.avg_snowfall DESC
		LIMIT $3
	`

	rows, err := r.db.Query(ctx, query, weekStart, prefecture, limit)
	if err != nil {
		return nil, fmt.Errorf("query snowiest resorts by prefecture: %w", err)
	}
	defer rows.Close()

	var results []models.WeeklyResortStats
	for rows.Next() {
		var stat models.WeeklyResortStats
		if err := rows.Scan(&stat.ResortID, &stat.Name, &stat.Prefecture, &stat.TotalSnowfall,
			&stat.TopElevationM, &stat.BaseElevationM, &stat.VerticalM, &stat.NumCourses, &stat.LongestCourseKM); err != nil {
			return nil, fmt.Errorf("scan result: %w", err)
		}
		results = append(results, stat)
	}

	return results, nil
}

func (r *ReaderRepository) GetSnowiestResortsForDateRange(ctx context.Context, startDate, endDate string, limit int) ([]models.WeeklyResortStats, error) {
	query := `
		WITH date_range_data AS (
			SELECT
				resort_id,
				EXTRACT(YEAR FROM date) as year,
				SUM(snowfall_cm) as total_snowfall
			FROM daily_snowfall
			WHERE
				CASE
					WHEN $1 <= $2 THEN
						TO_CHAR(date, 'MM-DD') >= $1 AND TO_CHAR(date, 'MM-DD') <= $2
					ELSE
						TO_CHAR(date, 'MM-DD') >= $1 OR TO_CHAR(date, 'MM-DD') <= $2
				END
			GROUP BY resort_id, year
		),
		avg_range_data AS (
			SELECT
				resort_id,
				AVG(total_snowfall) as avg_snowfall,
				COUNT(*) as years_with_data
			FROM date_range_data
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
		FROM resorts r
		LEFT JOIN avg_range_data ard ON r.id = ard.resort_id
		WHERE ard.years_with_data >= 1
		ORDER BY ard.avg_snowfall DESC NULLS LAST
		LIMIT $3
	`

	rows, err := r.db.Query(ctx, query, startDate, endDate, limit)
	if err != nil {
		return nil, fmt.Errorf("query snowiest resorts for date range: %w", err)
	}
	defer rows.Close()

	var results []models.WeeklyResortStats
	for rows.Next() {
		var stat models.WeeklyResortStats
		if err := rows.Scan(&stat.ResortID, &stat.Name, &stat.Prefecture, &stat.TotalSnowfall, &stat.YearsWithData,
			&stat.TopElevationM, &stat.BaseElevationM, &stat.VerticalM, &stat.NumCourses, &stat.LongestCourseKM); err != nil {
			return nil, fmt.Errorf("scan result: %w", err)
		}
		results = append(results, stat)
	}

	return results, nil
}

func (r *ReaderRepository) GetSnowiestResortsForDateRangeByPrefecture(ctx context.Context, startDate, endDate, prefecture string, limit int) ([]models.WeeklyResortStats, error) {
	query := `
		WITH date_range_data AS (
			SELECT
				resort_id,
				EXTRACT(YEAR FROM date) as year,
				SUM(snowfall_cm) as total_snowfall
			FROM daily_snowfall
			WHERE
				CASE
					WHEN $1 <= $2 THEN
						TO_CHAR(date, 'MM-DD') >= $1 AND TO_CHAR(date, 'MM-DD') <= $2
					ELSE
						TO_CHAR(date, 'MM-DD') >= $1 OR TO_CHAR(date, 'MM-DD') <= $2
				END
			GROUP BY resort_id, year
		),
		avg_range_data AS (
			SELECT
				resort_id,
				AVG(total_snowfall) as avg_snowfall,
				COUNT(*) as years_with_data
			FROM date_range_data
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
		FROM resorts r
		LEFT JOIN avg_range_data ard ON r.id = ard.resort_id
		WHERE r.prefecture = $3 AND ard.years_with_data >= 1
		ORDER BY ard.avg_snowfall DESC NULLS LAST
		LIMIT $4
	`

	rows, err := r.db.Query(ctx, query, startDate, endDate, prefecture, limit)
	if err != nil {
		return nil, fmt.Errorf("query snowiest resorts for date range by prefecture: %w", err)
	}
	defer rows.Close()

	var results []models.WeeklyResortStats
	for rows.Next() {
		var stat models.WeeklyResortStats
		if err := rows.Scan(&stat.ResortID, &stat.Name, &stat.Prefecture, &stat.TotalSnowfall, &stat.YearsWithData,
			&stat.TopElevationM, &stat.BaseElevationM, &stat.VerticalM, &stat.NumCourses, &stat.LongestCourseKM); err != nil {
			return nil, fmt.Errorf("scan result: %w", err)
		}
		results = append(results, stat)
	}

	return results, nil
}
