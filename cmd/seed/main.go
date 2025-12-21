package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func loadEnv() {
	file, err := os.Open(".env")
	if err != nil {
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 && os.Getenv(strings.TrimSpace(parts[0])) == "" {
			os.Setenv(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}
}

const (
	numUsers    = 100_000
	minTasks    = 5
	maxTasks    = 10
	batchSize   = 5000
	defaultPass = "Test@123"
)

var (
	taskTemplates = []string{
		"Review pull request for %s",
		"Update documentation for %s",
		"Fix bug in %s module",
		"Write tests for %s",
		"Deploy %s to staging",
		"Meeting about %s",
		"Research %s implementation",
		"Refactor %s code",
		"Setup CI/CD for %s",
		"Code review %s changes",
		"Optimize %s performance",
		"Debug %s issue",
		"Plan %s sprint",
		"Design %s architecture",
		"Implement %s feature",
	}
	projects = []string{
		"auth", "api", "frontend", "backend", "database",
		"payments", "notifications", "search", "analytics", "dashboard",
	}
)

type Analytics struct {
	PasswordHashTime time.Duration
	UserGenTime      time.Duration
	TaskGenTime      time.Duration
	UserWriteTime    time.Duration
	TaskWriteTime    time.Duration
	TotalTime        time.Duration
	TotalUsers       int
	TotalTasks       int64
	PeakRSS          uint64
}

func (a *Analytics) Print() {
	fmt.Printf("\n📈 SEEDING ANALYTICS:\n")
	fmt.Printf("   Password hashing:    %v\n", a.PasswordHashTime)
	fmt.Printf("   User data gen:       %v\n", a.UserGenTime)
	fmt.Printf("   Task data gen:       %v\n", a.TaskGenTime)
	fmt.Printf("   User DB write:       %v\n", a.UserWriteTime)
	fmt.Printf("   Task DB write:       %v\n", a.TaskWriteTime)
	fmt.Printf("   TOTAL TIME:          %v\n", a.TotalTime)
	fmt.Printf("   USERS:               %s\n", formatNumber(a.TotalUsers))
	fmt.Printf("   TASKS:               %s\n", formatNumber(int(a.TotalTasks)))
	fmt.Printf("   THROUGHPUT:          %.0f rows/sec\n", float64(a.TotalUsers+int(a.TotalTasks))/a.TotalTime.Seconds())
}

func main() {
	loadEnv()
	
	numWorkers := runtime.NumCPU()
	start := time.Now()

	fmt.Printf("💥 SEEDING PostgreSQL: %s users...\n", formatNumber(numUsers))
	fmt.Printf("   CPU Cores: %d | Batch Size: %s\n", numWorkers, formatNumber(batchSize))

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("DATABASE_URL not set. Add it to .env file")
	}

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		log.Fatal(err)
	}
	config.MaxConns = 50

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	fmt.Println("✅ Connected to PostgreSQL")

	analytics := &Analytics{TotalUsers: numUsers}

	// Hash password once
	hashStart := time.Now()
	passHash, err := bcrypt.GenerateFromPassword([]byte(defaultPass), 8)
	if err != nil {
		log.Fatal(err)
	}
	analytics.PasswordHashTime = time.Since(hashStart)

	// Generate user data
	genUserStart := time.Now()
	userData := generateUsersParallel(numUsers, string(passHash), numWorkers)
	analytics.UserGenTime = time.Since(genUserStart)

	// Write users
	writeUserStart := time.Now()
	writeUsersInBatches(pool, userData)
	analytics.UserWriteTime = time.Since(writeUserStart)

	// Generate task data
	genTaskStart := time.Now()
	taskData := generateTasksParallel(numUsers, numWorkers)
	analytics.TaskGenTime = time.Since(genTaskStart)

	// Write tasks
	writeTaskStart := time.Now()
	writeTasksInBatches(pool, taskData)
	analytics.TaskWriteTime = time.Since(writeTaskStart)

	analytics.TotalTime = time.Since(start)
	analytics.TotalTasks = int64(len(taskData))

	fmt.Printf("\n✅ SEED COMPLETE!\n")
	analytics.Print()

	fmt.Println("\n📧 Sample logins:")
	fmt.Println("   user1@example.com / Test@123")
	fmt.Println("   user1000@example.com / Test@123")
	fmt.Println("   user100000@example.com / Test@123")
}

type userRow struct {
	email    string
	passHash string
}

func generateUsersParallel(count int, passHash string, numWorkers int) []userRow {
	chunkSize := count / numWorkers
	results := make([][]userRow, numWorkers)
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		start := i * chunkSize
		end := start + chunkSize
		if i == numWorkers-1 {
			end = count
		}

		go func(idx, s, e int) {
			defer wg.Done()
			local := make([]userRow, 0, e-s)
			for j := s; j < e; j++ {
				local = append(local, userRow{
					email:    fmt.Sprintf("user%d@example.com", j+1),
					passHash: passHash,
				})
			}
			results[idx] = local
		}(i, start, end)
	}
	wg.Wait()

	flat := make([]userRow, 0, count)
	for _, r := range results {
		flat = append(flat, r...)
	}
	return flat
}

func writeUsersInBatches(pool *pgxpool.Pool, users []userRow) {
	ctx := context.Background()

	for i := 0; i < len(users); i += batchSize {
		end := i + batchSize
		if end > len(users) {
			end = len(users)
		}
		batch := users[i:end]

		// Use COPY for fast bulk insert
		rows := make([][]interface{}, len(batch))
		for j, u := range batch {
			rows[j] = []interface{}{u.email, u.passHash}
		}

		src := pgxCopySource(rows)
		_, err := pool.CopyFrom(
			ctx,
			[]string{"users"},
			[]string{"email", "password_hash"},
			&src,
		)
		if err != nil {
			log.Printf("User batch %d failed: %v", i/batchSize, err)
		}

		if (i/batchSize+1)%10 == 0 {
			fmt.Printf("   ... %s/%s users\n", formatNumber(end), formatNumber(len(users)))
		}
	}
}

type taskRow struct {
	userID    int
	text      string
	completed bool
	createdAt time.Time
}

func generateTasksParallel(numUsers, numWorkers int) []taskRow {
	chunkSize := numUsers / numWorkers
	results := make([][]taskRow, numWorkers)
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		start := i * chunkSize
		end := start + chunkSize
		if i == numWorkers-1 {
			end = numUsers
		}

		go func(idx, s, e int) {
			defer wg.Done()
			randSrc := rand.New(rand.NewSource(time.Now().UnixNano() + int64(idx)))
			local := make([]taskRow, 0)

			for j := s; j < e; j++ {
				userID := j + 1
				numTasks := minTasks + randSrc.Intn(maxTasks-minTasks+1)

				for k := 0; k < numTasks; k++ {
					task := fmt.Sprintf(
						taskTemplates[randSrc.Intn(len(taskTemplates))],
						projects[randSrc.Intn(len(projects))],
					)
					local = append(local, taskRow{
						userID:    userID,
						text:      task,
						completed: randSrc.Float32() < 0.3,
						createdAt: time.Now().Add(-time.Duration(randSrc.Intn(30*24)) * time.Hour),
					})
				}
			}
			results[idx] = local
		}(i, start, end)
	}
	wg.Wait()

	total := 0
	for _, r := range results {
		total += len(r)
	}
	flat := make([]taskRow, 0, total)
	for _, r := range results {
		flat = append(flat, r...)
	}
	return flat
}

func writeTasksInBatches(pool *pgxpool.Pool, tasks []taskRow) {
	ctx := context.Background()

	for i := 0; i < len(tasks); i += batchSize {
		end := i + batchSize
		if end > len(tasks) {
			end = len(tasks)
		}
		batch := tasks[i:end]

		rows := make([][]interface{}, len(batch))
		for j, t := range batch {
			rows[j] = []interface{}{t.userID, t.text, t.completed, t.createdAt}
		}

		src := pgxCopySource(rows)
		_, err := pool.CopyFrom(
			ctx,
			[]string{"tasks"},
			[]string{"user_id", "text", "completed", "created_at"},
			&src,
		)
		if err != nil {
			log.Printf("Task batch %d failed: %v", i/batchSize, err)
		}

		if (i/batchSize+1)%20 == 0 {
			fmt.Printf("   ... %s/%s tasks\n", formatNumber(end), formatNumber(len(tasks)))
		}
	}
}

// pgxCopySource implements pgx.CopyFromSource
type pgxCopySource [][]interface{}

func (s pgxCopySource) Next() bool {
	return len(s) > 0
}

func (s *pgxCopySource) Values() ([]interface{}, error) {
	if len(*s) == 0 {
		return nil, nil
	}
	row := (*s)[0]
	*s = (*s)[1:]
	return row, nil
}

func (s pgxCopySource) Err() error {
	return nil
}

func formatNumber(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	return fmt.Sprintf("%s,%03d", formatNumber(n/1000), n%1000)
}
