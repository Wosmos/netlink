package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	baseURL      = "http://localhost:8080"
	totalUsers   = 500              // Neon free tier: ~100 connections, so 500 concurrent users is realistic
	testDuration = 60 * time.Second
	rampUpTime   = 15 * time.Second
	defaultPass  = "Test@123"
)

type ErrorStats struct {
	Timeout     int64
	ConnRefused int64
	BadStatus   int64
	Other       int64
}

type Metrics struct {
	LoginSuccess   int64
	LoginFail      int64
	HomeSuccess    int64
	HomeFail       int64
	AddTaskSuccess int64
	AddTaskFail    int64
	ToggleSuccess  int64
	ToggleFail     int64
	DeleteSuccess  int64
	DeleteFail     int64
	LogoutSuccess  int64
	LogoutFail     int64

	LoginLatency         int64
	HomeLatency          int64
	AddTaskLatency       int64
	ToggleLatency        int64
	DeleteLatency        int64
	LogoutLatency        int64
	LoginLatencyCount    int64
	HomeLatencyCount     int64
	AddTaskLatencyCount  int64
	ToggleLatencyCount   int64
	DeleteLatencyCount   int64
	LogoutLatencyCount   int64

	LoginPeak   int64
	HomePeak    int64
	AddTaskPeak int64
	TogglePeak  int64
	DeletePeak  int64
	LogoutPeak  int64

	Errors ErrorStats
}

var metrics = &Metrics{}

func main() {
	fmt.Println("🔥 STRESS TEST - Go Todo App (PostgreSQL/Neon)")
	fmt.Printf("   Target: %s\n", baseURL)
	fmt.Printf("   Virtual Users: %s\n", formatNumber(totalUsers))
	fmt.Printf("   Duration: %v\n", testDuration)
	fmt.Printf("   Ramp-up: %v\n", rampUpTime)
	fmt.Printf("   CPU Cores: %d\n\n", runtime.NumCPU())

	if !checkServer() {
		fmt.Println("❌ Server not responding at", baseURL)
		fmt.Println("   Make sure server is running: go run main.go")
		return
	}
	fmt.Println("✅ Server is up!\n")

	stop := make(chan struct{})
	var wg sync.WaitGroup
	var activeUsers int64

	start := time.Now()

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				elapsed := time.Since(start)
				total := getTotalRequests()
				rps := float64(total) / elapsed.Seconds()
				fmt.Printf("   [%v] Active: %d | Req: %s | RPS: %.0f | Err: T:%d C:%d\n",
					elapsed.Round(time.Second),
					atomic.LoadInt64(&activeUsers),
					formatNumber(int(total)),
					rps,
					atomic.LoadInt64(&metrics.Errors.Timeout),
					atomic.LoadInt64(&metrics.Errors.ConnRefused),
				)
			}
		}
	}()

	usersPerBatch := totalUsers / int(rampUpTime.Seconds())
	if usersPerBatch < 1 {
		usersPerBatch = 1
	}

	fmt.Println("🚀 Ramping up users...")
	userID := 0
	rampTicker := time.NewTicker(time.Second)

rampLoop:
	for {
		select {
		case <-rampTicker.C:
			for i := 0; i < usersPerBatch && userID < totalUsers; i++ {
				userID++
				wg.Add(1)
				atomic.AddInt64(&activeUsers, 1)
				go func(uid int) {
					defer wg.Done()
					defer atomic.AddInt64(&activeUsers, -1)
					simulateUser(uid, stop)
				}(userID)
			}
			if userID >= totalUsers {
				break rampLoop
			}
		}
	}
	rampTicker.Stop()

	fmt.Printf("   All %s users launched!\n\n", formatNumber(totalUsers))

	remaining := testDuration - time.Since(start)
	if remaining > 0 {
		time.Sleep(remaining)
	}

	close(stop)
	fmt.Println("\n⏹️  Stopping users...")
	wg.Wait()

	printResults(time.Since(start))
}

func checkServer() bool {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(baseURL + "/login")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == 200
}

func simulateUser(userID int, stop chan struct{}) {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar:     jar,
		Timeout: 15 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	email := fmt.Sprintf("user%d@example.com", userID)
	randSrc := rand.New(rand.NewSource(time.Now().UnixNano() + int64(userID)))

	// Login with retries
	loggedIn := false
	for i := 0; i < 3 && !loggedIn; i++ {
		if doLogin(client, email) {
			loggedIn = true
		} else {
			time.Sleep(time.Duration(200*(i+1)) * time.Millisecond)
		}
	}
	if !loggedIn {
		return
	}

	taskCounter := 0

	for {
		select {
		case <-stop:
			doLogout(client)
			return
		default:
		}

		action := randSrc.Intn(100)

		switch {
		case action < 50:
			doHome(client)
		case action < 75:
			taskCounter++
			doAddTask(client, fmt.Sprintf("Task %d user %d", taskCounter, userID))
		case action < 88:
			doToggle(client, randSrc.Intn(10)+1)
		case action < 97:
			doDelete(client, randSrc.Intn(10)+1)
		default:
			doLogout(client)
			time.Sleep(300 * time.Millisecond)
			doLogin(client, email)
		}

		time.Sleep(time.Duration(150+randSrc.Intn(250)) * time.Millisecond)
	}
}

func trackError(err error, statusCode int) {
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline") {
			atomic.AddInt64(&metrics.Errors.Timeout, 1)
		} else if strings.Contains(errStr, "refused") || strings.Contains(errStr, "reset") {
			atomic.AddInt64(&metrics.Errors.ConnRefused, 1)
		} else {
			atomic.AddInt64(&metrics.Errors.Other, 1)
		}
	} else if statusCode >= 400 {
		atomic.AddInt64(&metrics.Errors.BadStatus, 1)
	}
}

func doLogin(client *http.Client, email string) bool {
	start := time.Now()
	resp, err := client.PostForm(baseURL+"/login", url.Values{
		"email":    {email},
		"password": {defaultPass},
	})

	latency := time.Since(start).Nanoseconds()
	atomic.AddInt64(&metrics.LoginLatency, latency)
	atomic.AddInt64(&metrics.LoginLatencyCount, 1)
	updatePeak(&metrics.LoginPeak, latency)

	if err != nil {
		trackError(err, 0)
		atomic.AddInt64(&metrics.LoginFail, 1)
		return false
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode == 303 || resp.StatusCode == 302 {
		atomic.AddInt64(&metrics.LoginSuccess, 1)
		return true
	}

	trackError(nil, resp.StatusCode)
	atomic.AddInt64(&metrics.LoginFail, 1)
	return false
}

func doHome(client *http.Client) bool {
	start := time.Now()
	resp, err := client.Get(baseURL + "/")

	latency := time.Since(start).Nanoseconds()
	atomic.AddInt64(&metrics.HomeLatency, latency)
	atomic.AddInt64(&metrics.HomeLatencyCount, 1)
	updatePeak(&metrics.HomePeak, latency)

	if err != nil {
		trackError(err, 0)
		atomic.AddInt64(&metrics.HomeFail, 1)
		return false
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode == 200 {
		atomic.AddInt64(&metrics.HomeSuccess, 1)
		return true
	}

	trackError(nil, resp.StatusCode)
	atomic.AddInt64(&metrics.HomeFail, 1)
	return false
}

func doAddTask(client *http.Client, text string) bool {
	start := time.Now()
	resp, err := client.PostForm(baseURL+"/add", url.Values{"task": {text}})

	latency := time.Since(start).Nanoseconds()
	atomic.AddInt64(&metrics.AddTaskLatency, latency)
	atomic.AddInt64(&metrics.AddTaskLatencyCount, 1)
	updatePeak(&metrics.AddTaskPeak, latency)

	if err != nil {
		trackError(err, 0)
		atomic.AddInt64(&metrics.AddTaskFail, 1)
		return false
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode == 303 || resp.StatusCode == 302 {
		atomic.AddInt64(&metrics.AddTaskSuccess, 1)
		return true
	}

	trackError(nil, resp.StatusCode)
	atomic.AddInt64(&metrics.AddTaskFail, 1)
	return false
}

func doToggle(client *http.Client, taskID int) bool {
	start := time.Now()
	resp, err := client.PostForm(baseURL+"/toggle", url.Values{"id": {fmt.Sprintf("%d", taskID)}})

	latency := time.Since(start).Nanoseconds()
	atomic.AddInt64(&metrics.ToggleLatency, latency)
	atomic.AddInt64(&metrics.ToggleLatencyCount, 1)
	updatePeak(&metrics.TogglePeak, latency)

	if err != nil {
		trackError(err, 0)
		atomic.AddInt64(&metrics.ToggleFail, 1)
		return false
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode == 303 || resp.StatusCode == 302 {
		atomic.AddInt64(&metrics.ToggleSuccess, 1)
		return true
	}

	trackError(nil, resp.StatusCode)
	atomic.AddInt64(&metrics.ToggleFail, 1)
	return false
}

func doDelete(client *http.Client, taskID int) bool {
	start := time.Now()
	resp, err := client.PostForm(baseURL+"/delete", url.Values{"id": {fmt.Sprintf("%d", taskID)}})

	latency := time.Since(start).Nanoseconds()
	atomic.AddInt64(&metrics.DeleteLatency, latency)
	atomic.AddInt64(&metrics.DeleteLatencyCount, 1)
	updatePeak(&metrics.DeletePeak, latency)

	if err != nil {
		trackError(err, 0)
		atomic.AddInt64(&metrics.DeleteFail, 1)
		return false
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode == 303 || resp.StatusCode == 302 {
		atomic.AddInt64(&metrics.DeleteSuccess, 1)
		return true
	}

	trackError(nil, resp.StatusCode)
	atomic.AddInt64(&metrics.DeleteFail, 1)
	return false
}

func doLogout(client *http.Client) bool {
	start := time.Now()
	resp, err := client.PostForm(baseURL+"/logout", url.Values{})

	latency := time.Since(start).Nanoseconds()
	atomic.AddInt64(&metrics.LogoutLatency, latency)
	atomic.AddInt64(&metrics.LogoutLatencyCount, 1)
	updatePeak(&metrics.LogoutPeak, latency)

	if err != nil {
		trackError(err, 0)
		atomic.AddInt64(&metrics.LogoutFail, 1)
		return false
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode == 303 || resp.StatusCode == 302 {
		atomic.AddInt64(&metrics.LogoutSuccess, 1)
		return true
	}

	trackError(nil, resp.StatusCode)
	atomic.AddInt64(&metrics.LogoutFail, 1)
	return false
}

func updatePeak(peak *int64, value int64) {
	for {
		old := atomic.LoadInt64(peak)
		if value <= old {
			return
		}
		if atomic.CompareAndSwapInt64(peak, old, value) {
			return
		}
	}
}

func getTotalRequests() int64 {
	return metrics.LoginSuccess + metrics.LoginFail +
		metrics.HomeSuccess + metrics.HomeFail +
		metrics.AddTaskSuccess + metrics.AddTaskFail +
		metrics.ToggleSuccess + metrics.ToggleFail +
		metrics.DeleteSuccess + metrics.DeleteFail +
		metrics.LogoutSuccess + metrics.LogoutFail
}

func avgLatency(total, count int64) float64 {
	if count == 0 {
		return 0
	}
	return float64(total) / float64(count) / 1e6
}

func printResults(totalTime time.Duration) {
	total := getTotalRequests()
	success := metrics.LoginSuccess + metrics.HomeSuccess + metrics.AddTaskSuccess +
		metrics.ToggleSuccess + metrics.DeleteSuccess + metrics.LogoutSuccess
	failed := total - success

	fmt.Println("\n" + strings.Repeat("=", 65))
	fmt.Println("📊 STRESS TEST RESULTS")
	fmt.Println(strings.Repeat("=", 65))

	fmt.Printf("\n⏱️  SUMMARY\n")
	fmt.Printf("   Duration:        %v\n", totalTime.Round(time.Second))
	fmt.Printf("   Total Requests:  %s\n", formatNumber(int(total)))
	fmt.Printf("   Successful:      %s (%.2f%%)\n", formatNumber(int(success)), float64(success)/float64(total)*100)
	fmt.Printf("   Failed:          %s (%.2f%%)\n", formatNumber(int(failed)), float64(failed)/float64(total)*100)
	fmt.Printf("   Throughput:      %.2f req/sec\n", float64(total)/totalTime.Seconds())

	fmt.Printf("\n🔴 ERROR BREAKDOWN\n")
	fmt.Printf("   Timeouts:        %d\n", metrics.Errors.Timeout)
	fmt.Printf("   Conn Refused:    %d\n", metrics.Errors.ConnRefused)
	fmt.Printf("   Bad Status:      %d\n", metrics.Errors.BadStatus)
	fmt.Printf("   Other:           %d\n", metrics.Errors.Other)

	fmt.Printf("\n� PRER-ENDPOINT BREAKDOWN\n")
	fmt.Println("   " + strings.Repeat("-", 58))
	fmt.Printf("   %-12s %10s %10s %10s %10s\n", "Endpoint", "Success", "Failed", "Avg(ms)", "Peak(ms)")
	fmt.Println("   " + strings.Repeat("-", 58))

	printEndpoint("LOGIN", metrics.LoginSuccess, metrics.LoginFail,
		avgLatency(metrics.LoginLatency, metrics.LoginLatencyCount), float64(metrics.LoginPeak)/1e6)
	printEndpoint("HOME", metrics.HomeSuccess, metrics.HomeFail,
		avgLatency(metrics.HomeLatency, metrics.HomeLatencyCount), float64(metrics.HomePeak)/1e6)
	printEndpoint("ADD_TASK", metrics.AddTaskSuccess, metrics.AddTaskFail,
		avgLatency(metrics.AddTaskLatency, metrics.AddTaskLatencyCount), float64(metrics.AddTaskPeak)/1e6)
	printEndpoint("TOGGLE", metrics.ToggleSuccess, metrics.ToggleFail,
		avgLatency(metrics.ToggleLatency, metrics.ToggleLatencyCount), float64(metrics.TogglePeak)/1e6)
	printEndpoint("DELETE", metrics.DeleteSuccess, metrics.DeleteFail,
		avgLatency(metrics.DeleteLatency, metrics.DeleteLatencyCount), float64(metrics.DeletePeak)/1e6)
	printEndpoint("LOGOUT", metrics.LogoutSuccess, metrics.LogoutFail,
		avgLatency(metrics.LogoutLatency, metrics.LogoutLatencyCount), float64(metrics.LogoutPeak)/1e6)

	fmt.Println("   " + strings.Repeat("-", 58))

	errorRate := float64(failed) / float64(total) * 100
	avgOverall := (avgLatency(metrics.LoginLatency, metrics.LoginLatencyCount) +
		avgLatency(metrics.HomeLatency, metrics.HomeLatencyCount) +
		avgLatency(metrics.AddTaskLatency, metrics.AddTaskLatencyCount)) / 3

	fmt.Printf("\n🏆 PERFORMANCE GRADE\n")
	if errorRate < 1 && avgOverall < 100 {
		fmt.Println("   ⭐⭐⭐⭐⭐ EXCELLENT - Production ready!")
	} else if errorRate < 5 && avgOverall < 500 {
		fmt.Println("   ⭐⭐⭐⭐ GOOD - Minor optimizations recommended")
	} else if errorRate < 10 && avgOverall < 1000 {
		fmt.Println("   ⭐⭐⭐ ACCEPTABLE - Needs optimization")
	} else if errorRate < 20 {
		fmt.Println("   ⭐⭐ POOR - Significant issues detected")
	} else {
		fmt.Println("   ⭐ CRITICAL - Major performance problems")
	}

	fmt.Printf("\n💡 DIAGNOSIS\n")
	if metrics.Errors.ConnRefused > 0 {
		fmt.Println("   • Connection refused = Server crashed or not running")
		fmt.Println("     Check server logs for errors")
	}
	if metrics.Errors.Timeout > int64(total)/10 {
		fmt.Println("   • High timeouts = DB connection pool exhausted or slow queries")
		fmt.Println("     Neon free tier: ~100 connections max")
	}
	if metrics.Errors.BadStatus > int64(total)/10 {
		fmt.Println("   • Bad status codes = Users don't exist in DB")
		fmt.Println("     Run: go run ./cmd/seed/main.go")
	}
	if avgLatency(metrics.LoginLatency, metrics.LoginLatencyCount) > 500 {
		fmt.Println("   • Slow login = bcrypt CPU-bound + network latency to Neon")
	}
	if errorRate < 5 {
		fmt.Println("   • App is performing well for Neon free tier!")
	}

	fmt.Println(strings.Repeat("=", 65))
}

func printEndpoint(name string, success, fail int64, avgMs, peakMs float64) {
	fmt.Printf("   %-12s %10s %10s %10.2f %10.2f\n",
		name, formatNumber(int(success)), formatNumber(int(fail)), avgMs, peakMs)
}

func formatNumber(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	return fmt.Sprintf("%s,%03d", formatNumber(n/1000), n%1000)
}
