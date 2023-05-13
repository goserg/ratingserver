package web

import (
	"regexp"

	"github.com/gofiber/fiber/v2"
)

type signupRequest struct {
	name     string
	password string
}

func parseSignUpRequest(ctx *fiber.Ctx) (req signupRequest, errs []string) {
	name := ctx.FormValue("username", "")
	errs = validateUserName(name)
	password := ctx.FormValue("password", "")
	errs = append(errs, validatePassword(password)...)
	passwordRepeat := ctx.FormValue("password-repeat", "")
	if passwordRepeat != password {
		errs = append(errs, "Пароль не совпадает с подтверждением.")
	}
	if errs != nil {
		return signupRequest{}, errs
	}
	return signupRequest{
		name:     name,
		password: password,
	}, nil
}

type signInRequest struct {
	name     string
	password string
}

func parseSignInRequest(ctx *fiber.Ctx) (req signInRequest, errs []string) {
	name := ctx.FormValue("username", "")
	errs = validateUserName(name)
	password := ctx.FormValue("password", "")
	errs = append(errs, validatePassword(password)...)
	if errs != nil {
		return signInRequest{}, errs
	}
	return signInRequest{
		name:     name,
		password: password,
	}, nil
}

func validatePassword(password string) []string {
	var errs []string
	if password == "" {
		errs = append(errs, "Пароль пользователя не должн быть пустым.")
	}
	return errs
}

func validateUserName(name string) []string {
	var errs []string
	if name == "" {
		errs = append(errs, "Имя пользователя не должно быть пустое.")
	}
	nameRegexp := regexp.MustCompile(`^[A-Za-z]\w+$`)
	if !nameRegexp.MatchString(name) {
		errs = append(errs, "Имя пользователя должно начинаться с латинской буквы и содержать только латинские буквы, цифры и знаки подчеркивания.")
	}
	return errs
}
