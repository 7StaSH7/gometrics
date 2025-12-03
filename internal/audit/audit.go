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

type AuditEvent struct {
	Timestamp int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}

type AuditObserver interface {
	Notify(ctx context.Context, event AuditEvent) error
}

type FileAuditObserver struct {
	FilePath string
}

func NewFileAuditObserver(filePath string) *FileAuditObserver {
	return &FileAuditObserver{
		FilePath: filePath,
	}
}

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

type HTTPAuditObserver struct {
	URL string
}

func NewHTTPAuditObserver(url string) *HTTPAuditObserver {
	return &HTTPAuditObserver{
		URL: url,
	}
}

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

type AuditSubject struct {
	observers []AuditObserver
}

func NewAuditSubject() *AuditSubject {
	return &AuditSubject{
		observers: make([]AuditObserver, 0),
	}
}

func (a *AuditSubject) Attach(observer AuditObserver) {
	a.observers = append(a.observers, observer)
}

func (a *AuditSubject) NotifyAll(ctx context.Context, event AuditEvent) {
	for _, observer := range a.observers {
		if err := observer.Notify(ctx, event); err != nil {
			logger.Log.Warn("Failed to notify audit observer", zap.Error(err))
		}
	}
}

func CreateAuditEvent(metrics []string, ipAddress string) AuditEvent {
	return AuditEvent{
		Timestamp: time.Now().Unix(),
		Metrics:   metrics,
		IPAddress: ipAddress,
	}
}
