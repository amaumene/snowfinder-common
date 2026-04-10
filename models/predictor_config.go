package models

import "encoding/json"

// PredictorResortConfig is the per-resort configuration loaded from
// the prediction_config table's config_data JSONB column.
type PredictorResortConfig struct {
	Name        string                      `json:"name"`
	Slug        string                      `json:"slug"`
	Prefecture  string                      `json:"prefecture"`
	Lat         *float64                    `json:"lat"`
	Lon         *float64                    `json:"lon"`
	Elevation   *int                        `json:"elevation"`
	Climatology map[string]ClimatologyEntry `json:"climatology"`
	BiasFactors map[string]float64          `json:"bias_factors"`
}

// ClimatologyEntry holds historical snowfall statistics for a single
// calendar day (keyed by "MM-DD" in the Climatology map).
type ClimatologyEntry struct {
	Avg float64  `json:"avg"`
	Std float64  `json:"std"`
	P10 *float64 `json:"p10,omitempty"`
	P25 *float64 `json:"p25,omitempty"`
	P50 *float64 `json:"p50,omitempty"`
	P75 *float64 `json:"p75,omitempty"`
	P90 *float64 `json:"p90,omitempty"`
}

// GlobalParams holds the global predictor parameters from
// prediction_global_params.params_data JSONB.
type GlobalParams struct {
	BlendWeights  map[string][]float64 `json:"blend_weights,omitempty"`
	BlendW0       float64              `json:"blend_w0,omitempty"`
	BlendDecay    float64              `json:"blend_decay,omitempty"`
	MBCapMult     float64              `json:"mb_cap_multiplier,omitempty"`
	MBCapFloor    float64              `json:"mb_cap_floor_cm,omitempty"`
	SWRThresholds map[string]SWREntry  `json:"swr_thresholds,omitempty"`
}

// SWREntry is a snow-water ratio threshold entry.
type SWREntry struct {
	BelowTemp *float64 `json:"below_temp,omitempty"`
	Ratio     float64  `json:"ratio"`
}

// PredictorConfig is the full config loaded from DB.
type PredictorConfig struct {
	Resorts        map[string]PredictorResortConfig `json:"resorts"`
	GlobalParams   GlobalParams                     `json:"global_params"`
	JMAOfficeCodes json.RawMessage                  `json:"jma_office_codes,omitempty"`
}
