package option

type Option[T any] union {
	None
	Some T
}

func From[T any](value T, ok bool) Option[T] {
	if ok {
		return Option.Some(value)
	}
	return Option[T].None()
}

func (o Option[T]) IsSome() bool {
	_, ok := o.(Some)
	return ok
}

func (o Option[T]) IsNone() bool { return !o.IsSome() }

func (o Option[T]) Get() (T, bool) {
	value, ok := o.(Some)
	return value, ok
}

func (o Option[T]) OrPanic() T {
	switch value := o.(variant) {
	case Some:
		return value
	case None:
		panic("option: OrPanic of None")
	}
}

func (o Option[T]) UnwrapOr(fallback T) T {
	switch value := o.(variant) {
	case Some:
		return value
	case None:
		return fallback
	}
}

func (o Option[T]) UnwrapOrElse(f func() T) T {
	switch value := o.(variant) {
	case Some:
		return value
	case None:
		return f()
	}
}

func (o Option[T]) OrElse(f func() Option[T]) Option[T] {
	switch o.(variant) {
	case Some:
		return o
	case None:
		return f()
	}
}

func Map[T, U any](o Option[T], f func(T) U) Option[U] {
	switch value := o.(variant) {
	case Some:
		return Option.Some(f(value))
	case None:
		return Option[U].None()
	}
}

func AndThen[T, U any](o Option[T], f func(T) Option[U]) Option[U] {
	switch value := o.(variant) {
	case Some:
		return f(value)
	case None:
		return Option[U].None()
	}
}

func (o Option[T]) Filter(f func(T) bool) Option[T] {
	if value, ok := o.(Some); ok && f(value) {
		return o
	}
	return Option[T].None()
}

func Flatten[T any](o Option[Option[T]]) Option[T] {
	return o.UnwrapOr(Option[T].None())
}
