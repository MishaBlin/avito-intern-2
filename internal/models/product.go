package models

import "time"

type Product struct {
	ID          string    `json:"id,omitempty"`
	DateTime    time.Time `json:"dateTime"`
	Type        string    `json:"type"`
	ReceptionID string    `json:"receptionId"`
}
