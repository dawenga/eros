# Eros
Yup, yet another error handling library. There are a lot of great things about Go and for the most part the idiom of 
simplicity serves it well. One way to do things, without a lot of hype. There are times however, that the idiomatic KISS 
approach doesn't work so well, especially when trying to manage complex, open-ended problems. Errors, in our opinion, is 
one of these complex issues, which results in an annoyingly large amount of redundant, boiler-plated tech-debt. 

## Why

Yes we're aware of the litany of error handling libraries out there. The main purpose behind Eros is an attempt to bring
an opinionated view of error handling to our code for conformity. We realize this will not be to everyone's liking 
however, it does strive to be clear, simple, complete as well as to conform / work with idiomatic Go. Lastly, it
promises to reduce the verbosity of your code... significantly by removing the boiler-plate.

We all know this pattern, all to well:
```go
    if err != nil { 
        return nil, errors.Wrap(err, "my message")
    }
```
In one of our internal projects, we found this same block over 300+ times in just two files. It's sort of a holly grail
to achieve a single handler for a function, that can be safely propagated and both fail fast or fail through depending 
on the contextual need.

Now with the magic of Generics, Eros can achieve that goal where previously it simply wasn't possible. There have been
third party 'generated' attempts, but they just didn't work well within the language themselves for this use case (again
IOHO).

Now that we know generics will be a main line feature in 1.18, we're bringing our ideas and sharing them with the world.

## How it works

Take the following example:
```go
// ReadFileBuffer - return the contents of a file using a Result object.
func ReadFileBuffer(filepath string) (res string, err error) {

	
	// pervasive tuple (value, error) return exemplified in the os package 
    fl, err := CheckVal(os.ReadFile(filepath))
	if err != nil {
		return "", errors.Wrapf(err, "failed to read file: %s", filepath)
    }
	// cast []byte to string
    res = string(fl)
	
	// we get the file info object, this doesn't stop us from moving forward
	fi, err := os.Stat(filepath)
	if err != nil {
	    err = errors.Wrapf(err, "failed to get FileInfo for: %s", filepath)	
    }   
	
	// if it's an empty file, remove it. Note that we could possibly have an error on fl but
	// we go ahead and remove it if res is less than 1 anyway
    if fi.Size() == 0 || len(res) < 1 {
        if err := os.Remove("test.txt")); err != nil {
            return "", errors.Wrapf(err, "failed to delete file: %s", filepath)
        }	
    }
	
    return
}
```
First, don't be so hard on me if you find the above example wrong. It's meant to show the intent, not to be something one
would put into production. That said, given the function above, consider that os.ReadFile returns the ever vigilant tuple
(value, err). We find this pattern proliferated throughout the language and runtime and in this version of ReadFileBuffer,
we pass the buck, up the chain in exactly the same manor, with an empty string and a wrapped error. os.Remove is a similar
pattern, but only returns an error. Let's consider for a moment what are the potential issues we're dealing with here:
- We have multiple error handler blocks that fail fast (exit the function)
- We have an error block that fails through and allows us to continue despite the error. We do want to trap said error
  however, by wrapping int and carrying it forward.
- We're wrapping errors manually, and unless we're easily shadowing the error we want to trap as can be
  seen in the last block where, in reality we now loose the error from the stat call by shadowing.
- It's hard to follow, it's not very clear why we're doing things when an actual error occurs

So, how can Eros help? here is a re-implementation using Eros:

```go
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
        res.Error = Wrapf(err, "failed to get FileInfo for: %s", filepath).WithCause(res.Error)
  
  })
  
  // if it's an empty file, remove it. Note that we could possibly have an error on fl but
  // we go ahead and remove it if res is less than 1 anyway
  if fi.Size() == 0 || len(res.Value) < 1 {
        Check(os.Remove("test.txt"))
  }
  return
}
```
the os.ReadFile call in Check, if we need to handle an error, our error handler takes care of it for us, otherwise we
assign the appropriate value to fl.


Take another scenario, when one implements our Result type as the return type of your function, you can simply call Check()
on the actual result, assigning the proper value once again:
```go
func main() {

    defer ErrorHandler(func(err Error) {
        fmt.Printf("we have a proper error %s", errors.Unwrap(err))
    })()
    
    fmt.Println("Reading file")
    res := ReadFileBuffer("test.txt").Check()
    
    fmt.Println(res)
}
```

See what happens when ReadFileBuffer returns a Result type natively, we can daisy-chain the Check() off the result to
get the string which then populates res. This is of course still type safe because of the use of ```Result[string]``` in
the function signature of ReadFileBuffer.

## Where did I come up with the idea?

So, admittedly, I sort of inherited much of the concept from rust, which very eloquently handles errors with a similar
mechanism. Unlike rust however, the concept of a single error handler in your function has been sort of an appealing / 
attempted goal that I've been after for quite some time. Now with the magic of generics, I can finally achieve the sort
of single call check I've been after for so long... Thanks go 1.18!

## Dependencies

This will only, unfortunately work with go 1.18+ and you should get a compile error if you're using a compiler previous 
to that one.

The unit tests depend on github.com/pkg/errors but the module is otherwise internally relent on just the go runtime

## Conclusions

This is a rather simple library, there really isn't much in the way of code or complexity here. It is my hope however, 
that it brings you a significant reduction in time, maintenance and boiler-plate happiness to you're coding life.

## License

This work is licensed under the [Apache 2.0 license](http://www.apache.org/licenses/LICENSE-2.0)

