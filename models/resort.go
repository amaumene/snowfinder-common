package models

import "time"

type Resort struct {
	ID              string    `json:"id" db:"id"`
	Slug            string    `json:"slug" db:"slug"`
	Name            string    `json:"name" db:"name"`
	Prefecture      string    `json:"prefecture" db:"prefecture"`
	Region          string    `json:"region" db:"region"`
	TopElevationM   *int      `json:"top_elevation_m" db:"top_elevation_m"`
	BaseElevationM  *int      `json:"base_elevation_m" db:"base_elevation_m"`
	VerticalM       *int      `json:"vertical_m" db:"vertical_m"`
	NumCourses      *int      `json:"num_courses" db:"num_courses"`
	LongestCourseKM *float64  `json:"longest_course_km" db:"longest_course_km"`
	SteepestCourseDeg *float64 `json:"steepest_course_deg" db:"steepest_course_deg"`
	LastUpdated     time.Time `json:"last_updated" db:"last_updated"`
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

type WeeklyStats struct {
	ResortID       string    `json:"resort_id" db:"resort_id"`
	WeekStart      time.Time `json:"week_start" db:"week_start"`
	WeekEnd        time.Time `json:"week_end" db:"week_end"`
	TotalSnowfallCM int      `json:"total_snowfall_cm" db:"total_snowfall_cm"`
	AvgDepthCM     float64   `json:"avg_depth_cm" db:"avg_depth_cm"`
	NumSnowDays    int       `json:"num_snow_days" db:"num_snow_days"`
}

type ResortURL struct {
	BaseURL  string
	InfoURL  string
	SnowURL  string
	Slug     string
	Prefecture string
	Area     string
}

type WeeklyResortStats struct {
	ResortID        string
	Name            string
	Prefecture      string
	TotalSnowfall   *int
	YearsWithData   *int
	TopElevationM   *int
	BaseElevationM  *int
	VerticalM       *int
	NumCourses      *int
	LongestCourseKM *float64
}
