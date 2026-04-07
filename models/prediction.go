package models

// PredictionData holds the full predictions output.
type PredictionData struct {
	GeneratedAt  string                `json:"generated_at"`
	Source       string                `json:"source"`
	ForecastDays int                   `json:"forecast_days"`
	Resorts      map[string]Prediction `json:"resorts"`
}

// Prediction holds forecast data for a single resort.
type Prediction struct {
	Name               string          `json:"name"`
	Slug               string          `json:"slug"`
	Prefecture         string          `json:"prefecture"`
	Latitude           *float64        `json:"lat"`
	Longitude          *float64        `json:"lon"`
	Elevation          *int            `json:"elevation"`
	Sources            []string        `json:"sources"`
	Daily              []DailyForecast `json:"daily"`
	HourlySnowfall     []float64       `json:"hourly_snowfall"`
	HourlyTemp         []float64       `json:"hourly_temp"`
	HourlyWind         []float64       `json:"hourly_wind"`
	HourlyWindGusts    []float64       `json:"hourly_wind_gusts"`
	HourlyWindDir      []int           `json:"hourly_wind_direction"`
	HourlyPrecip       []float64       `json:"hourly_precip"`
	HourlyRain         []float64       `json:"hourly_rain"`
	HourlyApparentTemp []float64       `json:"hourly_apparent_temp"`
	HourlyTimes        []string        `json:"hourly_times"`
}

// DailyForecast holds one day of forecast data.
type DailyForecast struct {
	Date              string  `json:"date"`
	SnowfallCM        float64 `json:"snowfall_cm"`
	TempMax           float64 `json:"temp_max"`
	TempMin           float64 `json:"temp_min"`
	PrecipitationMM   float64 `json:"precipitation_mm"`
	RainMM            float64 `json:"rain_mm"`
	WindSpeedMaxKmh   float64 `json:"wind_speed_max_kmh"`
	WindGustsMaxKmh   float64 `json:"wind_gusts_max_kmh"`
	ApparentTempMin   float64 `json:"apparent_temp_min"`
	WeatherCode       int     `json:"weather_code"`
	VsHistoricalAvgCM float64 `json:"vs_historical_avg_cm"`
	HistoricalAvgCM   float64 `json:"historical_avg_cm"`
	// Meteoblue-enriched fields (zero/nil when meteoblue not available)
	SnowmeltMM           float64     `json:"snowmelt_mm"`
	PrecipType           *string     `json:"precip_type"`
	SnowProbPct          *float64    `json:"snow_probability_pct"`
	SnowFraction         *float64    `json:"snow_fraction"`
	PowderProbability    *PowderProb `json:"powder_probability"`
	SnowfallRangeLow     float64     `json:"snowfall_range_low"`
	SnowfallRangeHigh    float64     `json:"snowfall_range_high"`
	HistoricalPercentile int         `json:"historical_percentile"`
	WindDirectionDeg     *int        `json:"wind_direction_deg"`
	Sunrise              *string     `json:"sunrise"`
	Sunset               *string     `json:"sunset"`
	// WxTech-enriched fields (omitted when wxtech not available)
	WxTechWeatherCode    *int     `json:"wxtech_weather_code,omitempty"`
	PrecipProbabilityPct *float64 `json:"precip_probability_pct,omitempty"`
}

// PowderProb holds ensemble-based snowfall probability thresholds.
type PowderProb struct {
	Exceeds5cm  int `json:"exceeds_5cm"`
	Exceeds10cm int `json:"exceeds_10cm"`
	Exceeds20cm int `json:"exceeds_20cm"`
	Exceeds30cm int `json:"exceeds_30cm"`
}
