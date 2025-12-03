package audit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateAuditEvent(t *testing.T) {
	metrics := []string{"Alloc", "Frees"}
	ipAddress := "192.168.0.42"
	before := time.Now().Unix()

	event := CreateAuditEvent(metrics, ipAddress)

	assert.GreaterOrEqual(t, event.Timestamp, before)
	assert.LessOrEqual(t, event.Timestamp, time.Now().Unix())

	assert.Equal(t, metrics, event.Metrics)
	assert.Equal(t, ipAddress, event.IPAddress)
}

func TestFileAuditObserver_Notify(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "audit_test_*.log")
	assert.NoError(t, err)

	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	observer := NewFileAuditObserver(tmpFile.Name())

	metrics := []string{"Alloc", "Frees"}
	event := AuditEvent{
		Timestamp: 12345678,
		Metrics:   metrics,
		IPAddress: "192.168.0.42",
	}

	err = observer.Notify(context.Background(), event)
	assert.NoError(t, err)

	content, err := os.ReadFile(tmpFile.Name())
	assert.NoError(t, err)

	expectedJSON, err := json.Marshal(event)
	assert.NoError(t, err)

	expectedLine := string(expectedJSON) + "\n"

	assert.Equal(t, expectedLine, string(content))
}

func TestHTTPAuditObserver_Notify(t *testing.T) {
	var receivedEvent AuditEvent
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		err := json.NewDecoder(r.Body).Decode(&receivedEvent)
		assert.NoError(t, err)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	observer := NewHTTPAuditObserver(server.URL)

	metrics := []string{"Alloc", "Frees"}
	event := AuditEvent{
		Timestamp: 12345678,
		Metrics:   metrics,
		IPAddress: "192.168.0.42",
	}

	err := observer.Notify(context.Background(), event)
	assert.NoError(t, err)
	assert.Equal(t, event, receivedEvent)
}

func TestHTTPAuditObserver_Notify_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	observer := NewHTTPAuditObserver(server.URL)

	event := AuditEvent{
		Timestamp: 12345678,
		Metrics:   []string{"Alloc"},
		IPAddress: "192.168.0.42",
	}

	err := observer.Notify(context.Background(), event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audit server returned non-success status")
}

func TestAuditSubject_NotifyAll(t *testing.T) {
	mockObserver1 := &MockAuditObserver{}
	mockObserver2 := &MockAuditObserver{}

	subject := NewAuditSubject()
	subject.Attach(mockObserver1)
	subject.Attach(mockObserver2)

	event := AuditEvent{
		Timestamp: 12345678,
		Metrics:   []string{"Alloc", "Frees"},
		IPAddress: "192.168.0.42",
	}

	subject.NotifyAll(context.Background(), event)

	assert.Equal(t, 1, mockObserver1.NotifyCallCount)
	assert.Equal(t, event, mockObserver1.LastEvent)
	assert.Equal(t, 1, mockObserver2.NotifyCallCount)
	assert.Equal(t, event, mockObserver2.LastEvent)
}

type MockAuditObserver struct {
	NotifyCallCount int
	LastEvent       AuditEvent
	Err             error
}

func (m *MockAuditObserver) Notify(ctx context.Context, event AuditEvent) error {
	m.NotifyCallCount++
	m.LastEvent = event
	return m.Err
}
