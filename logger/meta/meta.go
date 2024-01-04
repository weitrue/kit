package meta

import "go.uber.org/zap"

// Field key-value
type Field interface {
	Key() string
	Value() any
}

type field struct {
	key   string
	value any
}

func (m *field) Key() string {
	return m.key
}

func (m *field) Value() interface{} {
	return m.value
}

// NewField create meat
func NewField(key string, value any) Field {
	return &field{key: key, value: value}
}

// WrapZapMeta wrap meta to zap fields
func WrapZapMeta(err error, metas ...Field) (fields []zap.Field) {
	capacity := len(metas) + 1 // namespace meta
	if err != nil {
		capacity++
	}

	fields = make([]zap.Field, 0, capacity)
	if err != nil {
		fields = append(fields, zap.Error(err))
	}

	for _, m := range metas {
		fields = append(fields, zap.Any(m.Key(), m.Value()))
	}

	return
}
