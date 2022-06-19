package gophermart

type Credentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UseCases interface {
	PostWithdraw(*WithdrawProxy) error
	GetWithdrawals(userID uint64) ([]*WithdrawProxy, error)
	GetBalance(userID uint64) (*BalanceProxy, error)
}

type Storer interface {
	AddUser(*User) (uint64, error)
	GetUser(interface{}) (*User, error)
	DeleteUser(string) error

	AddSession(*Session) error
	GetSession(string) (*Session, error)
	DeleteSession(string) error

	AddOrder(*Order) error
	GetOrder(orderID uint64) (*Order, error)
	GetPullOrders(uint32) (map[uint64]*Order, error)
	GetUserOrders(userID uint64) ([]*Order, error)
	UpdateOrder(*Order) error

	GetBalance(userID uint64) (*Balance, error)
	AddWithdraw(*Withdraw) error
	GetUserWithdrawals(userID uint64) ([]*Withdraw, error)
}
