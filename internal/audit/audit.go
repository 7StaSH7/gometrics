package audit

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
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
	FilePath       string
	file           *os.File
	writer         *bufio.Writer
	mu             sync.Mutex
	bufSize        int
	flushThreshold int
}

// NewFileAuditObserver creates a new FileAuditObserver.
func NewFileAuditObserver(filePath string) *FileAuditObserver {
	return &FileAuditObserver{
		FilePath:       filePath,
		bufSize:        4096,
		flushThreshold: 2048,
	}
}

// ensureWriter lazily opens the file and creates a buffered writer.
func (f *FileAuditObserver) ensureWriter() error {
	if f.writer != nil {
		return nil
	}
	if f.file == nil {
		file, err := os.OpenFile(f.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		f.file = file
	}
	f.writer = bufio.NewWriterSize(f.file, f.bufSize)
	return nil
}

// Notify writes the audit event to the file (buffered).
func (f *FileAuditObserver) Notify(ctx context.Context, event AuditEvent) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.ensureWriter(); err != nil {
		return fmt.Errorf("failed to prepare audit file: %w", err)
	}
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}
	if _, err := f.writer.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write audit event to file: %w", err)
	}
	if f.writer.Buffered() >= f.flushThreshold {
		if err := f.writer.Flush(); err != nil {
			return fmt.Errorf("failed to flush audit buffer: %w", err)
		}
	}
	return nil
}

// Close flushes and closes the underlying file.
func (f *FileAuditObserver) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.writer != nil {
		if err := f.writer.Flush(); err != nil {
			return err
		}
	}
	if f.file != nil {
		if err := f.file.Close(); err != nil {
			return err
		}
		f.file = nil
		f.writer = nil
	}
	return nil
}

// HTTPAuditObserver implements AuditObserver for HTTP-based logging.
type HTTPAuditObserver struct {
	URL    string
	Client *http.Client
}

// NewHTTPAuditObserver creates a new HTTPAuditObserver.
func NewHTTPAuditObserver(url string) *HTTPAuditObserver {
	return &HTTPAuditObserver{
		URL:    url,
		Client: &http.Client{Timeout: 10 * time.Second},
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

	resp, err := h.Client.Do(req)
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

// MultiError aggregates multiple errors.
type MultiError struct {
	Errors []error
}

func (m *MultiError) Error() string {
	if len(m.Errors) == 0 {
		return ""
	}
	parts := make([]string, 0, len(m.Errors))
	for _, e := range m.Errors {
		if e != nil {
			parts = append(parts, e.Error())
		}
	}
	return strings.Join(parts, "; ")
}

// NotifyAll notifies all attached observers and returns aggregated error if any.
func (a *AuditSubject) NotifyAll(ctx context.Context, event AuditEvent) error {
	var errs []error
	for _, observer := range a.observers {
		if err := observer.Notify(ctx, event); err != nil {
			logger.Log.Warn("Failed to notify audit observer", zap.Error(err))
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return &MultiError{Errors: errs}
	}
	return nil
}

// CreateAuditEvent creates a new AuditEvent.
func CreateAuditEvent(metrics []string, ipAddress string) AuditEvent {
	return AuditEvent{
		Timestamp: time.Now().Unix(),
		Metrics:   metrics,
		IPAddress: ipAddress,
	}
}
