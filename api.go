package hyperliquid

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/valyala/fastjson"
)

// Pool of parsers to avoid allocations
var parserPool = sync.Pool{
	New: func() interface{} {
		return &fastjson.Parser{}
	},
}

type APIResponse[T any] struct {
	Status string
	Data   T
	Type   string
	Err    string
	Ok     bool
}

func (r *APIResponse[T]) UnmarshalJSON(data []byte) error {
	fmt.Printf("Unmarshalling API response: %s\n", string(data))
	// Get parser from pool
	parser := parserPool.Get().(*fastjson.Parser)
	defer parserPool.Put(parser)

	parsed, err := parser.ParseBytes(data)
	if err != nil {
		return fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Get status
	r.Status = string(parsed.GetStringBytes("status"))
	r.Ok = r.Status == "ok"

	if !r.Ok {
		// When status is not "ok", "response" is usually a string error message
		r.Err = string(parsed.GetStringBytes("response"))
		return nil
	}

	// When status is "ok", "response" contains "type" and "data"
	r.Type = string(parsed.GetStringBytes("response", "type"))
	responseData := parsed.GetStringBytes("response", "data")

	if responseData == nil {
		return fmt.Errorf("missing response.data field in successful response")
	}

	// Use fastjson's built-in unmarshaling if possible, fallback to json.Unmarshal
	if err := json.Unmarshal(responseData, &r.Data); err != nil {
		return fmt.Errorf("failed to unmarshal response data: %w", err)
	}

	return nil
}

type MixedValue json.RawMessage

func (mv *MixedValue) UnmarshalJSON(data []byte) error {
	*mv = data
	return nil
}

func (mv MixedValue) MarshalJSON() ([]byte, error) {
	return mv, nil
}

func (mv *MixedValue) String() (string, bool) {
	var s string
	if err := json.Unmarshal(*mv, &s); err != nil {
		return "", false
	}
	return s, true
}

func (mv *MixedValue) Object() (map[string]any, bool) {
	var obj map[string]any
	if err := json.Unmarshal(*mv, &obj); err != nil {
		return nil, false
	}
	return obj, true
}

func (mv *MixedValue) Array() ([]json.RawMessage, bool) {
	var arr []json.RawMessage
	if err := json.Unmarshal(*mv, &arr); err != nil {
		return nil, false
	}
	return arr, true
}

func (mv *MixedValue) Parse(v any) error {
	return json.Unmarshal(*mv, v)
}

func (mv *MixedValue) Type() string {
	if mv == nil || len(*mv) == 0 {
		return "null"
	}

	first := (*mv)[0]

	switch first {
	case '"':
		return "string"
	case '{':
		return "object"
	case '[':
		return "array"
	case 't', 'f':
		return "boolean"
	case 'n':
		return "null"
	default:
		return "number"
	}
}

type MixedArray []MixedValue

func (ma *MixedArray) UnmarshalJSON(data []byte) error {
	var rawArr []MixedValue
	if err := json.Unmarshal(data, &rawArr); err != nil {
		return err
	}

	return nil
}
