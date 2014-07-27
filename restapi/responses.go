package restapi

import (
	"encoding/json"
)

type statResponse struct {
	RequestCount    map[string]int `json:"request_count"`
	StartTime       string         `json:"start_time"`
	LastRequestTime string         `json:"last_request_time"`
	Len             int            `json:"len"`
	NumGoroutine    int            `json:"num_goroutine"`
}

type response struct {
	Key   string `json:"key,omitempty"`
	Value uint   `json:"value,omitempty"`
	Error string `json:"error,omitempty"`
}

func (r response) String() string {

	result, err := json.Marshal(r)

	if err != nil {
		result, _ = json.Marshal(response{Error: err.Error()})
	}

	return string(result)
}
