package response

import (
	"errors"
	"time"
)

// DelayedResponse is a struct that allows for a delayed response mechanism.
type DelayedResponse struct {
	timeoutInMinutes int
	valueQueue       chan string
	errorQueue       chan error
	responded        bool
	value            string
	err              error
}

// Value returns the value of the response, waiting for either a value or an error.
func (dr DelayedResponse) Value() (string, error) {
	if dr.responded {
		return dr.value, dr.err
	}

	select {
	case <-time.After(time.Duration(dr.timeoutInMinutes) * time.Minute):
		dr.err = errors.New("timeout waiting for response")
		dr.responded = true
		return "", dr.err
	case dr.err = <-dr.errorQueue:
		dr.responded = true
		return "", dr.err
	case dr.value = <-dr.valueQueue:
		dr.responded = true
		return dr.value, nil
	}
}
