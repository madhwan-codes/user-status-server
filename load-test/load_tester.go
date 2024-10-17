package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

const (
	baseURL          = "http://localhost:8080"
	numUsers         = 10000
	statusBatchSize  = 50
	heartbeatPercent = 80
	maxConcurrency   = 1000
	testDuration     = 5 * time.Second
	errorThreshold   = 0.01 // 1% error rate
	latencyThreshold = 100 * time.Millisecond
)

type Metrics struct {
	TotalRequests   int
	SuccessfulCalls int
	FailedCalls     int
	TotalDuration   time.Duration
	AverageDuration time.Duration
	MinDuration     time.Duration
	MaxDuration     time.Duration
	RequestsPerSec  float64
}

func main() {
	users := generateUsers(numUsers)
	maxUsers := findMaxUsers(users)
	fmt.Printf("Maximum number of live users that can be handled in 1 second: %d\n", maxUsers)
}

func generateUsers(n int) []string {
	users := make([]string, n)
	for i := 0; i < n; i++ {
		users[i] = fmt.Sprintf("user%d", i+1)
	}
	return users
}

func findMaxUsers(users []string) int {
	minUsers, maxUsers := 100, len(users)
	for minUsers < maxUsers {
		testUsers := (minUsers + maxUsers + 1) / 2
		metrics := runLoadTest(users, testUsers)

		errorRate := float64(metrics.FailedCalls) / float64(metrics.TotalRequests)
		if errorRate > errorThreshold || metrics.AverageDuration > latencyThreshold {
			maxUsers = testUsers - 1
		} else {
			minUsers = testUsers
		}
	}
	return minUsers
}

func runLoadTest(users []string, targetRPS int) Metrics {
	results := make(chan time.Duration, targetRPS*int(testDuration.Seconds()))
	errors := make(chan error, targetRPS*int(testDuration.Seconds()))

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxConcurrency)

	start := time.Now()
	ticker := time.NewTicker(time.Second / time.Duration(targetRPS))
	defer ticker.Stop()

	for time.Since(start) < testDuration {
		<-ticker.C
		wg.Add(1)
		go func() {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			var err error
			var duration time.Duration

			if rand.Intn(100) < heartbeatPercent {
				duration, err = sendHeartbeat(users[rand.Intn(len(users))])
			} else {
				duration, err = sendStatus(users)
			}

			if err != nil {
				errors <- err
			} else {
				results <- duration
			}
		}()
	}

	wg.Wait()
	close(results)
	close(errors)

	return calculateMetrics(results, errors, time.Since(start))
}

func sendHeartbeat(userID string) (time.Duration, error) {
	payload := map[string]string{"userId": userID}
	jsonPayload, _ := json.Marshal(payload)

	start := time.Now()
	resp, err := http.Post(baseURL+"/heartbeat", "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("heartbeat failed with status: %d", resp.StatusCode)
	}

	return time.Since(start), nil
}

func sendStatus(users []string) (time.Duration, error) {
	batch := make([]string, statusBatchSize)
	for i := 0; i < statusBatchSize; i++ {
		batch[i] = users[rand.Intn(len(users))]
	}

	payload := map[string][]string{"userIds": batch}
	jsonPayload, _ := json.Marshal(payload)

	start := time.Now()
	resp, err := http.Post(baseURL+"/status", "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("status check failed with status: %d", resp.StatusCode)
	}

	return time.Since(start), nil
}

func calculateMetrics(results <-chan time.Duration, errors <-chan error, totalDuration time.Duration) Metrics {
	var metrics Metrics
	var totalLatency time.Duration
	metrics.MinDuration = time.Duration(1<<63 - 1)

	for duration := range results {
		metrics.SuccessfulCalls++
		totalLatency += duration
		if duration < metrics.MinDuration {
			metrics.MinDuration = duration
		}
		if duration > metrics.MaxDuration {
			metrics.MaxDuration = duration
		}
	}

	metrics.FailedCalls = len(errors)
	metrics.TotalRequests = metrics.SuccessfulCalls + metrics.FailedCalls
	metrics.TotalDuration = totalDuration
	if metrics.SuccessfulCalls > 0 {
		metrics.AverageDuration = totalLatency / time.Duration(metrics.SuccessfulCalls)
	}
	metrics.RequestsPerSec = float64(metrics.TotalRequests) / totalDuration.Seconds()

	return metrics
}
