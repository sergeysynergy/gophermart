package basicstorage

import (
	"sync"

	"github.com/sergeysynergy/gophermart/internal/gophermart"
)

type Storage struct {
	counter uint64

	usersByLoginMu sync.RWMutex
	usersByLogin   map[string]*gophermart.User
	usersByID      map[uint64]*gophermart.User

	sessionsBySessionTokenMu sync.RWMutex
	sessionsBySessionToken   map[string]*gophermart.Session

	ordersByIDMu sync.RWMutex
	ordersByID   map[uint64]*gophermart.Order
}

func New() *Storage {
	return &Storage{
		usersByLogin:           make(map[string]*gophermart.User),
		usersByID:              make(map[uint64]*gophermart.User),
		sessionsBySessionToken: make(map[string]*gophermart.Session),
		ordersByID:             make(map[uint64]*gophermart.Order),
	}
}
