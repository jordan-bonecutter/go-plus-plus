package maps

import "option"

func Get[M ~map[K]V, K comparable, V any](m M, key K) option.Option[V] {
	value, ok := m[key]
	return option.From(value, ok)
}

func Take[M ~map[K]V, K comparable, V any](m M, key K) option.Option[V] {
	value := Get(m, key)
	delete(m, key)
	return value
}
