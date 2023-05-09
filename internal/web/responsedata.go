package web

import (
	"ratingserver/auth/users"
	"ratingserver/internal/web/webpath"
)

type Data map[string]any

func NewData(title string) Data {
	return Data{
		"Title": title,
		"Path":  webpath.Path(),
	}
}

func (m Data) User(user users.User) Data {
	m["User"] = user
	return m
}

func (m Data) Data(key string, value any) Data {
	m[key] = value
	return m
}

func (m Data) Errors(errs ...error) Data {
	e, ok := m["Erorrs"]
	if !ok {
	}
	ers, ok := e.([]error)
	if !ok {
		m["Errors"] = errs
		return m
	}
	m["Errors"] = append(ers, errs...)
	return m
}
