package mapping

import (
	"github.com/weitrue/kit/jsonx"
)

const jsonTagKey = "json"

var jsonUnmarshaler = NewUnmarshaler(jsonTagKey)

// UnmarshalJsonBytes unmarshals content into v.
func UnmarshalJsonBytes(content []byte, v any, opts ...UnmarshalOption) error {
	return unmarshalJsonBytes(content, v, getJsonUnmarshaler(opts...))
}

// UnmarshalJsonMap unmarshals content from m into v.
func UnmarshalJsonMap(m map[string]any, v any, opts ...UnmarshalOption) error {
	return getJsonUnmarshaler(opts...).Unmarshal(m, v)
}

func getJsonUnmarshaler(opts ...UnmarshalOption) *Unmarshaler {
	if len(opts) > 0 {
		return NewUnmarshaler(jsonTagKey, opts...)
	}

	return jsonUnmarshaler
}

func unmarshalJsonBytes(content []byte, v any, unmarshaler *Unmarshaler) error {
	var m any
	if err := jsonx.Unmarshal(content, &m); err != nil {
		return err
	}

	return unmarshaler.Unmarshal(m, v)
}
