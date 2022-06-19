package gophermart

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"sync"
)

type User struct {
	ID       uint64
	Login    string
	Password []byte
}

func (u *User) CheckPassword(password string) bool {
	if err := bcrypt.CompareHashAndPassword(u.Password, []byte(password)); err != nil {
		return false
	}

	return true
}

type Users struct {
	mu      sync.RWMutex
	storage Storer
	byLogin map[string]*User
	byID    map[uint64]*User
}

func newUsers(st Storer) *Users {
	return &Users{
		storage: st,
		byLogin: make(map[string]*User),
		byID:    make(map[uint64]*User),
	}
}

func HashPass(password string) ([]byte, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	if err != nil {
		return nil, err
	}

	return hashedPassword, nil
}

func (urs *Users) Add(creds *Credentials) (uint64, error) {
	// проверим наличие пользователя в кэше
	urs.mu.RLock()
	_, ok := urs.byLogin[creds.Login]
	urs.mu.RUnlock()
	if ok {
		return 0, ErrLoginAlreadyTaken
	}

	hashedPassword, err := HashPass(creds.Password)
	if err != nil {
		return 0, err
	}

	u := &User{
		Login:    creds.Login,
		Password: hashedPassword,
	}

	id, err := urs.storage.AddUser(u)
	if err != nil {
		return 0, err
	}
	u.ID = id

	// закэшируем полученного пользователя
	urs.mu.Lock()
	urs.byLogin[u.Login] = u
	urs.byID[u.ID] = u
	urs.mu.Unlock()

	return u.ID, nil
}

func (urs *Users) Get(byKey interface{}) (*User, error) {
	var err error
	var u *User
	var ok bool

	urs.mu.RLock()
	switch key := byKey.(type) {
	case string:
		u, ok = urs.byLogin[key]
	case uint64:
		u, ok = urs.byID[key]
	default:
		urs.mu.RUnlock()
		return nil, fmt.Errorf("given type not implemented")
	}
	urs.mu.RUnlock()

	if !ok {
		u, err = urs.storage.GetUser(byKey)
		if err != nil {
			return nil, err
		}
		// закэшируем полученного пользователя
		urs.mu.Lock()
		urs.byLogin[u.Login] = u
		urs.byID[u.ID] = u
		urs.mu.Unlock()
	}

	return u, nil
}

func (urs *Users) Delete(login string) error {
	urs.mu.Lock()
	delete(urs.byLogin, login)
	urs.mu.Unlock()

	err := urs.storage.DeleteUser(login)
	if err != nil {
		return err
	}

	return nil
}
