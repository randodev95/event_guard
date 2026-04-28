package normalization

import (
	"github.com/tidwall/gjson"
)

type NormalizedEvent struct {
	Event      string                 `json:"event"`
	Properties map[string]interface{} `json:"properties"`
	UserID     string                 `json:"user_id"`
	Provider   string                 `json:"provider"`
}

func Normalize(data []byte) (*NormalizedEvent, error) {
	res := gjson.ParseBytes(data)
	
	event := res.Get("event").String()
	userID := res.Get("userId").String()
	if userID == "" {
		userID = res.Get("anonymousId").String()
	}
	properties, ok := res.Get("properties").Value().(map[string]interface{})
	if !ok {
		properties = make(map[string]interface{})
	}

	return &NormalizedEvent{
		Event:      event,
		UserID:     userID,
		Properties: properties,
	}, nil
}
