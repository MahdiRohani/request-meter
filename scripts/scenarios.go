package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

type StatResp struct {
	Name  string `json:"Name"`
	Count int64  `json:"Count"`
}

var client = &http.Client{Timeout: 2 * time.Second}
var db *sql.DB

func initDB() {
	var err error
	db, err = sql.Open("postgres", "postgres://admin:password@localhost:5432/meter_db?sslmode=disable")
	if err != nil {
		panic(err)
	}
}

func now() string {
	return time.Now().Format("15:04:05.000")
}

func getInstanceStat(port int, counter string) int64 {
	url := fmt.Sprintf("http://localhost:%d/stats/%s", port, counter)
	resp, err := client.Get(url)
	if err != nil {
		return -1
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var s StatResp
	json.Unmarshal(body, &s)
	return s.Count
}

func getDirectDBStat(counter string) int64 {
	var count int64
	err := db.QueryRow("SELECT count FROM counters WHERE name = $1", counter).Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

func sendRequestsGradually(port int, counter string, total int, duration time.Duration, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}
	delay := duration / time.Duration(total)
	url := fmt.Sprintf("http://localhost:%d/counter/%s", port, counter)
	for i := 0; i < total; i++ {
		req, _ := http.NewRequest("POST", url, nil)
		resp, err := client.Do(req)
		if err == nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
		time.Sleep(delay)
	}
	fmt.Printf("[%s] Finished sending %d reqs to Port %d\n", now(), total, port)
}

func startLiveMonitor(counter string, interval time.Duration, stopChan chan bool) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-stopChan:
			ticker.Stop()
			return
		case <-ticker.C:
			c1 := getInstanceStat(8081, counter)
			c2 := getInstanceStat(8082, counter)
			c3 := getInstanceStat(8083, counter)
			dbCount := getDirectDBStat(counter)
			fmt.Printf("[%s] MONITOR | 8081: %4d | 8082: %4d | 8083: %4d || 🐘 Direct Postgres DB: %4d\n",
				now(), c1, c2, c3, dbCount)
		}
	}
}

func main() {
	initDB()
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run scenarios.go [1|2|4]")
		return
	}
	scenario := os.Args[1]

	switch scenario {
	case "1":
		counter := "scen1_single"
		db.Exec("DELETE FROM counters WHERE name = $1", counter)
		stop := make(chan bool)
		fmt.Printf("[%s] 🚀 Scenario 1: Sending 1000 reqs to 8081 over 1.5 seconds...\n", now())
		go startLiveMonitor(counter, 150*time.Millisecond, stop)
		sendRequestsGradually(8081, counter, 1000, 1500*time.Millisecond, nil)
		time.Sleep(1 * time.Second)
		stop <- true

	case "2":

		counter := "scen2_multi"
		db.Exec("DELETE FROM counters WHERE name = $1", counter)
		stop := make(chan bool)
		var wg sync.WaitGroup
		fmt.Printf("[%s] 🚀 Scenario 2: 1000->8081, 500->8082, 800->8083 over 2 seconds...\n", now())
		go startLiveMonitor(counter, 200*time.Millisecond, stop)
		wg.Add(3)
		go sendRequestsGradually(8081, counter, 1000, 2*time.Second, &wg)
		go sendRequestsGradually(8082, counter, 500, 2*time.Second, &wg)
		go sendRequestsGradually(8083, counter, 800, 2*time.Second, &wg)
		wg.Wait()
		time.Sleep(1500 * time.Millisecond)
		stop <- true

	case "4":

		counter := "scen4_one_click"
		db.Exec("DELETE FROM counters WHERE name = $1", counter)
		stop := make(chan bool)
		fmt.Printf("[%s] Scenario 4: Sending ONLY 1 request to 8081...\n", now())
		go startLiveMonitor(counter, 100*time.Millisecond, stop)
		time.Sleep(200 * time.Millisecond)
		sendRequestsGradually(8081, counter, 1, 1*time.Millisecond, nil)
		time.Sleep(1500 * time.Millisecond)
		stop <- true
	}
}
