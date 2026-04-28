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

// Normalize is a convenience wrapper for backward compatibility.
// It uses the default mapper to transform raw JSON.
func Normalize(data []byte) (*NormalizedEvent, error) {
	return NewDefaultMapper().Map(data)
}

// Mapper handles the mapping from raw JSON to the canonical NormalizedEvent.
type Mapper struct {
	// IdentityPaths defines where to look for identity properties.
	// We support multiple paths (camelCase, snake_case, nested).
	IdentityPaths map[string][]string
}

func NewDefaultMapper() *Mapper {
	return &Mapper{
		IdentityPaths: map[string][]string{
			"userId":      {"userId", "user_id", "context.userId", "context.user_id"},
			"anonymousId": {"anonymousId", "anonymous_id", "context.anonymousId"},
			"wallet_address": {"wallet_address", "walletAddress", "properties.wallet_address"},
		},
	}
}

// Normalize transforms raw JSON into a canonical NormalizedEvent.
// It resolves casing and nesting issues by following defined search paths.
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

	// Properties extraction: fallback from "properties" to root if missing
	properties := make(map[string]interface{})
	propsRes := res.Get("properties")
	if propsRes.Exists() && propsRes.IsObject() {
		properties = propsRes.Value().(map[string]interface{})
	} else {
		// Fallback: If no "properties" block, we might take the root but exclude standard keys
		// However, for strict FAANG logic, we usually expect a clear separation.
		// For now, let's keep it safe.
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

// Flatten takes a nested map and flattens it into dot-notation.
// e.g. {"a": {"b": 1}} -> {"a.b": 1}
// This makes validation against complex structures much simpler.
func Flatten(m map[string]interface{}, prefix string) map[string]interface{} {
	out := make(map[string]interface{})
	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}

		if inner, ok := v.(map[string]interface{}); ok {
			for ik, iv := range Flatten(inner, key) {
				out[ik] = iv
			}
		} else {
			out[key] = v
		}
	}
	return out
}
