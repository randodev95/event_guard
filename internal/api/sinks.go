package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

// FileSink writes rejected events to a local file. Good for local dev.
type FileSink struct {
	mu   sync.Mutex
	path string
	file *os.File
}

func NewFileSink(path string) *FileSink {
	return &FileSink{path: path}
}

func (s *FileSink) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.file != nil {
		err := s.file.Close()
		s.file = nil
		return err
	}
	return nil
}

func (s *FileSink) Push(payload []byte, errors []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.file == nil {
		f, err := os.OpenFile(s.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		s.file = f
	}

	entry := map[string]interface{}{
		"payload": json.RawMessage(payload),
		"errors":  errors,
		"ts":      time.Now().Unix(),
	}
	data, _ := json.Marshal(entry)
	s.file.Write(data)
	s.file.Write([]byte("\n"))
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
	// Blocking push implements backpressure.
	// Production systems should prefer this over dropping events silently.
	s.queue <- asyncMsg{payload: payload, errors: errors}
	return nil
}
