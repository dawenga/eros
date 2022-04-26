package eros

import (
	"fmt"
	"io/ioutil"
	"os"
)

// ExampleCheck - test fail fast instead of fail through
func ExampleCheck() {

	// Top level error handler
	defer ErrorHandler(func(err *Error) {
		// we know res.err is != nil but we also have val if we need it
		fmt.Println(err.Unwrap())
	})()

	fl := Cast(os.Open("/opt/abc/baddir/file")).Check()
	defer fl.Close()

	// Output: open /opt/abc/baddir/file: no such file or directory

}

// ExampleHandle - test fail through instead of fail fast
func ExampleHandle() {

	close := true
	fl := Cast(os.Open("/opt/abc/baddir/file")).Handle(func(err *Error) {
		// we know res.err is != nil but we also have val if we need it
		fmt.Println(err.Unwrap())
		close = false
		// we could if we wanted to exit at this point but, why not use the defult
		// handler in such a case, which exists on the way out
		// return
	})

	//note that we did not exit the function and we're failing through so
	//this happens either way... be careful here
	if close {
		defer fl.Close()
	}

	// Output: open /opt/abc/baddir/file: no such file or directory
}

// ExampleCheckNotNil - demonstrates the usage of CheckNotNil()
func ExampleCheckNotNil() {

	// Top level error handler
	defer ErrorHandler(func(err *Error) {
		fmt.Println(err.Error())
	})() // <-- don't forget (), or your panic will propagate all the way out

	// nil int
	var ni *int
	a := 1
	ni = &a

	v := CheckNotNil(ni, "ni is pointing to a real value")

	fmt.Printf("v is 1 because it's a copy of the pointer ni %d", *v)

	ni = nil

	v = CheckNotNil(ni, "ni is now nil")
	fmt.Printf("we'll never get here because ni is nil and v isn't assigned; %d", *v)

	// Output: v is 1 because it's a copy of the pointer ni 1 ni is now nil (cause count 0)

}

// ExampleCheckAndSet -Test both the global handler (in ReadFileBuffer) and a
// local handler. A local handler isn't run on the defer (or on the way out)
// which is useful if you want to fail through and keep going
func ExampleCheckAndSet() {

	var e *Error

	// Top level error handler - Check will call this if we can't write to test.txt
	// so to will os.Remove and ReadFileBuffer with Result.Check()
	defer ErrorHandler(func(err *Error) {
		e.WithCause(err) // <-- add error to our chain
		fmt.Println(e.Error())
	})() // <-- don't forget (), or your panic will propagate all the way out

	// This should get us the file data as a string or invoke our err handler
	res := ReadFileBuffer("/opt/badk/dsa/fksdkf").Handle(func(err *Error) {
		e.WithCause(err) // <-- adds error to our chain
	})

	Check(ioutil.WriteFile("test.txt", []byte("Hello Eros!"), 0755))

	res = ReadFileBuffer("test.txt").Check()
	if len(res) == 0 {
		e.WithCause(Newf("failed to assert res is 0. res is; %s", res))
	} else {
		fmt.Printf("Successfully got the contents for test.txt")
	}

	// if we fail to remove the file, we'll invoke the handler
	Check(os.Remove("test.txt"))

	// Output: Successfully got the contents for test.txt
}

// ReadFileBuffer - return the contents of a file using a Result object.
func ReadFileBuffer(filepath string) (res Result[string]) {

	// Top level error handler
	defer ErrorHandler(func(err *Error) {
		res.Error = err.WithCause(res.Error)

	})()

	// pervasive tuple (value, error) return exemplified in the os package
	fl := CheckVal(os.ReadFile(filepath))
	res.Value = string(fl)

	// we get the file info object, this doesn't stop us from moving forward. Note that
	// WithCause is nil safe and will set itself
	fi := Cast(os.Stat(filepath)).Handle(func(err *Error) {
		res.Error = Wrapf(err, "failed to get FileInfo for: %s", filepath).
			WithCause(res.Error)

	})

	// if it's an empty file, remove it. Note that we could possibly have an error on fl but
	// we go ahead and remove it if res is less than 1 anyway
	if fi.Size() == 0 || len(res.Value) < 1 {
		Check(os.Remove("test.txt"))
	}
	return
}
