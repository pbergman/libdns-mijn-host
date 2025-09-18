package client

import (
	"errors"
)

type status struct {
	Status            int    `json:"status"`
	StatusDescription string `json:"status_description"`
}

func (s status) Code() int {
	return s.Status
}

func (s status) CodeDescription() string {
	return s.StatusDescription
}

func (s status) Error() error {

	if 200 != s.Status {
		return errors.New(s.StatusDescription)
	}

	return nil
}

type StatusResponse interface {
	Code() int
	CodeDescription() string
	Error() error
}
