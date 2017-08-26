package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/McKael/checkup"
)

const giga = 1000000000

func apiCheck(c *gin.Context) {
	format := c.Query("format")

	updateData() // Refresh data

	var ts time.Time
	r, err := getLatestCheck(&ts)
	if err != nil {
		c.AbortWithError(424, errors.New("cannot get latest check"))
		log.Printf("Could not get latest check: %v", err)
		return
	}

	allHealthy := true
	for _, result := range r {
		if !result.Healthy {
			allHealthy = false
		}
	}

	status := "OK"
	if !allHealthy {
		status = "DEGRADED"
	}

	if format == "json" {
		c.JSON(http.StatusOK, &map[string]string{
			"status":    status,
			"timestamp": fmt.Sprintf("%d", ts.Unix()),
		})
		return
	}
	c.String(200, "%s\t%v", status, ts.Round(time.Second).String())
}

func apiStatus(c *gin.Context) {
	format := c.Query("format")   // Output format (json, plain)
	details := c.Query("details") // If set to 1, display detailed results (plain output)

	updateData() // Refresh data

	var ts time.Time
	r, err := getLatestCheck(&ts)
	if err != nil {
		c.AbortWithError(424, errors.New("cannot get latest check"))
		log.Printf("Could not get latest check: %v", err)
		return
	}

	if format == "" || format == "json" {
		c.JSON(http.StatusOK, &map[string]interface{}{
			"results":   r,
			"timestamp": fmt.Sprintf("%d", ts.UnixNano()),
		})
		return
	}

	// Build a plain string
	buffer := new(bytes.Buffer)

	for _, siteCheck := range r {
		if details == "1" {
			fmt.Fprint(buffer, siteCheck)
		} else {
			fprintResult(buffer, siteCheck)
		}
		fmt.Fprintf(buffer, "» Timestamp: %v\n\n", time.Unix(siteCheck.Timestamp/giga, 0))
		//fmt.Fprintln(buffer)
	}
	c.String(200, "Check timestamp: %v\n\n%s", ts.Round(time.Second), buffer.String())
}

func apiStatusSite(c *gin.Context) {
	site := c.Param("site")

	format := c.Query("format")   // Output format (json, plain)
	details := c.Query("details") // If set to 1, display detailed results (plain output)

	if site == "" {
		apiStatus(c) // Redirecting...
		return
	}

	updateData() // Refresh data

	r, err := getSiteLatestCheck(site)
	if err != nil {
		c.AbortWithError(424, errors.New("cannot get site check"))
		log.Printf("Could not get site latest check: %v", err)
		return
	}

	if format == "" || format == "json" {
		c.JSON(http.StatusOK, r)
		return
	}

	// Build a plain string
	buffer := new(bytes.Buffer)
	if details == "1" {
		fmt.Fprint(buffer, r)
	} else {
		fprintResult(buffer, r)
	}
	fmt.Fprintf(buffer, "» Timestamp: %v\n\n", time.Unix(r.Timestamp/giga, 0))
	c.String(200, "%s", buffer.String())
}

func apiStats(c *gin.Context) {
	site := c.Param("site")

	format := c.Query("format")      // Output format (json, plain)
	limitStr := c.Query("limit")     // Limit data to the N latest items
	timeStartStr := c.Query("start") // Start timestamp (UNIX epoch)
	timeEndStr := c.Query("end")     // End timestamp (UNIX epoch)

	var limit int
	if z, err := strconv.Atoi(limitStr); err == nil && z > 0 {
		limit = z
	}
	timeStart := strSecondsToUnixNano(timeStartStr)
	timeEnd := strSecondsToUnixNano(timeEndStr)

	updateData() // Refresh data

	s, err := GetStats(site, timeStart, timeEnd, limit)
	if err != nil {
		c.AbortWithError(424, errors.New("cannot get statistics"))
		log.Printf("Could not get stats: %v", err)
		return
	}

	if format == "" || format == "json" {
		c.JSON(http.StatusOK, s)
		return
	}

	buffer := new(bytes.Buffer)
	fmt.Fprintf(buffer, "First check timestamp: %v\n", time.Unix(s.FirstTimestamp/giga, 0))
	fmt.Fprintf(buffer, "Last  check timestamp: %v\n", time.Unix(s.LastTimestamp/giga, 0))
	fmt.Fprintf(buffer, "Check count: %d\n", s.ItemsCount)
	fmt.Fprintf(buffer, "Healthy/Degraded/Down counts: %d/%d/%d\n", s.HealthyCount, s.DegradedCount, s.DownCount)
	fmt.Fprintf(buffer, "Healthy/Degraded/Down %% rate: %.02f/%.02f/%.02f\n", s.HealthyPC, s.DegradedPC, s.DownPC)
	c.String(200, "%s", buffer.String())
}

func apiTimeline(c *gin.Context) {
	site := c.Param("site")

	format := c.Query("format")       // Output format (json, plain)
	allStates := c.Query("allstates") // Set to 1 to retrieve all state changes
	limitStr := c.Query("limit")      // Limit data to the N latest items
	timeStartStr := c.Query("start")  // Start timestamp (UNIX epoch)
	timeEndStr := c.Query("end")      // End timestamp (UNIX epoch)
	reverse := c.Query("reverse")     // Reverse the order for plain output
	details := c.Query("details")     // If set to 1, display detailed results (plain output)

	var limit int
	if z, err := strconv.Atoi(limitStr); err == nil && z > 0 {
		limit = z
	}
	timeStart := strSecondsToUnixNano(timeStartStr)
	timeEnd := strSecondsToUnixNano(timeEndStr)

	//log.Printf("* Request timeline for '%s'\n", site)

	updateData() // Refresh data

	tl, err := GetTimeline(site, allStates != "1", timeStart, timeEnd, limit)
	if err != nil {
		c.AbortWithError(424, errors.New("cannot get timeline"))
		log.Printf("Could not get timeline: %v", err)
		return
	}

	/*
		if len(tl) == 0 {
			c.AbortWithError(204, errors.New("empty result"))
			return
		}
	*/

	if format == "" || format == "json" {
		c.JSON(http.StatusOK, tl)
		return
	}

	if reverse == "1" {
		reverseTL(tl)
	}

	// Build a plain string
	buffer := new(bytes.Buffer)

	for _, tli := range tl {
		if details == "1" {
			fmt.Fprint(buffer, tli.Result)
		} else {
			fprintResult(buffer, *tli.Result)
		}
		if allStates == "1" {
			fmt.Fprintf(buffer, "State change: %v\n", tli.StateChange)
		}
		fmt.Fprintf(buffer, "» Timestamp: %v\n\n", time.Unix(0, tli.Timestamp))
	}

	c.String(200, "%s", buffer.String())
}

func apiCheckup(c *gin.Context) {
	limitStr := c.Query("limit")                // Limit data to the N latest items
	timeStartStr := c.Query("start")            // Start timestamp (UNIX epoch)
	timeEndStr := c.Query("end")                // End timestamp (UNIX epoch)
	statsTimeStartStr := c.Query("stats_start") // Start timestamp for stats
	statsTimeEndStr := c.Query("stats_end")     // End timestamp for stats

	var limit int
	if z, err := strconv.Atoi(limitStr); err == nil && z > 0 {
		limit = z
	}
	timeStart := strSecondsToUnixNano(timeStartStr)
	timeEnd := strSecondsToUnixNano(timeEndStr)

	updateData() // Refresh data

	// Status
	var ts time.Time
	r, err := getLatestCheck(&ts)
	if err != nil {
		c.AbortWithError(424, errors.New("cannot get latest check"))
		log.Printf("Could not get latest check: %v", err)
		return
	}

	allHealthy := true
	for _, result := range r {
		if !result.Healthy {
			allHealthy = false
		}
	}

	status := "OK"
	if !allHealthy {
		status = "DEGRADED"
	}

	// Timeline
	tl, err := GetTimeline("", true, timeStart, timeEnd, limit)
	if err != nil {
		c.AbortWithError(424, errors.New("cannot get timeline"))
		log.Printf("Could not get timeline: %v", err)
		return
	}

	// Statistics
	statsTimeStart := strSecondsToUnixNano(statsTimeStartStr)
	statsTimeEnd := strSecondsToUnixNano(statsTimeEndStr)
	if statsTimeStart == 0 {
		statsTimeStart = timeStart
	}
	if statsTimeEnd == 0 {
		statsTimeEnd = timeEnd
	}
	stats, err := GetStats("", statsTimeStart, statsTimeEnd, 0)
	if err != nil {
		c.AbortWithError(424, errors.New("cannot get statistics"))
		log.Printf("Could not get stats: %v", err)
		return
	}

	c.JSON(http.StatusOK, &map[string]interface{}{
		"status":       status,
		"last_results": r,
		"stats":        stats,
		"timeline":     tl,
		"timestamp":    fmt.Sprintf("%d", ts.UnixNano()),
	})
	return
}

func reverseTL(tl Timeline) {
	for i, j := 0, len(tl)-1; i < j; i, j = i+1, j-1 {
		tl[i], tl[j] = tl[j], tl[i]
	}
}

func fprintResult(w io.Writer, r checkup.Result) {
	fmt.Fprintf(w, "== %s - %s\n", r.Title, r.Endpoint)
	assess := r.Status()
	fmt.Fprintf(w, " Assessment: %s\n", assess)
	if assess != "healthy" {
		fmt.Fprintf(w, "    Results: %v\n", r.Times)
	}
}

func strSecondsToUnixNano(strSeconds string) int64 {
	if z, err := strconv.ParseInt(strSeconds, 10, 64); err == nil && z > 0 {
		return z * giga
	}
	return 0
}
