package basicstorage

import (
	"github.com/sergeysynergy/gophermart/internal/gophermart"
	"github.com/sergeysynergy/gophermart/pkg/loon"
	"log"
	"strconv"
)

func (s *Storage) AddOrder(o *gophermart.Order) error {
	id := strconv.Itoa(int(o.ID))
	if !loon.IsValid(id) {
		log.Printf("[WARNING] Failed to add new order - %s: %s\n", gophermart.ErrOrderInvalidFormat, id)
		return gophermart.ErrOrderInvalidFormat
	}

	s.ordersByIDMu.Lock()
	s.ordersByID[o.ID] = o
	s.ordersByIDMu.Unlock()

	return nil
}

func (s *Storage) GetOrder(id uint64) (*gophermart.Order, error) {
	return nil, nil
}

func (s *Storage) GetPullOrders(_ uint32) (map[uint64]*gophermart.Order, error) {
	s.ordersByIDMu.RLock()
	defer s.ordersByIDMu.RUnlock()

	orders := make(map[uint64]*gophermart.Order)
	for k, v := range s.ordersByID {
		if v.Status == gophermart.StatusNew || v.Status == gophermart.StatusProcessing {
			orders[k] = v
		}
	}

	return orders, nil
}

func (s *Storage) UpdateOrder(o *gophermart.Order) error {
	id := strconv.Itoa(int(o.ID))
	if !loon.IsValid(id) {
		return gophermart.ErrOrderInvalidFormat
	}

	s.ordersByIDMu.Lock()
	s.ordersByID[o.ID] = o
	s.ordersByIDMu.Unlock()

	return nil
}
