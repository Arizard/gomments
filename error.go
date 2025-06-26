package gomments

import "fmt"

type ServiceError interface {
	error
	Status() int
}

type serviceError struct {
	status int
	err    error
}

func (e serviceError) Error() string {
	return e.err.Error()
}

func (e serviceError) Status() int {
	return e.status
}

func Error(c int, e error) serviceError {
	return serviceError{status: c, err: e}
}

func Errorf(c int, s string, args ...any) serviceError {
	return serviceError{status: c, err: fmt.Errorf(s, args...)}
}
