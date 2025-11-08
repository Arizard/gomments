package gomments

import "fmt"

type ServiceError struct {
	status int
	err    error
}

func (e ServiceError) Error() string {
	return e.err.Error()
}

func (e ServiceError) Status() int {
	return e.status
}

func Error(c int, e error) ServiceError {
	return ServiceError{status: c, err: e}
}

func Errorf(c int, s string, args ...any) ServiceError {
	return ServiceError{status: c, err: fmt.Errorf(s, args...)}
}
