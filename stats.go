package main

import (
	"log"

	"github.com/pkg/errors"
)

// Stats contains statistics for a give site.
type Stats struct {
	FirstTimestamp, LastTimestamp          int64
	ItemsCount                             int
	HealthyCount, DegradedCount, DownCount int
	HealthyPC, DegradedPC, DownPC          float64
}

// GetStats returns computed statistics for a given title.
// If title is empty, stats for all sites in the global timeline are returned.
// Start & End timestamps can be provided, they are ignored if set to 0.
// Those timestamps are in nanoseconds since 1970-01-01 00:00:00 UTC.
// The maximum number of studied items can be set with the 'limit' value.
func GetStats(title string, startTS, endTS int64, limit int) (Stats, error) {
	var stats Stats

	tl, err := GetTimeline(title, false, startTS, endTS, limit)
	if err != nil {
		return stats, errors.Wrap(err, "cannot get timeline")
	}

	stats.ItemsCount = len(tl)

	if stats.ItemsCount == 0 {
		return stats, nil
	}

	// Set timestamps; tl is supposed to be sorted
	stats.FirstTimestamp = tl[0].Result.Timestamp
	stats.LastTimestamp = tl[stats.ItemsCount-1].Result.Timestamp

	for _, cr := range tl {
		if cr.Result.Healthy {
			stats.HealthyCount++
			continue
		}
		if cr.Result.Down {
			stats.DownCount++
			continue
		}
		if cr.Result.Degraded {
			stats.DegradedCount++
		}
		log.Println("Warning: check result with no result status")
	}

	stats.HealthyPC = float64(stats.HealthyCount*100) / float64(stats.ItemsCount)
	stats.DegradedPC = float64(stats.DegradedCount*100) / float64(stats.ItemsCount)
	stats.DownPC = float64(stats.DownCount*100) / float64(stats.ItemsCount)

	return stats, nil
}
