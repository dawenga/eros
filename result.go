package eros

import "reflect"

// Result - represents the traditional (value, error) tuple as an actual return
// type.
type Result[T any] struct {
	Value T
	Error error
}

// Handler - type'ifies our error handler function
type Handler func(res *Error)

// Check - raises a panic if err != nil
func (r Result[T]) Check() T {
	if r.Error != nil {
		panic(CastOrWrap(r.Error))
	}
	return r.Value
}

// Handle - handles the error in a lambda, then still // returns T. This gives
// the user the opportunity to decide whether or not fail through instead of fast
func (r Result[T]) Handle(handler Handler) T {
	if r.Error != nil {
		handler(CastOrWrap(r.Error))
	}
	return r.Value
}

// Check - is used to apply a default handler (or a full on panic) to an existing
// function that only returns an error.
func Check(err error) {
	if err != nil {
		panic(CastOrWrap(err))
	}
	return
}

// CheckNotNil - Prove val isn't nil and return val, otherwise invoke the error handler
func CheckNotNil[T any](val T, msg string) T {
	v := reflect.ValueOf(val)
	if v.IsNil() {
		panic(New(msg))
	}
	return val
}

// CheckVal (checks) without casting and returns the value portion of the value/error
// tuple
func CheckVal[T any](val T, err error) T {
	result := Result[T]{
		Value: val,
		Error: err,
	}
	return result.Check()
}

// Cast - Cast the return contents to a result type, which can either check or handle
// a result.
func Cast[T any](val T, err error) (res *Result[T]) {
	result := Result[T]{
		Value: val,
		Error: err,
	}
	return &result
}

// ErrorHandler - handle but only get err instead of the full result. This lack
// of information may for the most part beO OK especially in legacy situations.
// Note; this will work even if the panic' error is wrapped / nested deep
func ErrorHandler(handler Handler) func() {
	return func() {
		// if we're in a panic, then
		if r := recover(); r != nil {
			// check to see if we're an eros.Error
			if e, ok := r.(error); ok {
				handler(CastOrWrap(e))
				return
			}
			// we can keep panicking, this isn't coming from us
			panic(r)
		}
	}
}
