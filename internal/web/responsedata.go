package web

import (
	"ratingserver/auth/users"
	"ratingserver/internal/web/webpath"
)

type data struct {
	Title  string
	Path   map[string]string
	User   users.User
	Errors []string
	Data   map[string]any
}

func Data(title string) data {
	return data{
		Title: title,
		Path:  webpath.Path(),
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

func (m data) WithErrors(errs ...string) data {
	m.Errors = append(m.Errors, errs...)
	return m
}
