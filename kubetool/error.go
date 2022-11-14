package kubetool

import "github.com/pkg/errors"

const (
	errNotReady        			  = "nodeNotReady"
	rescueTypeUncodron        = "uncordon"
	rescueTypePostJob         = "postJob"
)

// Errors represent typed error
type Errors struct {
	code       string
	err        error
	rescueStep string
}

// NewErrNodeNotReady permit to return error of type nodeNotReady
func NewErrNodeNotReady(nodeName string) error {
	return &Errors{
		code: errNotReady,
		err:  errors.Errorf("Node %s is not on ready state", nodeName),
	}
}

// NewRescueError permit to return error of type rescue that need uncordon step
func NewRescueUncordonError(err error) error {
	return &Errors{
		err:        err,
		rescueStep: rescueTypeUncodron,
	}
}

// NewRescueError permit to return error of type rescue that need uncordon and post job step
func NewRescuePostJobError(err error) error {
	return &Errors{
		err:        err,
		rescueStep: rescueTypePostJob,
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

// IsRescueUncordon permit to check if error need to invoke uncordon as rescue step
func IsRescueUncordon(err error) bool {
	errors, ok := err.(*Errors)
	if ok && errors.rescueStep == rescueTypeUncodron {
		return true
	}

	return false
}

// IsRescuePostJob permit to check if error need to invoke uncordon and post job as rescue step
func IsRescuePostJob(err error) bool {
	errors, ok := err.(*Errors)
	if ok && errors.rescueStep == rescueTypePostJob {
		return true
	}

	return false
}

// Err implement error interface
func (e *Errors) Error() string {
	return e.err.Error()
}
