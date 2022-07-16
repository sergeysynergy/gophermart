package gophermart

import (
	"github.com/sergeysynergy/gophermart/pkg/loon"
	"strconv"
	"time"
)

type Withdraw struct {
	OrderID     uint64
	UserID      uint64
	Sum         uint64
	ProcessedAt time.Time
}

type WithdrawProxy struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	UserID      uint64  `json:"-"`
	ProcessedAt string  `json:"processed_at"`
}

type withdrawals struct {
	linker *GopherMart
}

func newWithdrawals(linker *GopherMart) *withdrawals {
	return &withdrawals{
		linker: linker,
	}
}

func (ws *withdrawals) GetWithdrawals(userID uint64) ([]*Withdraw, error) {
	wds, err := ws.linker.storage.GetUserWithdrawals(userID)
	if err != nil {
		return nil, err
	}

	if len(wds) == 0 {
		return nil, ErrNoContent
	}

	return wds, nil
}

func (ws *withdrawals) Add(withdraw *Withdraw) error {
	// Проверим номер заказа на соответствие алгоритму Луна
	strOrderID := strconv.Itoa(int(withdraw.OrderID))
	if !loon.IsValid(strOrderID) {
		return ErrOrderInvalidFormat
	}

	err := ws.linker.storage.AddWithdraw(withdraw)
	if err != nil {
		return err
	}

	// баланс изменился, удалим запись из кэша баланса
	ws.linker.Balances.mu.Lock()
	delete(ws.linker.Balances.byUserID, withdraw.UserID)
	ws.linker.Balances.mu.Unlock()

	return nil
}
