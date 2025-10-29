package repository

import (
	"context"

	"github.com/amaumene/snowfinder-common/models"
)

// Reader provides read-only access to the database.
// This interface is used by the web application to query data.
type Reader interface {
	GetResortBySlug(ctx context.Context, slug string) (*models.Resort, error)
	GetResortByID(ctx context.Context, id string) (*models.Resort, error)
	GetSnowiestResortsForWeek(ctx context.Context, weekStart string, limit int) ([]models.WeeklyResortStats, error)
	GetSnowiestResortsForWeekByPrefecture(ctx context.Context, weekStart, prefecture string, limit int) ([]models.WeeklyResortStats, error)
	GetSnowiestResortsForDateRange(ctx context.Context, startDate, endDate string, limit int) ([]models.WeeklyResortStats, error)
	GetSnowiestResortsForDateRangeByPrefecture(ctx context.Context, startDate, endDate, prefecture string, limit int) ([]models.WeeklyResortStats, error)
	GetAllResortsWithPeaks(ctx context.Context) ([]models.ResortWithPeaks, error)
	GetPeakPeriodsForResort(ctx context.Context, resortID string) ([]models.PeakPeriod, error)
}

// Writer provides full read-write access to the database.
// This interface is used by the scraper to save collected data.
type Writer interface {
	Reader
	SaveResort(ctx context.Context, resort *models.Resort) error
	SaveSnowDepthReadings(ctx context.Context, readings []models.SnowDepthReading) error
	SaveDailySnowfall(ctx context.Context, snowfalls []models.DailySnowfall) error
}
