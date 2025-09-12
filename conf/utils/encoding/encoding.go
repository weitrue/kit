package encoding

import (
	"bytes"
	"encoding/json"
	"github.com/pelletier/go-toml/v2"
	"github.com/weitrue/kit/utils/convert"
	"gopkg.in/yaml.v2"
)

// TomlToJson converts TOML data into its JSON representation.
func TomlToJson(data []byte) ([]byte, error) {
	var val any
	if err := toml.NewDecoder(bytes.NewReader(data)).Decode(&val); err != nil {
		return nil, err
	}

	return encodeToJSON(val)
}

// YamlToJson converts YAML data into its JSON representation.
func YamlToJson(data []byte) ([]byte, error) {
	var val any
	if err := yaml.Unmarshal(data, &val); err != nil {
		return nil, err
	}

	val = toStringKeyMap(val)

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(val); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// encodeToJSON encodes the given value into its JSON representation.
func encodeToJSON(val any) ([]byte, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(val); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func toStringKeyMap(v any) any {
	switch v := v.(type) {
	case []any:
		return convertSlice(v)
	case map[any]any:
		return convertKeyToString(v)
	case bool, string:
		return v
	case int, uint, int8, uint8, int16, uint16, int32, uint32, int64, uint64, float32, float64:
		return convertNumberToJsonNumber(v)
	default:
		return convert.Repr(v)
	}
}

// convertSlice processes slice items to ensure key compatibility.
func convertSlice(in []any) []any {
	res := make([]any, len(in))
	for i, v := range in {
		res[i] = toStringKeyMap(v)
	}
	return res
}

// convertKeyToString ensures all keys of the map are of type string.
func convertKeyToString(in map[any]any) map[string]any {
	res := make(map[string]any)
	for k, v := range in {
		res[convert.Repr(k)] = toStringKeyMap(v)
	}
	return res
}

// convertNumberToJsonNumber converts numbers into json.Number type for compatibility.
func convertNumberToJsonNumber(in any) json.Number {
	return json.Number(convert.Repr(in))
}
