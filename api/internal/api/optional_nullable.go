package api

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/danielgtaylor/huma/v2"
)

// OptionalNullable distinguishes omitted PATCH fields from explicit nulls.
type OptionalNullable[T any] struct {
	Sent  bool
	Value *T
}

func (o *OptionalNullable[T]) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &o.Value); err != nil {
		return fmt.Errorf("decode nullable field: %w", err)
	}
	o.Sent = true
	return nil
}

func (o OptionalNullable[T]) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(o.Value)
	if err != nil {
		return nil, fmt.Errorf("encode nullable field: %w", err)
	}
	return data, nil
}

func (o OptionalNullable[T]) IsZero() bool {
	return !o.Sent
}

func (o OptionalNullable[T]) Schema(r huma.Registry) *huma.Schema {
	schema := *r.Schema(reflect.TypeFor[T](), true, "")
	schema.Nullable = true
	return &schema
}
