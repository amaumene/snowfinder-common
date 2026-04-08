package models

import "time"

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

type SnowDepthReading struct {
	ResortID string    `json:"resort_id"`
	Date     time.Time `json:"date"`
	DepthCM  int       `json:"depth_cm"`
}

type DailySnowfall struct {
	ResortID   string    `json:"resort_id"`
	Date       time.Time `json:"date"`
	SnowfallCM int       `json:"snowfall_cm"`
}

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

type ResortWithPeaks struct {
	Resort Resort       `json:"resort"`
	Peaks  []PeakPeriod `json:"peaks"`
}

type FailedScrapeAttempt struct {
	ID           string     `json:"id"`
	ResortURL    string     `json:"resort_url"`
	ErrorMessage string     `json:"error_message"`
	FailedAt     time.Time  `json:"failed_at"`
	Retried      bool       `json:"retried"`
	RetriedAt    *time.Time `json:"retried_at"`
}
