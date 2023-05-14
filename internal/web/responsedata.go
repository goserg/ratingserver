package web

import (
	"errors"

	"github.com/goserg/ratingserver/auth/users"
	"github.com/goserg/ratingserver/internal/web/webpath"
)

type data struct {
	Title  string
	Path   map[string]string
	User   users.User
	Errors []string
	Data   map[string]any
}

func newData(title string) data {
	return data{
		Title: title,
		Path:  webpath.Path(),
		Data:  make(map[string]any),
	}
}

func (m data) WithUser(user users.User) data {
	m.User = user
	return m
}

func (m data) With(key string, value any) data {
	if m.Data == nil {
		m.Data = make(map[string]any)
	}
	m.Data[key] = value
	return m
}

type multierr interface {
	Unwrap() []error
}

func unwrap(err error) []error {
	var merr multierr
	if errors.As(err, &merr) {
		var errs []error
		for _, err := range merr.Unwrap() {
			errs = append(errs, unwrap(err)...)
		}
		return errs
	}
	return []error{err}
}

func (m data) WithErrors(err error) data {
	for _, err := range unwrap(err) {
		m.Errors = append(m.Errors, err.Error())
	}
	return m
}
