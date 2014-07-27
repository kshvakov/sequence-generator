package restapi

import (
	"encoding/json"
	"runtime"
	"time"
)

type Stat struct {
	requestCount    map[string]int
	startTime       time.Time
	lastRequestTime time.Time
}

func (s *Stat) add(commandName string) {

	s.requestCount[commandName]++

	s.lastRequestTime = time.Now()
}

func (s *Stat) getStat() (string, error) {

	response := statResponse{
		RequestCount:    s.requestCount,
		StartTime:       s.startTime.Format("2006-01-02 15:04:05"),
		LastRequestTime: s.lastRequestTime.Format("2006-01-02 15:04:05"),
		Len:             sequenceGenerator.Len(),
		NumGoroutine:    runtime.NumGoroutine(),
	}

	result, err := json.Marshal(response)

	if err != nil {

		return "", err
	}

	return string(result), nil

}

func NewStat() *Stat {
	return &Stat{
		requestCount: make(map[string]int),
		startTime:    time.Now(),
	}
}
