package kubetool

import "github.com/pkg/errors"

const (
	errNotReady string = "nodeNotReady"
)

// Errors represent typed error
type Errors struct {
	code string
	err  error
}

// NewErrNodeNotReady permit to return error of type nodeNotReady
func NewErrNodeNotReady(nodeName string) error {
	return &Errors{
		code: errNotReady,
		err:  errors.Errorf("Node %s is not on ready state", nodeName),
	}
}

// IsErrNodeNotReady permit to check if error is type of nodeNotReady
func IsErrNodeNotReady(err error) bool {
	errors, ok := err.(*Errors)
	if ok && errors.code == errNotReady {
		return true
	}

	return false
}

// Err implement error interface
func (e *Errors) Error() string {
	return e.err.Error()
}
