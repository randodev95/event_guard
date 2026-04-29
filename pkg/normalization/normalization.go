package normalization

import (
	"fmt"
	"github.com/tidwall/gjson"
	"sync"
)

const MaxFlattenDepth = 100

// NormalizedEvent is the canonical representation used by the validator.
type NormalizedEvent struct {
	Event      string                 `json:"event"`
	Identity   map[string]string      `json:"identity"`
	Properties map[string]interface{} `json:"properties"`
}

var eventPool = sync.Pool{
	New: func() interface{} {
		return &NormalizedEvent{
			Identity:   make(map[string]string, 8),
			Properties: make(map[string]interface{}, 32),
		}
	},
}

// AcquireEvent gets a NormalizedEvent from the pool.
func AcquireEvent() *NormalizedEvent {
	return eventPool.Get().(*NormalizedEvent)
}

// ReleaseEvent returns a NormalizedEvent to the pool after clearing it.
func ReleaseEvent(e *NormalizedEvent) {
	e.Reset()
	eventPool.Put(e)
}

// Reset clears the event data for reuse.
func (e *NormalizedEvent) Reset() {
	e.Event = ""
	clear(e.Identity)
	clear(e.Properties)
}

// Mapper handles the mapping from raw JSON to the canonical NormalizedEvent.
type Mapper struct {
	IdentityPaths map[string][]string
	EventPaths    []string
}

func NewDefaultMapper() *Mapper {
	return &Mapper{
		IdentityPaths: map[string][]string{
			"userId":         {"userId", "user_id", "context.userId", "context.user_id"},
			"anonymousId":    {"anonymousId", "anonymous_id", "context.anonymousId"},
			"wallet_address": {"wallet_address", "walletAddress", "properties.wallet_address"},
		},
		EventPaths: []string{"event", "event_name", "type", "properties.event_name"},
	}
}

// Map transforms raw JSON into a canonical NormalizedEvent.
func (m *Mapper) Map(data []byte) (*NormalizedEvent, error) {
	res := gjson.ParseBytes(data)
	norm := AcquireEvent()

	norm.Event = m.firstPath(res, m.EventPaths)

	for key, paths := range m.IdentityPaths {
		if val := m.firstPath(res, paths); val != "" {
			norm.Identity[key] = val
		}
	}

	// Fallback: use anonymousId if userId is missing.
	if norm.Identity["userId"] == "" && norm.Identity["anonymousId"] != "" {
		norm.Identity["userId"] = norm.Identity["anonymousId"]
	}

	// Flatten "properties"
	if propsRes := res.Get("properties"); propsRes.Exists() && propsRes.IsObject() {
		if err := FlattenGJSON(propsRes, "properties", 0, norm.Properties); err != nil {
			ReleaseEvent(norm)
			return nil, err
		}
	}

	// Flatten "context"
	if ctxRes := res.Get("context"); ctxRes.Exists() && ctxRes.IsObject() {
		if err := FlattenGJSON(ctxRes, "context", 0, norm.Properties); err != nil {
			ReleaseEvent(norm)
			return nil, err
		}
	}

	return norm, nil
}

func (m *Mapper) firstPath(res gjson.Result, paths []string) string {
	for _, p := range paths {
		if val := res.Get(p).String(); val != "" {
			return val
		}
	}
	return ""
}

// FlattenGJSON flattens a gjson.Result into a destination map.
func FlattenGJSON(res gjson.Result, prefix string, depth int, out map[string]interface{}) error {
	if depth > MaxFlattenDepth {
		return fmt.Errorf("excessive json nesting: depth exceeds limit of %d", MaxFlattenDepth)
	}

	var err error
	res.ForEach(func(key, value gjson.Result) bool {
		k := key.String()
		fullKey := k
		if prefix != "" {
			fullKey = prefix + "." + k
		}

		if value.IsObject() {
			if e := FlattenGJSON(value, fullKey, depth+1, out); e != nil {
				err = e
				return false
			}
		} else {
			out[fullKey] = value.Value()
		}
		return true
	})

	return err
}
