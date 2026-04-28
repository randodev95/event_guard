package normalization

import (
	"testing"
)

func TestNormalize_Segment_EventName(t *testing.T) {
	payload := []byte(`{
		"event": "Order Completed",
		"userId": "user_123",
		"properties": {
			"total": 50.00
		}
	}`)

	normalized, err := Normalize(payload)
	if err != nil {
		t.Fatalf("Normalize failed: %v", err)
	}

	if normalized.Event != "Order Completed" {
		t.Errorf("Expected event 'Order Completed', got '%s'", normalized.Event)
	}

	if normalized.Identity["userId"] != "user_123" {
		t.Errorf("Expected userId 'user_123', got '%s'", normalized.Identity["userId"])
	}

	if total, ok := normalized.Properties["total"].(float64); !ok || total != 50.00 {
		t.Errorf("Expected property total 50.00, got %v", normalized.Properties["total"])
	}
}

func TestNormalize_AnonymousID(t *testing.T) {
	payload := []byte(`{
		"event": "Page Viewed",
		"anonymousId": "anon_456",
		"properties": {}
	}`)

	normalized, err := Normalize(payload)
	if err != nil {
		t.Fatalf("Normalize failed: %v", err)
	}

	if normalized.Identity["userId"] != "anon_456" {
		t.Errorf("Expected userId 'anon_456' (from anonymousId), got '%s'", normalized.Identity["userId"])
	}
}
