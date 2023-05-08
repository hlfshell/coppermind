package service

import (
	"github.com/hlfshell/coppermind/internal/store"
	"github.com/hlfshell/coppermind/pkg/users"
)

type UserService struct {
	db store.Store
}

func NewUserService(db store.Store) *UserService {
	return &UserService{
		db: db,
	}
}

func (service *UserService) CreateUser(user *users.User, password string) error {
	return service.db.CreateUser(user, password)
}

func (service *UserService) ResetPassword(userId string, token string, password string) error {
	return service.db.ResetPassword(userId, token, password)
}
