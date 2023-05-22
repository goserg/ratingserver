package web

import (
	"errors"
	"regexp"

	"github.com/gofiber/fiber/v2"
)

type signupRequest struct {
	name     string
	email    string
	password string
}

func parseSignUpRequest(ctx *fiber.Ctx) (signupRequest, error) {
	var err error
	name := ctx.FormValue("username", "")
	err = errors.Join(err, validateUserName(name))
	email := ctx.FormValue("email", "")
	err = errors.Join(err, validateEmail(email))
	password := ctx.FormValue("password", "")
	err = errors.Join(err, validatePassword(password))
	passwordRepeat := ctx.FormValue("password-repeat", "")
	if passwordRepeat != password {
		err = errors.Join(errors.New("пароль не совпадает с подтверждением"))
	}
	if err != nil {
		return signupRequest{}, err
	}
	return signupRequest{
		name:     name,
		email:    email,
		password: password,
	}, nil
}

type signInRequest struct {
	name     string
	password string
}

func parseSignInRequest(ctx *fiber.Ctx) (req signInRequest, err error) {
	name := ctx.FormValue("username", "")
	err = errors.Join(err, validateUserName(name))
	password := ctx.FormValue("password", "")
	err = errors.Join(err, validatePassword(password))
	if err != nil {
		return signInRequest{}, err
	}
	return signInRequest{
		name:     name,
		password: password,
	}, nil
}

func validatePassword(password string) error {
	if password == "" {
		return errors.New("пароль пользователя не должн быть пустым")
	}
	return nil
}

func validateUserName(name string) error {
	var err error
	if name == "" {
		err = errors.Join(err, errors.New("имя пользователя не должно быть пустое"))
	}
	nameRegexp := regexp.MustCompile(`^[A-Za-z]\w+$`)
	if !nameRegexp.MatchString(name) {
		err = errors.Join(err, errors.New("имя пользователя должно начинаться с латинской буквы и содержать только латинские буквы, цифры и знаки подчеркивания"))
	}
	return err
}

func validateEmail(email string) error {
	emailRegexp := regexp.MustCompile(`^[\w-.]+@([\w-]+\.)+[\w-]{2,4}$`)
	if !emailRegexp.MatchString(email) {
		return errors.New("неверный email")
	}
	return nil
}
