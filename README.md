# go-optional

A simple, generic `Optional[T]` type for Go 1.18+ that provides a safe way to represent values that may or may not be present.

## Features

- **Generics**: Works with any type `T`.
- **JSON Support**: Implements `json.Marshaler` and `json.Unmarshaler`. Empty values are handled as `null`.
- **Pointer Integration**: Easily convert to/from pointers.
- **Fluent API**: Methods like `Or(defaultValue)` for easy value retrieval.

## Requirements

- Go 1.18 or later (requires generics).

## Installation

```bash
go get github.com/Palladium-blockchain/go-optional
```

## Usage

### Importing

```go
import "github.com/Palladium-blockchain/go-optional/pkg/optional"
```

### Basic Examples

```go
// Create a new Optional with a value
o := optional.New(42)
if !o.IsEmpty() {
    v, _ := o.Get()
    fmt.Println(v) // Output: 42
}

// Create an empty Optional
empty := optional.Empty[int]()
fmt.Println(empty.Or(100)) // Output: 100

// Create from a pointer (nil pointer results in an empty Optional)
var ptr *string
o2 := optional.FromPtr(ptr)
fmt.Println(o2.IsEmpty()) // Output: true

// Convert to a pointer (empty Optional results in a nil pointer)
p := o.ToPtr()
```

### JSON Support

`Optional[T]` is useful for handling JSON fields where you need to distinguish between a field being absent/null and a field having its zero value.

```go
type User struct {
    Email optional.Optional[string] `json:"email"`
}

// Unmarshaling "{"email": null}" or "{}" results in an empty Email field.
// Unmarshaling "{"email": "test@example.com"}" results in a non-empty Email field.

// Marshaling an empty Optional results in "null".
```

## API Reference

- `New[T](value T)`: Returns an `Optional[T]` containing the given value.
- `FromPtr[T](ptr *T)`: Returns an `Optional[T]` from a pointer. If the pointer is `nil`, the result is empty.
- `Empty[T]()`: Returns an empty `Optional[T]`.
- `(o Optional[T]) IsEmpty() bool`: Returns `true` if no value is present.
- `(o Optional[T]) Get() (T, bool)`: Returns the value and a boolean indicating if it's present.
- `(o Optional[T]) ToPtr() *T`: Returns a pointer to a copy of the value, or `nil` if empty.
- `(o Optional[T]) Or(defaultValue T) T`: Returns the value if present, otherwise returns `defaultValue`.
- `(o *Optional[T]) Set(value T)`: Sets the value and marks the optional as non-empty.
- `(o *Optional[T]) Unset()`: Removes the value and marks the optional as empty.

## Running Tests

To run the test suite, use the following command:

```bash
go test ./...
```
