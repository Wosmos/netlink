package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	baseURL        = "http://localhost:8080"
	numUsers       = 100  // Concurrent virtual users
	testDuration   = 30   // Seconds
	tasksPerUser   = 5    // Tasks each user creates
)

var (
	successCount int64
	errorCount   int64
	totalLatency int64
)

func main() {
	fmt.Printf("🚀 Load Test Starting\n")
	fmt.Printf("   Users: %d concurrent\n", numUsers)
	fmt.Printf("   Duration: %ds\n", testDuration)
	fmt.Printf("   Tasks per user: %d\n\n", tasksPerUser)

	var wg sync.WaitGroup
	start := time.Now()
	stop := make(chan struct{})

	// Launch virtual users
	for i := 0; i < numUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			simulateUser(userID, stop)
		}(i)
	}

	// Run for specified duration
	time.Sleep(time.Duration(testDuration) * time.Second)
	close(stop)
	wg.Wait()

	elapsed := time.Since(start)
	total := successCount + errorCount
	avgLatency := float64(0)
	if successCount > 0 {
		avgLatency = float64(totalLatency) / float64(successCount) / float64(time.Millisecond)
	}

	fmt.Printf("\n📊 Results\n")
	fmt.Printf("   Total Requests: %d\n", total)
	fmt.Printf("   Successful: %d\n", successCount)
	fmt.Printf("   Failed: %d\n", errorCount)
	fmt.Printf("   Requests/sec: %.2f\n", float64(total)/elapsed.Seconds())
	fmt.Printf("   Avg Latency: %.2fms\n", avgLatency)
	fmt.Printf("   Error Rate: %.2f%%\n", float64(errorCount)/float64(total)*100)
}

func simulateUser(userID int, stop chan struct{}) {
	// Each user gets their own cookie jar (session)
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar:     jar,
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects
		},
	}

	email := fmt.Sprintf("loadtest%d_%d@test.com", userID, time.Now().UnixNano())
	password := "testpass123"

	// Register
	if !doRequest(client, "POST", "/register", url.Values{
		"email":            {email},
		"password":         {password},
		"confirm_password": {password},
	}) {
		return
	}

	// Main loop - create tasks, toggle, delete
	taskNum := 0
	for {
		select {
		case <-stop:
			return
		default:
		}

		// View home
		doRequest(client, "GET", "/", nil)

		// Add task
		taskNum++
		doRequest(client, "POST", "/add", url.Values{
			"task": {fmt.Sprintf("Task %d from user %d", taskNum, userID)},
		})

		// Toggle task (using task ID 1 as example)
		doRequest(client, "POST", "/toggle", url.Values{"id": {"1"}})

		// Small delay to simulate real user
		time.Sleep(50 * time.Millisecond)
	}
}

func doRequest(client *http.Client, method, path string, data url.Values) bool {
	start := time.Now()

	var req *http.Request
	var err error

	fullURL := baseURL + path

	if method == "POST" && data != nil {
		req, err = http.NewRequest(method, fullURL, strings.NewReader(data.Encode()))
		if err != nil {
			atomic.AddInt64(&errorCount, 1)
			return false
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req, err = http.NewRequest(method, fullURL, nil)
		if err != nil {
			atomic.AddInt64(&errorCount, 1)
			return false
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		atomic.AddInt64(&errorCount, 1)
		return false
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body) // Drain body

	latency := time.Since(start)
	atomic.AddInt64(&totalLatency, int64(latency))

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		atomic.AddInt64(&successCount, 1)
		return true
	}

	atomic.AddInt64(&errorCount, 1)
	return false
}
