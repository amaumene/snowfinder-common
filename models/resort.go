package models

import "time"

type Resort struct {
	ID                string    `json:"id" db:"id"`
	Slug              string    `json:"slug" db:"slug"`
	Name              string    `json:"name" db:"name"`
	Prefecture        string    `json:"prefecture" db:"prefecture"`
	Region            string    `json:"region" db:"region"`
	TopElevationM     *int      `json:"top_elevation_m" db:"top_elevation_m"`
	BaseElevationM    *int      `json:"base_elevation_m" db:"base_elevation_m"`
	VerticalM         *int      `json:"vertical_m" db:"vertical_m"`
	NumCourses        *int      `json:"num_courses" db:"num_courses"`
	LongestCourseKM   *float64  `json:"longest_course_km" db:"longest_course_km"`
	SteepestCourseDeg *float64  `json:"steepest_course_deg" db:"steepest_course_deg"`
	LastUpdated       time.Time `json:"last_updated" db:"last_updated"`
}

type SnowDepthReading struct {
	ResortID string    `json:"resort_id" db:"resort_id"`
	Date     time.Time `json:"date" db:"date"`
	DepthCM  int       `json:"depth_cm" db:"depth_cm"`
	Season   string    `json:"season" db:"season"`
}

type DailySnowfall struct {
	ResortID   string    `json:"resort_id" db:"resort_id"`
	Date       time.Time `json:"date" db:"date"`
	SnowfallCM int       `json:"snowfall_cm" db:"snowfall_cm"`
	Season     string    `json:"season" db:"season"`
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
	ID                  string    `json:"id" db:"id"`
	ResortID            string    `json:"resort_id" db:"resort_id"`
	PeakRank            int       `json:"peak_rank" db:"peak_rank"`
	StartDate           string    `json:"start_date" db:"start_date"`
	EndDate             string    `json:"end_date" db:"end_date"`
	CenterDate          string    `json:"center_date" db:"center_date"`
	AvgDailySnowfall    float64   `json:"avg_daily_snowfall" db:"avg_daily_snowfall"`
	TotalPeriodSnowfall float64   `json:"total_period_snowfall" db:"total_period_snowfall"`
	ProminenceScore     float64   `json:"prominence_score" db:"prominence_score"`
	YearsOfData         int       `json:"years_of_data" db:"years_of_data"`
	ConfidenceLevel     string    `json:"confidence_level" db:"confidence_level"`
	CalculatedAt        time.Time `json:"calculated_at" db:"calculated_at"`
}

type ResortWithPeaks struct {
	Resort Resort
	Peaks  []PeakPeriod
}

type FailedScrapeAttempt struct {
	ID           string     `json:"id" db:"id"`
	ResortURL    string     `json:"resort_url" db:"resort_url"`
	ErrorMessage string     `json:"error_message" db:"error_message"`
	FailedAt     time.Time  `json:"failed_at" db:"failed_at"`
	Retried      bool       `json:"retried" db:"retried"`
	RetriedAt    *time.Time `json:"retried_at" db:"retried_at"`
}
