package verifier

import (
	"time"
)

type Deliverability string

const (
	Deliverable   Deliverability = "DELIVERABLE"
	Undeliverable Deliverability = "UNDELIVERABLE"
	Risky         Deliverability = "RISKY"
	Unknown       Deliverability = "UNKNOWN"
)

type Result struct {
	Email             string         `json:"email"`
	IsValidFormat     bool           `json:"is_valid_format"`
	IsDisposableEmail bool           `json:"is_disposable_email"`
	Deliverability    Deliverability `json:"deliverability"`
	QualityScore      float64        `json:"quality_score"`
	Provider          string         `json:"provider"`
	Timestamp         time.Time      `json:"timestamp"`
	Error             string         `json:"error,omitempty"`
}
