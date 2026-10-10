package userrepo

import (
	"auth/internal/utils"
	"fmt"
)

type UserRepo map[string]string

func New() *UserRepo {
	return &UserRepo{}
}

func (r *UserRepo) AddUser(username, password string) {
	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		fmt.Printf("hashing password failed: %v", err)
	}
	(*r)[username] = passwordHash
}

func (r *UserRepo) Authenticate(username, password string) bool {
	return utils.VerifyPassword((*r)[username], password)
}
