package userrepo

type UserRepo map[string]string

func New() *UserRepo {
	return &UserRepo{}
}

func (r *UserRepo) AddUser(username, password string) {
	(*r)[username] = password
}

func (r *UserRepo) Authenticate(username, password string) bool {
	return (*r)[username] == password
}
