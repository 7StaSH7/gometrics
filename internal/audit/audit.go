package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/7StaSH7/gometrics/internal/logger"
	"go.uber.org/zap"
)

// AuditEvent represents an audit event with timestamp, metrics, and IP address.
type AuditEvent struct {
	Timestamp int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}

// AuditObserver defines the interface for audit observers.
type AuditObserver interface {
	Notify(ctx context.Context, event AuditEvent) error
}

// FileAuditObserver implements AuditObserver for file-based logging.
type FileAuditObserver struct {
	FilePath string
}

// NewFileAuditObserver creates a new FileAuditObserver.
func NewFileAuditObserver(filePath string) *FileAuditObserver {
	return &FileAuditObserver{
		FilePath: filePath,
	}
}

// Notify writes the audit event to the file.
func (f *FileAuditObserver) Notify(ctx context.Context, event AuditEvent) error {
	file, err := os.OpenFile(f.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open audit file: %w", err)
	}
	defer file.Close()

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	_, err = file.Write(append(data, '\n'))
	if err != nil {
		return fmt.Errorf("failed to write audit event to file: %w", err)
	}

	return nil
}

// HTTPAuditObserver implements AuditObserver for HTTP-based logging.
type HTTPAuditObserver struct {
	URL string
}

// NewHTTPAuditObserver creates a new HTTPAuditObserver.
func NewHTTPAuditObserver(url string) *HTTPAuditObserver {
	return &HTTPAuditObserver{
		URL: url,
	}
}

// Notify sends the audit event via HTTP POST.
func (h *HTTPAuditObserver) Notify(ctx context.Context, event AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.URL, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send audit event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("audit server returned non-success status: %d", resp.StatusCode)
	}

	return nil
}

// AuditSubject manages a list of audit observers.
type AuditSubject struct {
	observers []AuditObserver
}

// NewAuditSubject creates a new AuditSubject.
func NewAuditSubject() *AuditSubject {
	return &AuditSubject{
		observers: make([]AuditObserver, 0),
	}
}

// Attach adds an observer to the subject.
func (a *AuditSubject) Attach(observer AuditObserver) {
	a.observers = append(a.observers, observer)
}

// NotifyAll notifies all attached observers.
func (a *AuditSubject) NotifyAll(ctx context.Context, event AuditEvent) {
	for _, observer := range a.observers {
		if err := observer.Notify(ctx, event); err != nil {
			logger.Log.Warn("Failed to notify audit observer", zap.Error(err))
		}
	}
}

// CreateAuditEvent creates a new AuditEvent.
func CreateAuditEvent(metrics []string, ipAddress string) AuditEvent {
	return AuditEvent{
		Timestamp: time.Now().Unix(),
		Metrics:   metrics,
		IPAddress: ipAddress,
	}
}
