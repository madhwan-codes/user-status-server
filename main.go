package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gomodule/redigo/redis"
)

var redisPool *redis.Pool

func main() {
	// Initialize Redis connection pool
	redisPool = &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "redis:6379")
		},
	}

	// Set up HTTP server
	http.HandleFunc("/heartbeat", heartbeatHandler)
	http.HandleFunc("/status", statusHandler)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func heartbeatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID string `json:"userId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	conn := redisPool.Get()
	defer conn.Close()

	_, err := conn.Do("SET", req.UserID, time.Now().Unix())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserIDs []string `json:"userIds"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	conn := redisPool.Get()
	defer conn.Close()

	// Prepare the arguments for MGET
	args := make([]interface{}, len(req.UserIDs))
	for i, userID := range req.UserIDs {
		args[i] = userID
	}

	// Execute MGET command
	values, err := redis.Values(conn.Do("MGET", args...))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	statuses := make(map[string]time.Time)
	for i, value := range values {
		if value == nil {
			// Key does not exist
			statuses[req.UserIDs[i]] = time.Time{}
		} else {
			lastActivityUnix, err := redis.Int64(value, nil)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			statuses[req.UserIDs[i]] = time.Unix(lastActivityUnix, 0)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(statuses); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
