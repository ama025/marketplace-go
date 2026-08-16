package shared

import "fmt"

type StatusNotFound struct {
	Resource string
	Key      any
}

func (e StatusNotFound) Error() string {
	return fmt.Sprintf("%s with key '%v' not found", e.Resource, e.Key)
}

type StatusBadRequest struct {
	Message string
}

func (e StatusBadRequest) Error() string {
	return fmt.Sprintf("bad request: %s", e.Message)
}

type StatusConflict struct {
	Resource string
	Message  string
}

func (e StatusConflict) Error() string {
	return fmt.Sprintf("conflict on %s: %s", e.Resource, e.Message)
}

type StatusUnauthorized struct {
	Message string
}

func (e StatusUnauthorized) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("unauthorized: %s", e.Message)
	}
	return "unauthorized"
}

type StatusForbidden struct {
	Message string
}

func (e StatusForbidden) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("forbidden: %s", e.Message)
	}
	return "forbidden"
}

func IsNotFound(err error) bool {
	_, ok := err.(StatusNotFound)
	return ok
}

func IsBadRequest(err error) bool {
	_, ok := err.(StatusBadRequest)
	return ok
}

func IsConflict(err error) bool {
	_, ok := err.(StatusConflict)
	return ok
}