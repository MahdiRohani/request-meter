package main

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

const (
	TotalRequests = 30000
	Concurrency   = 500
	CounterName   = "black_friday_sale"
)

var endpoints = []string{
	"http://localhost:8081",
	"http://localhost:8082",
	"http://localhost:8083",
}

func main() {
	fmt.Printf("Starting Load Test: %d requests with %d concurrency...\n", TotalRequests, Concurrency)
	startTime := time.Now()

	var successCount int32
	var errorCount int32

	jobs := make(chan int, TotalRequests)
	var wg sync.WaitGroup

	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        1000,
			MaxIdleConnsPerHost: 300,
		},
		Timeout: 5 * time.Second,
	}

	for i := 0; i < Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for jobID := range jobs {
				url := fmt.Sprintf("%s/counter/%s", endpoints[jobID%3], CounterName)

				req, _ := http.NewRequest("POST", url, nil)
				resp, err := client.Do(req)

				if err != nil {
					atomic.AddInt32(&errorCount, 1)
					continue
				}

				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()

				if resp.StatusCode == http.StatusOK {
					atomic.AddInt32(&successCount, 1)
				} else {
					atomic.AddInt32(&errorCount, 1)
				}
			}
		}()
	}

	for i := 0; i < TotalRequests; i++ {
		jobs <- i
	}
	close(jobs)

	wg.Wait()
	duration := time.Since(startTime)

	fmt.Printf("Load Test Completed in %v\n", duration)
	fmt.Printf("Success: %d | Errors: %d\n", successCount, errorCount)

	fmt.Println("Waiting 1 second for final cache flushes to Postgres...")
	time.Sleep(1 * time.Second)

	verifyURL := fmt.Sprintf("%s/stats/%s", endpoints[0], CounterName)
	resp, err := http.Get(verifyURL)
	if err != nil {
		fmt.Printf("Failed to verify stats: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("🎯 Final Database Result: %s\n", string(body))
}
