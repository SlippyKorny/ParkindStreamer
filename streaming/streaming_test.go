package streaming

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

// getMemUsage returns the system allocated memory in bytes
func getMemUsage() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return m.Sys
}

// TestCameraStreaming tests the functionality of the camera session
func TestCameraStreaming(t *testing.T) {
	camCount := 1
	cs, err := NewCameraSession(camCount)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}
	defer cs.Close()

	// Declare arrays for storing information about memory usage
	var memStats []uint64
	// var memAvg uint64 = 0

	// Run a loop until 5 seconds pass
	start := time.Now()
	for time.Since(start).Seconds() < 5.9 {
		// load frames
		ptrs, err := cs.GetFrames()
		if err != nil {
			t.Log(err.Error())
			t.Fail()
		}

		// record the memory usage
		mu := getMemUsage()
		memStats = append(memStats, mu)
		fmt.Printf("\rptrs address: %d; Taken RAM: %d", ptrs[0], mu) // so that the compiler ignores unused error

		// wait for 0.3 seconds
		time.Sleep(time.Millisecond * 300)
	}
	fmt.Println()

	// Check if there's a memory leak and if there is one then fail the test
	hBound := float64(memStats[0]) + (float64(0.05) * float64(memStats[0]))
	for i := 1; i < len(memStats); i++ {
		if float64(memStats[i]) > hBound {
			t.Logf("memory disparity between the first and the %d th entry exceeding 5 precent", i)
			t.Log(memStats)
			t.Fail()
		}
	}
}
