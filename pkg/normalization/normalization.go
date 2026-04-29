package normalization

import (
	"github.com/tidwall/gjson"
)

// NormalizedEvent is the canonical representation used by the validator.
// It is agnostic to the input format (Segment, Snowplow, Custom).
type NormalizedEvent struct {
	Event      string                 `json:"event"`
	Identity   map[string]string      `json:"identity"` // e.g. {"userId": "...", "anonymousId": "..."}
	Properties map[string]interface{} `json:"properties"`
}

// Mapper handles the mapping from raw JSON to the canonical NormalizedEvent.
type Mapper struct {
	// IdentityPaths defines where to look for identity properties.
	// We support multiple paths (camelCase, snake_case, nested).
	IdentityPaths map[string][]string
}

// NewDefaultMapper initializes a Mapper with standard search paths for common tracking schemas.
func NewDefaultMapper() *Mapper {
	return &Mapper{
		IdentityPaths: map[string][]string{
			"userId":         {"userId", "user_id", "context.userId", "context.user_id"},
			"anonymousId":    {"anonymousId", "anonymous_id", "context.anonymousId"},
			"wallet_address": {"wallet_address", "walletAddress", "properties.wallet_address"},
		},
	}
}

// Map transforms raw JSON into a canonical NormalizedEvent.
func (m *Mapper) Map(data []byte) (*NormalizedEvent, error) {
	res := gjson.ParseBytes(data)

	event := m.firstPath(res, []string{"event", "event_name", "type"})

	identity := make(map[string]string)
	for key, paths := range m.IdentityPaths {
		if val := m.firstPath(res, paths); val != "" {
			identity[key] = val
		}
	}

	// Senior Pattern: Logical Fallback
	// If userId is missing but anonymousId exists, we treat anonymousId as the identity.
	if identity["userId"] == "" && identity["anonymousId"] != "" {
		identity["userId"] = identity["anonymousId"]
	}

	// 3. Properties extraction with Deep Flattening
	properties := make(map[string]interface{})
	
	// Flatten "properties" block
	if propsRes := res.Get("properties"); propsRes.Exists() && propsRes.IsObject() {
		val := propsRes.Value()
		if m, ok := val.(map[string]interface{}); ok {
			for k, v := range Flatten(m, "", 0) {
				properties[k] = v
			}
		}
	}
	
	// Flatten "context" block (Standard in Segment/Rudderstack)
	if ctxRes := res.Get("context"); ctxRes.Exists() && ctxRes.IsObject() {
		val := ctxRes.Value()
		if m, ok := val.(map[string]interface{}); ok {
			for k, v := range Flatten(m, "context", 0) {
				properties[k] = v
			}
		}
	}

	return &NormalizedEvent{
		Event:      event,
		Identity:   identity,
		Properties: properties,
	}, nil
}

func (m *Mapper) firstPath(res gjson.Result, paths []string) string {
	for _, p := range paths {
		if val := res.Get(p).String(); val != "" {
			return val
		}
	}
	return ""
}

// Flatten flattens a nested map into dot-notation (e.g. {"a": {"b": 1}} -> {"a.b": 1}).
func Flatten(m map[string]interface{}, prefix string, depth int) map[string]interface{} {
	out := make(map[string]interface{})
	if depth > 10 {
		return out // Stop at 10 levels deep to prevent stack overflow
	}
	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}

		if inner, ok := v.(map[string]interface{}); ok {
			for ik, iv := range Flatten(inner, key, depth+1) {
				out[ik] = iv
			}
		} else {
			out[key] = v
		}
	}
	return out
}
