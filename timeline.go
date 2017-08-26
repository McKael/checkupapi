package main

import (
	"log"

	"github.com/pkg/errors"

	"github.com/McKael/checkup"
)

// TimelineState is a Timeline item.
type TimelineState struct {
	Result      *checkup.Result
	StateChange bool
	Timestamp   int64
}

// Timeline is a list of checks for a give site.
type Timeline []TimelineState

var globalTimeline Timeline
var timelines map[string]Timeline

func buildTimelines(checkFiles []string) {
	// Note: 'timelines' is not reset, so that the build can be incremental
	if timelines == nil {
		timelines = make(map[string]Timeline)
	}

	for _, f := range checkFiles {
		t := index[f]
		results, err := storageReader.Fetch(f)
		if err != nil {
			log.Printf("Error: cannot read '%s' (%s), skipping...\n", f, err.Error())
			// Remove from the index so that it's retried later.
			delete(index, f)
			continue
		}

		for _, r := range results {
			r := r
			changed := timelineAppend(r.Title, t, r)
			gtlst := TimelineState{
				Result:      &r,
				Timestamp:   t,
				StateChange: changed,
			}
			globalTimeline = append(globalTimeline, gtlst)
		}
	}
}

func timelineAppend(title string, timestamp int64, r checkup.Result) (changed bool) {
	tlst := TimelineState{
		Result:    &r,
		Timestamp: timestamp,
	}

	tl, ok := timelines[title]
	if !ok {
		tlst.StateChange = true
	} else {
		l := len(tl)
		if l > 0 { // changed ?
			prevRes := tl[l-1].Result
			if prevRes.Healthy != r.Healthy ||
				prevRes.Degraded != r.Degraded ||
				prevRes.Down != r.Down {
				tlst.StateChange = true
			}
		}
	}
	changed = tlst.StateChange
	timelines[title] = append(timelines[title], tlst)
	return
}

// GetTimeline returns the timeline for a given title.
// If changes is true, only the checks with a state change are returned.
// Start & End timestamps can be provided, they are ignored if set to 0.
// Those timestamps are in nanoseconds since 1970-01-01 00:00:00 UTC.
// The maximum number of items to be returned can be set with the 'limit' value.
func GetTimeline(title string, changes bool, startTS, endTS int64, limit int) (Timeline, error) {
	var tl Timeline
	if title == "" {
		tl = globalTimeline
	} else {
		var ok bool
		tl, ok = timelines[title]
		if !ok {
			return nil, errors.Errorf("timeline for '%s' not found", title)
		}
	}

	if !changes && startTS == 0 && endTS == 0 {
		// Return all checks
		if limit == 0 || limit >= len(tl) {
			return tl, nil
		}
		return tl[len(tl)-limit:], nil
	}

	// Return only state changes
	var newTl Timeline

	for _, tlst := range tl {
		// Filter boundaries
		if startTS > 0 && tlst.Result.Timestamp < startTS {
			continue
		}
		if endTS > 0 && tlst.Result.Timestamp > endTS {
			continue
		}
		// Filter state changes
		if changes && !tlst.StateChange {
			continue
		}
		newTl = append(newTl, tlst)
	}

	// Return all checks
	if limit > 0 && limit < len(newTl) {
		return newTl[len(newTl)-limit:], nil
	}
	return newTl, nil
}
