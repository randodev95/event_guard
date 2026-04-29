package normalization

import (
	"fmt"
	"strings"
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
		if norm.Properties["properties.a.b.c"] != 1.0 {
			t.Errorf("Expected 1.0, got %v", norm.Properties["properties.a.b.c"])
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

	t.Run("Excessive Depth Error", func(t *testing.T) {
		// Create 101 levels of nesting
		var nesting strings.Builder
		nesting.WriteString(`{"event": "Login", "properties": {`)
		for i := 0; i < 101; i++ {
			fmt.Fprintf(&nesting, `"d%d":{`, i)
		}
		nesting.WriteString(`"val":1`)
		for i := 0; i < 101; i++ {
			nesting.WriteString(`}`)
		}
		nesting.WriteString(`}}`)
		
		payload := []byte(nesting.String())
		mapper := NewDefaultMapper()
		_, err := mapper.Map(payload)
		if err == nil {
			t.Error("Expected error for excessive depth (101 levels), got nil")
		}
	})
}

func BenchmarkMapper_Map(b *testing.B) {
	payload := []byte(`{
		"event": "Login",
		"userId": "u123",
		"anonymousId": "anon_456",
		"context": {
			"app": {
				"name": "EventGuard",
				"version": "1.0.0"
			},
			"device": {
				"type": "ios",
				"model": "iPhone 13"
			}
		},
		"properties": {
			"price": 29.99,
			"currency": "USD",
			"items": ["a", "b", "c"],
			"metadata": {
				"source": "organic",
				"campaign": "launch"
			}
		}
	}`)
	mapper := NewDefaultMapper()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = mapper.Map(payload)
	}
}
