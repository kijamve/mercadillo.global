package H

// Collection is a generic type that wraps a slice of items.
type Collection[T any] struct {
	items []T
}

// NewCollection creates a new Collection.
func NewCollection[T any](items []T) *Collection[T] {
	return &Collection[T]{items: items}
}
func (c *Collection[T]) Count() int {
	return len(c.items)
}

func (c *Collection[T]) ToArray() []T {
	return c.items
}

// Map applies a function to each element in the collection and returns a new collection.
func (c *Collection[T]) Map(fn func(T) T) *Collection[T] {
	result := make([]T, len(c.items))
	for i, item := range c.items {
		result[i] = fn(item)
	}
	return NewCollection(result)
}

// Pluck extracts a specific field from each element in the collection (assuming the field is of type T).
func (c *Collection[T]) Pluck(fn func(T) any) []any {
	result := make([]any, len(c.items))
	for i, item := range c.items {
		result[i] = fn(item)
	}
	return result
}
func (c *Collection[T]) Chunk(size int) []*Collection[T] {
	if size <= 0 {
		return nil
	}

	var chunks []*Collection[T]
	for i := 0; i < len(c.items); i += size {
		end := i + size
		if end > len(c.items) {
			end = len(c.items)
		}
		chunks = append(chunks, NewCollection(c.items[i:end]))
	}
	return chunks
}

// Each iterates over each element in the collection and applies the given function.
func (c *Collection[T]) Each(fn func(T)) {
	for _, item := range c.items {
		fn(item)
	}
}

// Filter returns a new collection with elements that satisfy the predicate function.
func (c *Collection[T]) Filter(predicate func(T) bool) *Collection[T] {
	var result []T
	for _, item := range c.items {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return NewCollection(result)
}

// Reject returns a new collection with elements that do not satisfy the predicate function.
func (c *Collection[T]) Reject(predicate func(T) bool) *Collection[T] {
	var result []T
	for _, item := range c.items {
		if !predicate(item) {
			result = append(result, item)
		}
	}
	return NewCollection(result)
}

// Merge merges the current collection with another collection.
func (c *Collection[T]) Merge(other *Collection[T]) *Collection[T] {
	result := append(c.items, other.items...)
	return NewCollection(result)
}

// First returns the first element of the collection or nil if empty.
func (c *Collection[T]) First() *T {
	if len(c.items) > 0 {
		return &c.items[0]
	}
	return nil
}

// Last returns the last element of the collection or nil if empty.
func (c *Collection[T]) Last() *T {
	if len(c.items) > 0 {
		return &c.items[len(c.items)-1]
	}
	return nil
}

// IsEmpty checks if the collection is empty.
func (c *Collection[T]) IsEmpty() bool {
	return len(c.items) == 0
}

func (c *Collection[T]) Add(item T) {
	c.items = append(c.items, item)
}

func (c *Collection[T]) Contains(predicate func(T) bool) bool {
	for _, item := range c.items {
		if predicate(item) {
			return true
		}
	}
	return false
}
