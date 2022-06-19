package gophermart

import "time"

type GopherMart struct {
	storage Storer

	Users       *Users
	Sessions    *sessions
	Orders      *orders
	Balances    *balances
	Withdrawals *withdrawals
}

func New(st Storer) *GopherMart {
	gm := &GopherMart{
		storage:  st,
		Users:    newUsers(st),
		Sessions: newSessions(st),
	}
	gm.Orders = newOrders(gm)
	gm.Balances = newBalance(gm)
	gm.Withdrawals = newWithdrawals(gm)

	return gm
}

func WithTestOrders(st Storer) {
	orders := map[uint64]*Order{
		2486622125: {
			ID:         2486622125,
			UserID:     1,
			Status:     StatusNew,
			UploadedAt: time.Now(),
		},
		8646498256: {
			ID:         8646498256,
			UserID:     1,
			Status:     StatusNew,
			UploadedAt: time.Now(),
		},
		5697726585: {
			ID:         5697726585,
			UserID:     1,
			Status:     StatusNew,
			UploadedAt: time.Now(),
		},
		79927398713: {
			ID:         79927398713,
			UserID:     1,
			Status:     StatusProcessing,
			UploadedAt: time.Now(),
		},
		23459035: {
			ID:         23459035,
			UserID:     1,
			Status:     StatusProcessing,
			UploadedAt: time.Now(),
		},
		3147484905: {
			ID:         3147484905,
			UserID:     1,
			Status:     StatusProcessing,
			UploadedAt: time.Now(),
		},
	}
	for _, order := range orders {
		st.AddOrder(order)
	}
}
