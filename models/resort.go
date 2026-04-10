package models

import "time"

// Resort represents a ski resort with its physical characteristics.
type Resort struct {
	ID                string    `json:"id"`
	Slug              string    `json:"slug"`
	Name              string    `json:"name"`
	Prefecture        string    `json:"prefecture"`
	Region            string    `json:"region"`
	TopElevationM     *int      `json:"top_elevation_m"`
	BaseElevationM    *int      `json:"base_elevation_m"`
	VerticalM         *int      `json:"vertical_m"`
	NumCourses        *int      `json:"num_courses"`
	LongestCourseKM   *float64  `json:"longest_course_km"`
	SteepestCourseDeg *float64  `json:"steepest_course_deg"`
	LastUpdated       time.Time `json:"last_updated"`
}

// SnowDepthReading is a point-in-time snow depth measurement at a resort.
type SnowDepthReading struct {
	ResortID string    `json:"resort_id"`
	Date     time.Time `json:"date"`
	DepthCM  int       `json:"depth_cm"`
}

// DailySnowfall records the total snowfall in centimetres for a single day at a resort.
type DailySnowfall struct {
	ResortID   string    `json:"resort_id"`
	Date       time.Time `json:"date"`
	SnowfallCM int       `json:"snowfall_cm"`
}

// WeeklyResortStats aggregates average snowfall statistics for a resort over a
// calendar date range, used by the "snowiest resorts" query.
type WeeklyResortStats struct {
	ResortID        string   `json:"resort_id"`
	Name            string   `json:"name"`
	Prefecture      string   `json:"prefecture"`
	TotalSnowfall   *int     `json:"total_snowfall"`
	YearsWithData   *int     `json:"years_with_data"`
	TopElevationM   *int     `json:"top_elevation_m"`
	BaseElevationM  *int     `json:"base_elevation_m"`
	VerticalM       *int     `json:"vertical_m"`
	NumCourses      *int     `json:"num_courses"`
	LongestCourseKM *float64 `json:"longest_course_km"`
}

// PeakPeriod describes a historically significant snowfall peak window for a resort.
// StartDate, EndDate, and CenterDate are formatted as "MM-DD".
type PeakPeriod struct {
	ID                  string    `json:"id"`
	ResortID            string    `json:"resort_id"`
	PeakRank            int       `json:"peak_rank"`
	StartDate           string    `json:"start_date"`
	EndDate             string    `json:"end_date"`
	CenterDate          string    `json:"center_date"`
	AvgDailySnowfall    float64   `json:"avg_daily_snowfall"`
	TotalPeriodSnowfall float64   `json:"total_period_snowfall"`
	ProminenceScore     float64   `json:"prominence_score"`
	YearsOfData         int       `json:"years_of_data"`
	ConfidenceLevel     string    `json:"confidence_level"`
	CalculatedAt        time.Time `json:"calculated_at"`
}

// ResortWithPeaks bundles a Resort with its associated PeakPeriods.
type ResortWithPeaks struct {
	Resort Resort       `json:"resort"`
	Peaks  []PeakPeriod `json:"peaks"`
}

// FailedScrapeAttempt records a scrape that failed and whether it has been retried.
type FailedScrapeAttempt struct {
	ID           string     `json:"id"`
	ResortURL    string     `json:"resort_url"`
	ErrorMessage string     `json:"error_message"`
	FailedAt     time.Time  `json:"failed_at"`
	Retried      bool       `json:"retried"`
	RetriedAt    *time.Time `json:"retried_at"`
}
