package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"
)

// FileSink writes rejected events to a local file. Good for local dev.
type FileSink struct {
	mu   sync.Mutex
	path string
}

func NewFileSink(path string) *FileSink {
	return &FileSink{path: path}
}

func (s *FileSink) Close() error {
	return nil // File is opened per-push in current impl. Good for stability, bad for throughput.
}

func (s *FileSink) Push(payload []byte, errors []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	f, err := os.OpenFile(s.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	
	entry := map[string]interface{}{
		"payload": string(payload),
		"errors":  errors,
	}
	data, _ := json.Marshal(entry)
	f.Write(data)
	f.Write([]byte("\n"))
	return nil
}

// WebhookSink pushes rejected events to a remote HTTP endpoint (e.g., Kinesis proxy, Kafka REST).
type WebhookSink struct {
	URL    string
	client *http.Client
}

func NewWebhookSink(url string) *WebhookSink {
	return &WebhookSink{
		URL: url,
		client: &http.Client{
			Timeout: 2 * time.Second,
		},
	}
}

func (s *WebhookSink) Close() error {
	s.client.CloseIdleConnections()
	return nil
}

func (s *WebhookSink) Push(payload []byte, errors []string) error {
	entry := map[string]interface{}{
		"payload": string(payload),
		"errors":  errors,
		"source":  "event_guard_proxy",
	}
	data, _ := json.Marshal(entry)
	
	resp, err := s.client.Post(s.URL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("webhook post failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	return nil
}

// AsyncSink wraps another sink with a buffer and worker pool.
type AsyncSink struct {
	inner   Sink
	queue   chan asyncMsg
	workers int
}

type asyncMsg struct {
	payload []byte
	errors  []string
}

func NewAsyncSink(inner Sink, bufferSize, workers int) *AsyncSink {
	s := &AsyncSink{
		inner:   inner,
		queue:   make(chan asyncMsg, bufferSize),
		workers: workers,
	}
	for i := 0; i < workers; i++ {
		go s.run()
	}
	return s
}

func (s *AsyncSink) run() {
	for msg := range s.queue {
		_ = s.inner.Push(msg.payload, msg.errors)
	}
}

func (s *AsyncSink) Close() error {
	close(s.queue)
	return s.inner.Close()
}

func (s *AsyncSink) Push(payload []byte, errors []string) error {
	select {
	case s.queue <- asyncMsg{payload: payload, errors: errors}:
		return nil
	default:
		slog.Error("DLQ buffer full, dropping event", "workers", s.workers)
		return fmt.Errorf("async sink buffer full, event dropped")
	}
}
