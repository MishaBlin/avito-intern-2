package models

import "time"

type PVZ struct {
	ID               string    `json:"id,omitempty"`
	RegistrationDate time.Time `json:"registrationDate,omitempty"`
	City             string    `json:"city"`
}
