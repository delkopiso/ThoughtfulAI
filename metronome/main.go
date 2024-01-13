package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	dbSource           = "db.sqlite"
	serverListenerAddr = ":8081"

	requestTSFormat = "2006-01-02T15:04:05Z"
	dbTSFormat      = "2006-01-02 15:04:05.999-07"
)

type Usage struct {
	StartTimestamp string `json:"startTimestamp"`
	EndTimestamp   string `json:"endTimestamp"`
	EventCount     int    `json:"eventCount"`
}

type Event struct {
	CustomerID    string
	EventType     string
	TransactionID string
	Timestamp     time.Time
}

func main() {
	http.HandleFunc("/usage", handleUsage)
	log.Printf("Listening on %s", serverListenerAddr)
	if err := http.ListenAndServe(serverListenerAddr, nil); err != nil {
		log.Fatalf("failed to start server: %s", err)
	}
}

func countEvents(events []Event, customerID, startTime, endTime string) ([]Usage, error) {
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.Before(events[j].Timestamp)
	})

	var usage []Usage

	start, err := time.Parse(requestTSFormat, startTime)
	if err != nil {
		return nil, fmt.Errorf("failed to parse startTime: %s", err)
	}
	end, err := time.Parse(requestTSFormat, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to parse endTime: %s", err)
	}

	intervalStart := start.Hour()
	eventCount := 0

	matchingEvents := make([]Event, 0, len(events))
	for _, event := range events {
		if event.CustomerID == customerID {
			if event.Timestamp.Sub(start) >= 0 && event.Timestamp.Before(end) {
				matchingEvents = append(matchingEvents, event)
			}
		}
	}

	for index := range matchingEvents {
		event := matchingEvents[index]
		eventCount += 1

		// fmt.Printf("event %s - %s: %s\n", event.TransactionID, event.CustomerID, event.Timestamp)
		shouldComplete := index > 0 && (event.Timestamp.Hour() > matchingEvents[index-1].Timestamp.Hour() || index == len(matchingEvents)-1) &&
			matchingEvents[index-1].Timestamp.Hour() >= start.Hour()
		if shouldComplete {
			fmt.Println("start", start)
			fmt.Println("end", end)
			fmt.Println("intervalStart", intervalStart)
			fmt.Println("event.Timestamp", event.Timestamp)
			fmt.Println("events[index-1].Timestamp", matchingEvents[index-1].Timestamp)
			fmt.Println("shouldComplete", shouldComplete)
			intervalStart = matchingEvents[index-1].Timestamp.Hour()
			// eventCount = len(final)
			// final = final[:0]
			windowStart := matchingEvents[index-1].Timestamp.Truncate(time.Hour)
			windowEnd := windowStart.Add(time.Hour)
			usage = append(usage, Usage{
				StartTimestamp: windowStart.Format(requestTSFormat),
				EndTimestamp:   windowEnd.Format(requestTSFormat),
				EventCount:     eventCount,
			})
			eventCount = 0
		}
	}
	fmt.Println("events total", len(matchingEvents))
	_ = intervalStart

	return usage, nil
}

func handleUsage(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", dbSource)
	if err != nil {
		log.Printf("failed to open sqlite: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM usage")
	if err != nil {
		log.Printf("failed to query sqlite: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var event Event
		var timestamp string
		if err := rows.Scan(&event.CustomerID, &event.EventType, &event.TransactionID, &timestamp); err != nil {
			log.Printf("failed to scan data: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		event.Timestamp, err = time.Parse(dbTSFormat, timestamp)
		if err != nil {
			log.Printf("failed to parse timestamp: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		events = append(events, event)
	}

	var request = struct {
		CustomerID string `json:"customer_id"`
		StartTime  string `json:"start_timestamp"`
		EndTime    string `json:"end_timestamp"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Printf("failed to decode request: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	usage, err := countEvents(events, request.CustomerID, request.StartTime, request.EndTime)
	if err != nil {
		log.Printf("failed to count events: %s", err)
		http.Error(w, "", 500)
		return
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(usage); err != nil {
		log.Printf("failed to write response: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
