package gophermart

import (
	"sync"
)

type Balance struct {
	UserID    uint64
	Current   uint64
	Withdrawn uint64
}

type BalanceProxy struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type balances struct {
	linker   *GopherMart
	mu       sync.RWMutex
	byUserID map[uint64]*Balance
}

func newBalance(linker *GopherMart) *balances {
	return &balances{
		linker:   linker,
		byUserID: make(map[uint64]*Balance),
	}
}

func (bs *balances) Get(userID uint64) (*Balance, error) {
	var err error

	bs.mu.RLock()
	b, ok := bs.byUserID[userID]
	bs.mu.RUnlock()
	if !ok {
		b, err = bs.linker.storage.GetBalance(userID)
		if err != nil {
			return nil, err
		}

		// закэшируем полученный баланс пользователя
		bs.mu.Lock()
		bs.byUserID[userID] = b
		bs.mu.Unlock()
	}

	return b, nil
}
