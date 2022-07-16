package basicstorage

import (
	"fmt"
	"github.com/sergeysynergy/gophermart/internal/gophermart"
)

func (s *Storage) AddUser(u *gophermart.User) (uint64, error) {
	s.usersByLoginMu.Lock()
	defer s.usersByLoginMu.Unlock()

	if _, ok := s.usersByLogin[u.Login]; ok {
		return 0, gophermart.ErrLoginAlreadyTaken
	}

	s.counter++
	user := &gophermart.User{
		ID:       s.counter,
		Login:    u.Login,
		Password: u.Password,
	}
	s.usersByLogin[user.Login] = user
	s.usersByID[user.ID] = user

	return user.ID, nil
}

func (s *Storage) GetUser(_key interface{}) (*gophermart.User, error) {
	s.usersByLoginMu.RLock()
	defer s.usersByLoginMu.RUnlock()

	var u *gophermart.User
	var ok bool

	switch key := _key.(type) {
	case string:
		u, ok = s.usersByLogin[key]
		if !ok {
			return nil, fmt.Errorf("user with login `%s` not found", u.Login)
		}
	case uint64:
		u, ok = s.usersByID[key]
		if !ok {
			return nil, fmt.Errorf("user with ID `%d` not found", u.ID)
		}
	default:
		return nil, fmt.Errorf("given type not implemented")
	}

	return u, nil
}

func (s *Storage) DeleteUser(login string) error {
	return fmt.Errorf("method not impemented")
}
