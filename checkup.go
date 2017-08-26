package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/McKael/checkup"
)

var config checkup.Checkup
var index map[string]int64
var checkupFiles []string
var storageReader checkup.StorageReader

var lastUpdate struct {
	timestamp            time.Time
	tsMutex, updateMutex sync.Mutex
}

func loadInitialData() error {
	// Check we're using a supported storage backend
	if sr, ok := config.Storage.(checkup.StorageReader); ok {
		storageReader = sr
	} else {
		return errors.New("unsupported storage backend")
	}

	// Get index
	idx, err := storageReader.GetIndex()
	if err != nil {
		return errors.Wrap(err, "cannot read index")
	}

	index = idx

	chkFiles := make([]string, 0, len(index))
	for i := range index {
		chkFiles = append(chkFiles, i)
	}

	sort.Slice(chkFiles, func(i, j int) bool {
		return index[chkFiles[i]] < index[chkFiles[j]]
	})

	checkupFiles = chkFiles
	buildTimelines(checkupFiles)
	return nil
}

func updateData() error {
	//log.Print("Updating data") // DBG

	lastUpdate.tsMutex.Lock()
	if time.Since(lastUpdate.timestamp) < 20*time.Second {
		//log.Print("NOT updating data") // DBG
		lastUpdate.tsMutex.Unlock()
		return nil
	}
	lastUpdate.timestamp = time.Now()
	lastUpdate.tsMutex.Unlock()

	lastUpdate.updateMutex.Lock()
	defer lastUpdate.updateMutex.Unlock()

	// Refresh index
	indexNew, err := storageReader.GetIndex()
	if err != nil {
		return errors.Wrap(err, "cannot read index")
	}

	var newChkFiles []string
	for i := range indexNew {
		if _, ok := index[i]; !ok {
			newChkFiles = append(newChkFiles, i)
			index[i] = indexNew[i]
		}
	}

	//log.Printf("Updating data: %d more records\n", len(newChkFiles)) // DBG

	sort.Slice(newChkFiles, func(i, j int) bool {
		return index[newChkFiles[i]] < index[newChkFiles[j]]
	})

	checkupFiles = append(checkupFiles, newChkFiles...)
	sort.Slice(checkupFiles, func(i, j int) bool {
		return index[checkupFiles[i]] < index[checkupFiles[j]]
	})

	buildTimelines(newChkFiles)

	return nil
}

func loadCheckupConfiguration(configFile string) error {
	configBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		errors.Wrap(err, "cannot read configuration file")
	}

	var c checkup.Checkup
	err = json.Unmarshal(configBytes, &c)
	if err != nil {
		errors.Wrap(err, "cannot parse configuration file")
	}

	config = c
	return nil
}

func getLatestCheck(t *time.Time) ([]checkup.Result, error) {
	// TODO use cached data
	l := len(checkupFiles)
	if l == 0 {
		return nil, nil
	}

	*t = time.Unix(0, index[checkupFiles[l-1]]).Local()
	return storageReader.Fetch(checkupFiles[l-1])
}

func getSiteLatestCheck(site string) (checkup.Result, error) {
	var checkRes checkup.Result
	stl, err := GetTimeline(site, false, 0, 0, 1)
	if err != nil {
		return checkRes, err
	}

	l := len(stl)
	if l == 0 {
		return checkRes, nil
	}

	checkRes = *(stl[l-1].Result)
	return checkRes, nil
}

func startUpdateTimer() {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			if err := updateData(); err != nil {
				log.Println("Error while updating: ", err)
			}
		}
	}()
}

func initializeOutput() {
	resetLogOutput()
	disableColors()
}

// resetLogOutput reinitiaizes log output that is automatically
// disabled by checkup (!)
func resetLogOutput() {
	log.SetOutput(os.Stderr)
}

// disableColors prevents checkup from using ANSI escape
// sequences in strings
func disableColors() {
	checkup.DisableColor()
}
