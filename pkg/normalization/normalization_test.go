package normalization

import (
	"testing"
)

func TestMapper_Map(t *testing.T) {
	t.Run("Basic Mapping", func(t *testing.T) {
		payload := []byte(`{"event": "Login", "userId": "u1", "properties": {"platform": "ios"}}`)
		mapper := NewDefaultMapper()
		norm, err := mapper.Map(payload)
		if err != nil {
			t.Fatal(err)
		}
		if norm.Event != "Login" {
			t.Errorf("Expected Login, got %s", norm.Event)
		}
		if norm.Identity["userId"] != "u1" {
			t.Errorf("Expected userId u1, got %s", norm.Identity["userId"])
		}
	})

	t.Run("Deep Nesting", func(t *testing.T) {
		payload := []byte(`{"event": "Login", "properties": {"a": {"b": {"c": 1}}}}`)
		mapper := NewDefaultMapper()
		norm, _ := mapper.Map(payload)
		if norm.Properties["a.b.c"] != 1.0 {
			t.Errorf("Expected 1.0, got %v", norm.Properties["a.b.c"])
		}
	})

	t.Run("AnonymousID Mapping", func(t *testing.T) {
		payload := []byte(`{
			"event": "Page Viewed",
			"anonymousId": "anon_456",
			"properties": {}
		}`)
		mapper := NewDefaultMapper()
		norm, err := mapper.Map(payload)
		if err != nil {
			t.Fatal(err)
		}
		if norm.Identity["userId"] != "anon_456" {
			t.Errorf("Expected userId 'anon_456' (from anonymousId), got '%s'", norm.Identity["userId"])
		}
	})
}
