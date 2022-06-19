package gophermart

import (
	"fmt"
	"github.com/google/uuid"
	"log"
	"strconv"
	"strings"
	"time"
)

func (g *GopherMart) Register(creds *Credentials) (*Session, error) {
	_, err := g.Users.Add(creds)
	if err != nil {
		return nil, err
	}

	session, err := g.Login(creds, "")
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (g *GopherMart) Login(creds *Credentials, oldToken string) (*Session, error) {
	user, err := g.Users.Get(creds.Login)
	if err != nil {
		return nil, err
	}

	check := user.CheckPassword(creds.Password)
	if !check {
		return nil, ErrInvalidPair
	}

	if oldToken != "" {
		err = g.Sessions.Delete(oldToken)
		if err != nil {
			log.Println("[ERROR]", err)
		}
	}

	// Создадим новый токен при помощи библиотеки github.com/google/uuid
	newToken := uuid.NewString()
	expiresAt := time.Now().Add(600 * time.Second)

	// Запишем сессию в хранилище
	s := &Session{
		UserID: user.ID,
		Token:  newToken,
		Expiry: expiresAt,
	}
	err = g.Sessions.Add(s)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (g *GopherMart) Logout(token string) error {
	err := g.Sessions.Delete(token)
	if err != nil {
		return err
	}

	return nil
}

func (g *GopherMart) PostOrders(orderID, userID uint64) error {
	err := g.Orders.Add(orderID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (g *GopherMart) GetOrders(userID uint64) ([]*OrderProxy, error) {
	ors, err := g.Orders.GetUserOrders(userID)
	if err != nil {
		return nil, err
	}

	//{ need
	//	"number": "346436439",
	//	"status": "INVALID",
	//	"uploaded_at": "2020-12-09T16:09:53+03:00"
	//}
	layout := "2006-01-02T15:04:05-07:00"

	orsPr := make([]*OrderProxy, 0)
	for _, o := range ors {
		po := &OrderProxy{
			Number:     fmt.Sprint(o.ID),
			Status:     strings.TrimSpace(o.Status),
			Accrual:    float64(o.Accrual) / 100,
			UploadedAt: o.UploadedAt.Format(layout),
		}

		orsPr = append(orsPr, po)
	}

	return orsPr, nil
}

func (g *GopherMart) PostWithdraw(wpr *WithdrawProxy) error {
	orderID, err := strconv.Atoi(wpr.Order)
	if err != nil {
		return ErrOrderInvalidFormat
	}

	withdraw := &Withdraw{
		OrderID: uint64(orderID),
		UserID:  wpr.UserID,
		Sum:     uint64(wpr.Sum * 100),
	}

	err = g.Withdrawals.Add(withdraw)
	if err != nil {
		return err
	}

	return nil
}

func (g *GopherMart) GetWithdrawals(userID uint64) ([]*WithdrawProxy, error) {
	wds, err := g.Withdrawals.GetWithdrawals(userID)
	if err != nil {
		return nil, err
	}

	wdsPr := make([]*WithdrawProxy, len(wds))
	for _, v := range wds {
		wpr := &WithdrawProxy{
			Order:       fmt.Sprint(v.OrderID),
			Sum:         float64(v.Sum) / 100,
			ProcessedAt: v.ProcessedAt.Format(time.RFC3339),
		}
		wdsPr = append(wdsPr, wpr)
	}

	return wdsPr, nil
}

func (g *GopherMart) GetBalance(userID uint64) (*BalanceProxy, error) {
	bl, err := g.Balances.Get(userID)
	if err != nil {
		return nil, err
	}

	// храним баланс в копейках, отдаём в рублях: поэтому делим на 100
	blPr := &BalanceProxy{
		Current:   float64(bl.Current) / 100,
		Withdrawn: float64(bl.Withdrawn) / 100,
	}

	return blPr, nil
}
