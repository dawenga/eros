package eros

import (
	"testing"

	"github.com/pkg/errors"
)

var (
	NewErrorInstance      = New("This is an eros Error")
	ComparedErrorInstance = New("This is an eros Error")
)

func TestAs(t *testing.T) {
	var e Error
	type args struct {
		err    error
		target interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"Test wrapped Error with finding target",
			args{
				errors.Wrapf(NewErrorInstance, "this is our wrapped error"),
				&e,
			},
			true,
		},
		{
			"Test Error as an instance can be cast",
			args{
				NewErrorInstance,
				&e,
			},
			true,
		},
		{
			"Test wrapped Error with no eros Error in chain",
			args{
				errors.Wrapf(errors.New("this is not an eros error"), "this is our wrapped error"),
				&Error{},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := As(tt.args.err, tt.args.target); got != tt.want {
				t.Errorf("As() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIs(t *testing.T) {

	type args struct {
		err    error
		target error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"Test Error that is an eros.Error",
			args{
				NewErrorInstance,
				NewErrorInstance,
			},
			true,
		},
		{
			"Test Error that is an eros.Error but, wrapped / nexted",
			args{
				errors.Wrapf(NewErrorInstance, "this is our wrapped error"),
				NewErrorInstance,
			},
			true,
		},
		{
			"Test Error that is not an eros.Error",
			args{
				errors.Wrapf(New("This is an eros Error"), "this is our wrapped error"),
				&Error{},
			},
			false,
		},
		{
			"Test Error that should be comparable to even though they are different",
			args{
				ComparedErrorInstance,
				NewErrorInstance,
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Is(tt.args.err, tt.args.target); got != tt.want {
				t.Errorf("Is() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnwrap(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"Test wrapped Error with finding target",
			args{
				errors.Wrap(New("This is an eros Error"), "this is our wrapped error"),
			},
			true,
		},
		{
			"Test wrapped Error with finding target",
			args{
				errors.Wrapf(New("This is an eros Error"), "this is our wrapped error %s", "with a wrapped formatter"),
			},
			true,
		},
		{
			"Test wrapped Error with no wrapping",
			args{
				New("there is no wrapped error here"),
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Unwrap(tt.args.err); (err != nil) != tt.wantErr {
				t.Errorf("Unwrap() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
