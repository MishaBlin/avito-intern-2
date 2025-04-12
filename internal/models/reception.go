package models

import "time"

type Reception struct {
	ID       string    `json:"id,omitempty"`
	DateTime time.Time `json:"dateTime"`
	PvzID    string    `json:"pvzId"`
	Status   string    `json:"status"`
}
