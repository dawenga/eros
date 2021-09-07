package eros

import (
	"fmt"
	"reflect"
)

/*
	Eros re-implements all the standard stuff to keep it all in house. This provides
    a uniform experience / package for the user of the module in addition to keeping
 	things idiomatic and backwards compatible.

	One aught to be able to drop in Eros and start using it interactively within one's
	existing project without issue. For this reason, we fully implement IS, AS, Wrap,
	Unwrap allowing us to be backwards compatible to previous compilers that don't
	necessarily support those features.

*/

// New - Just return an error and string
func New(msg string) *Error {
	return &Error{
		msg,
		nil,
		nil,
		0,
	}
}

// errorType - type of an error interface
var errorType = reflect.TypeOf((*error)(nil)).Elem()

//Newf - New, with format. Just syntax sugar
func Newf(format string, args ...interface{}) *Error {
	return New(fmt.Sprintf(format, args...))
}

//Count - returns the depth count of the errors
func (s Error) Count() int {
	return s.count
}

//Error - implement the error interface
func (e Error) Error() string {
	cause := ""
	if e.next != nil {
		cause = e.next.Error()
	}
	if e.cause != nil {
		cause = fmt.Sprintf(" %s\n root cause; %s", cause, e.cause.Error())
	}
	return fmt.Sprintf(" %s (cause count %d)\n%s", e.msg, e.count, cause)
}

//Unwrap - implement the Unwrap interface
func (e *Error) Unwrap() error {
	if e != nil {
		if e.next != nil {
			return e.next
		} else {
			return e.cause
		}
	}
	return nil
}

// CastOrWrap - cast an interface error to an *Error. If not possible, wrap it.
func CastOrWrap(err error) *Error {
	de := dereference(err)
	if e, ok := de.(Error); ok {
		return &e
	} else {
		return Wrap(err, "cast to eros.Error")
	}
}

// Wrap - Wrap an error
func Wrap(err error, msg string) *Error {
	return &Error{
		msg,
		err,
		nil,
		1,
	}
}

//WithCause - appends a new cause error to the chain. This is nil safe
func (e *Error) WithCause(err error) *Error {
	if e == nil {
		e = CastOrWrap(err)
	} else if err != nil && !Is(e, err) {
		v := CastOrWrap(err)
		if e.next != nil {
			e.next = v.WithCause(e.next)
		} else {
			e.next = v
		}
		e.count = (v.count + 1)
	}
	return e
}

// Wrapf - Wrap an error... with formatting
func Wrapf(err error, msg string, vars ...interface{}) *Error {
	return Wrap(err, fmt.Sprintf(msg, vars...))
}

// Is - test for equality
func Is(err, target error) bool {
	if target == nil {
		return err == target
	}

	isComparable := reflect.TypeOf(target).Comparable()
	for {
		if isComparable && err == target {
			return true
		}
		if x, ok := err.(interface{ Is(error) bool }); ok && x.Is(target) {
			return true
		}
		if err = Unwrap(err); err == nil {
			return false
		}
	}
}

// dereference. As only works with instances, not pointers
func dereference(err error) error {
	if err != nil {
		val := reflect.ValueOf(err)
		typ := val.Type()
		if typ.Kind() == reflect.Ptr && !val.IsNil() {
			if e, ok := val.Elem().Interface().(error); ok {
				err = e
			}
		}
	}
	return err
}

// As - check and assign, in consideration of the entire chain. Note
// that our version dereferences pointers an allows AS to succeed
func As(err error, target interface{}) bool {
	if target == nil {
		panic("errors: target cannot be nil")
	}
	val := reflect.ValueOf(target)
	typ := val.Type()
	if typ.Kind() != reflect.Ptr || val.IsNil() {
		panic("errors: target must be a non-nil pointer")
	}
	targetType := typ.Elem()
	if targetType.Kind() != reflect.Interface && !targetType.Implements(errorType) {
		panic("errors: *target must be interface or implement error")
	}
	for err != nil {
		de := dereference(err)
		if reflect.TypeOf(de).AssignableTo(targetType) {
			val.Elem().Set(reflect.ValueOf(de))
			return true
		}
		if x, ok := de.(interface{ As(interface{}) bool }); ok && x.As(target) {
			return true
		}
		err = Unwrap(err)
	}
	return false
}

// Unwrap -  unwrap an error
func Unwrap(err error) error {
	u, ok := err.(interface {
		Unwrap() error
	})
	if !ok {
		return nil
	}
	return u.Unwrap()
}

// Error - our own version of an error, which can wrap others
type Error struct {
	msg   string
	cause error
	next  *Error
	count int
}
