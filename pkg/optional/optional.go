package optional

import "encoding/json"

type Optional[T any] struct {
	value    T
	hasValue bool
}

func New[T any](value T) Optional[T] {
	return Optional[T]{
		value:    value,
		hasValue: true,
	}
}

func FromPtr[T any](value *T) Optional[T] {
	if value == nil {
		return Optional[T]{
			value:    *new(T),
			hasValue: false,
		}
	}
	return Optional[T]{
		value:    *value,
		hasValue: true,
	}
}

func Empty[T any]() Optional[T] {
	return Optional[T]{}
}

func (o Optional[T]) IsEmpty() bool {
	return !o.hasValue
}

func (o Optional[T]) Get() (T, bool) {
	return o.value, o.hasValue
}

// ToPtr creates a new copy of T
func (o Optional[T]) ToPtr() *T {
	if !o.hasValue {
		return nil
	}
	v := o.value
	return &v
}

func (o Optional[T]) Or(value T) T {
	if !o.hasValue {
		return value
	}
	return o.value
}

func (o *Optional[T]) Set(value T) {
	o.hasValue = true
	o.value = value
}

func (o *Optional[T]) Unset() {
	o.hasValue = false
	o.value = *new(T)
}

// MarshalJSON implements json.Marshaler.
// Empty optionals are encoded as JSON null.
func (o Optional[T]) MarshalJSON() ([]byte, error) {
	if !o.hasValue {
		return []byte("null"), nil
	}
	return json.Marshal(o.value)
}

// UnmarshalJSON implements json.Unmarshaler.
// JSON null unsets the optional; otherwise it parses into T and sets it.
func (o *Optional[T]) UnmarshalJSON(data []byte) error {
	// Must handle nil receiver defensively
	if o == nil {
		return nil
	}

	// Fast-path for null (also handles "null" with surrounding whitespace via Valid check below)
	if len(data) == 4 && string(data) == "null" {
		o.Unset()
		return nil
	}

	// More robust null handling (whitespace, etc.)
	trimmed := make([]byte, 0, len(data))
	for _, b := range data {
		if b != ' ' && b != '\n' && b != '\r' && b != '\t' {
			trimmed = append(trimmed, b)
		}
	}
	if string(trimmed) == "null" {
		o.Unset()
		return nil
	}

	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	o.Set(v)
	return nil
}
