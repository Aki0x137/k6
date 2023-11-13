package fs

// newFsError creates a new Error object of the provided kind and with the
// provided message.
func newFsError(k errorKind, message string) *fsError {
	return &fsError{
		Name:    k.String(),
		Message: message,
		kind:    k,
	}
}

// errorKind indicates the kind of file system error that has occurred.
//
// Its string representation is generated by the `enumer` tool. The
// `enumer` tool is run by the `go generate` command. See the `go generate`
// command documentation.
// The tool itself is not tracked as part of the k6 go.mod file, and
// therefore must be installed manually using `go install github.com/dmarkham/enumer`.
//
//go:generate enumer -type=errorKind -output errors_gen.go
type errorKind uint8

const (
	// NotFoundError is emitted when a file is not found.
	NotFoundError errorKind = iota + 1

	// InvalidResourceError is emitted when a resource is invalid: for
	// instance when attempting to open a directory, which is not supported.
	InvalidResourceError

	// ForbiddenError is emitted when an operation is forbidden.
	ForbiddenError

	// TypeError is emitted when an incorrect type has been used.
	TypeError

	// EOFError is emitted when the end of a file has been reached.
	EOFError
)

// fsError represents a custom error object emitted by the fs module.
//
// It is used to provide a more detailed error message to the user, and
// provide a concrete error type that can be used to differentiate between
// different types of errors.
//
// Exposing error types to the user in a way that's compatible with some
// JavaScript error handling constructs such as `instanceof` is still non-trivial
// in Go. See the [dedicated goja issue] with have opened for more details.
//
// [dedicated goja issue]: https://github.com/dop251/goja/issues/529
type fsError struct {
	// Name contains the name of the error as formalized by the [ErrorKind]
	// type.
	Name string `json:"name"`

	// Message contains the error message as presented to the user.
	Message string `json:"message"`

	// kind contains the kind of error that has occurred.
	kind errorKind
}

// Ensure that the Error type implements the Go `error` interface.
var _ error = (*fsError)(nil)

// Error implements the Go `error` interface.
func (e *fsError) Error() string {
	return e.Name + ": " + e.Message
}
