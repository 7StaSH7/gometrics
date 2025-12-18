// Package model defines the data models for metrics.
package model

// Metric types constants
const (
	// Counter represents a counter metric type
	Counter = "counter"
	// Gauge represents a gauge metric type
	Gauge = "gauge"
)

// Metrics represents a metric data structure used for storing and transmitting
// metric information. It includes ID, type, delta for counters, value for gauges,
// and optional hash for integrity verification.
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}
