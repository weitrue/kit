package conf

import (
	"github.com/weitrue/kit/utils/encoding"
	"github.com/weitrue/kit/utils/jsonx"
	"github.com/weitrue/kit/utils/mapping"
	"reflect"
)

// children and mapField should not be both filled.
// named fields and map cannot be bound to the same field name.
type fieldInfo struct {
	children map[string]*fieldInfo
	mapField *fieldInfo
}

// LoadFromJsonBytes loads config into v from content json bytes.
func LoadFromJsonBytes(content []byte, v any) error {
	info, err := buildFieldsInfo(reflect.TypeOf(v), "")
	if err != nil {
		return err
	}

	var m map[string]any
	if err = jsonx.Unmarshal(content, &m); err != nil {
		return err
	}

	lowerCaseKeyMap := toLowerCaseKeyMap(m, info)

	return mapping.UnmarshalJsonMap(lowerCaseKeyMap, v, mapping.WithCanonicalKeyFunc(toLowerCase))
}

// LoadFromYamlBytes loads config into v from content yaml bytes.
func LoadFromYamlBytes(content []byte, v any) error {
	b, err := encoding.YamlToJson(content)
	if err != nil {
		return err
	}

	return LoadFromJsonBytes(b, v)
}

// LoadFromTomlBytes loads config into v from content toml bytes.
func LoadFromTomlBytes(content []byte, v any) error {
	b, err := encoding.TomlToJson(content)
	if err != nil {
		return err
	}

	return LoadFromJsonBytes(b, v)
}

func toLowerCaseKeyMap(m map[string]any, info *fieldInfo) map[string]any {
	res := make(map[string]any)

	for k, v := range m {
		ti, ok := info.children[k]
		if ok {
			res[k] = toLowerCaseInterface(v, ti)
			continue
		}

		lk := toLowerCase(k)
		if ti, ok = info.children[lk]; ok {
			res[lk] = toLowerCaseInterface(v, ti)
		} else if info.mapField != nil {
			res[k] = toLowerCaseInterface(v, info.mapField)
		} else {
			res[k] = v
		}
	}

	return res
}

func toLowerCaseInterface(v any, info *fieldInfo) any {
	switch vv := v.(type) {
	case map[string]any:
		return toLowerCaseKeyMap(vv, info)
	case []any:
		var arr []any
		for _, vvv := range vv {
			arr = append(arr, toLowerCaseInterface(vvv, info))
		}
		return arr
	default:
		return v
	}
}
