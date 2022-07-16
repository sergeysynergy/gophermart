package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/sergeysynergy/gophermart/internal/gophermart"
	"log"
	"time"
)

func (s *Storage) initOrders(ctx context.Context) error {
	tableName := "orders"
	_, err := s.db.ExecContext(ctx, "select * from "+tableName+";")
	if err != nil {
		queryCreateTable := `
			CREATE TABLE ` + tableName + ` (
				id bigint NOT NULL,
				user_id bigint NOT NULL,
				status char(256) NOT NULL, 
				accrual bigint,
				uploaded_at timestamp NOT NULL, 
				PRIMARY KEY (id)
			);
		`

		_, err = s.db.ExecContext(ctx, queryCreateTable)
		if err != nil {
			return err
		}

		log.Printf("[DEBUG] table `%s` created", tableName)
	}

	err = s.initOrdersStatements()
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) initOrdersStatements() error {
	tableName := "orders"
	var err error
	var stmt *sql.Stmt

	// добавление нового заказа
	stmt, err = s.db.PrepareContext(
		s.ctx,
		"INSERT INTO "+tableName+" (id, user_id, status, uploaded_at) VALUES ($1, $2, $3, $4)",
	)
	if err != nil {
		return err
	}
	s.stmts["ordersInsert"] = stmt

	// запрос заказа по ID
	stmt, err = s.db.PrepareContext(
		s.ctx,
		"SELECT * FROM "+tableName+" WHERE id=$1",
	)
	if err != nil {
		return err
	}
	s.stmts["orderGetByID"] = stmt

	// обновление заказа
	stmt, err = s.db.PrepareContext(
		s.ctx,
		"UPDATE "+tableName+" SET status = $2, accrual = $3 WHERE id = $1",
	)
	if err != nil {
		return err
	}
	s.stmts["ordersUpdate"] = stmt

	// запрос заказа по ID
	stmt, err = s.db.PrepareContext(
		s.ctx,
		"SELECT * FROM "+tableName+" WHERE id=$1",
	)
	if err != nil {
		return err
	}
	s.stmts["ordersGetByID"] = stmt

	// запрос списка заказов пользователя по user_id
	stmt, err = s.db.PrepareContext(
		s.ctx,
		"SELECT * FROM "+tableName+" WHERE user_id=$1 order by uploaded_at",
	)
	if err != nil {
		return err
	}
	s.stmts["ordersGetForUser"] = stmt

	// запрос заказов для очереди обработки: только со статусом NEW и PROCESSING
	stmt, err = s.db.PrepareContext(
		s.ctx,
		"SELECT * FROM "+tableName+" WHERE status='NEW' or status='PROCESSING' order by uploaded_at LIMIT $1",
	)
	if err != nil {
		return err
	}
	s.stmts["ordersGetForPool"] = stmt

	return nil
}

func (s *Storage) AddOrder(o *gophermart.Order) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txInsert := tx.StmtContext(s.ctx, s.stmts["ordersInsert"])
	txGetByID := tx.StmtContext(s.ctx, s.stmts["ordersGetByID"])
	var bo gophermart.Order
	date := new(string)
	accrual := new(sql.NullInt64)

	row := txGetByID.QueryRowContext(s.ctx, o.ID)
	err = row.Scan(&bo.ID, &bo.UserID, &bo.Status, accrual, date)
	if err != nil {
		if err == sql.ErrNoRows {
			// добавим новую запись в случае отсутствия результата
			_, err = txInsert.ExecContext(s.ctx, o.ID, o.UserID, o.Status, o.UploadedAt)
			if err != nil {
				return err
			}

			// всё хорошо, выполним транзакцию
			err = tx.Commit()
			if err != nil {
				return fmt.Errorf("add order transaction failed - %w", err)
			}
			return nil
		}

		return err
	}

	// заказ уже существует, обработаем ошибку
	if bo.UserID == o.UserID {
		return gophermart.ErrOrderAlreadyLoadedByUser
	}
	return gophermart.ErrOrderAlreadyLoadedByAnotherUser
}

func (s *Storage) GetOrder(orderID uint64) (*gophermart.Order, error) {
	o := &gophermart.Order{}
	accrual := new(sql.NullInt64)
	date := new(string)

	row := s.stmts["orderGetByID"].QueryRowContext(s.ctx, orderID)
	err := row.Scan(&o.ID, &o.UserID, &o.Status, accrual, date)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("order not found - %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get order - %w", err)
	}

	if accrual.Valid {
		o.Accrual = uint64(accrual.Int64)
	}

	if o.UploadedAt, err = time.Parse(time.RFC3339, *date); err != nil {
		return nil, err
	}

	return o, nil
}

func (s *Storage) GetUserOrders(id uint64) ([]*gophermart.Order, error) {
	orders := make([]*gophermart.Order, 0)

	rows, err := s.stmts["ordersGetForUser"].QueryContext(s.ctx, id)
	if err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var bo gophermart.Order
		accrual := new(sql.NullInt64)
		date := new(string)

		err = rows.Scan(&bo.ID, &bo.UserID, &bo.Status, accrual, date)
		if err != nil {
			return nil, err
		}

		if accrual.Valid {
			bo.Accrual = uint64(accrual.Int64)
		}

		if bo.UploadedAt, err = time.Parse(time.RFC3339, *date); err != nil {
			return nil, err
		}

		orders = append(orders, &bo)
	}

	return orders, nil
}

func (s *Storage) GetPullOrders(limit uint32) (map[uint64]*gophermart.Order, error) {
	orders := make(map[uint64]*gophermart.Order)

	rows, err := s.stmts["ordersGetForPool"].QueryContext(s.ctx, limit)
	if err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var bo gophermart.Order
		accrual := new(sql.NullInt64)
		date := new(string)

		err = rows.Scan(&bo.ID, &bo.UserID, &bo.Status, accrual, date)
		if err != nil {
			return nil, err
		}

		if accrual.Valid {
			bo.Accrual = uint64(accrual.Int64)
		}

		if bo.UploadedAt, err = time.Parse(time.RFC3339, *date); err != nil {
			return nil, err
		}

		orders[bo.ID] = &bo
	}

	return orders, nil
}

func (s *Storage) UpdateOrder(o *gophermart.Order) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txUpdateOrder := tx.StmtContext(s.ctx, s.stmts["ordersUpdate"])
	txUpdateBalance := tx.StmtContext(s.ctx, s.stmts["balanceUpdate"])
	txGetBalance := tx.StmtContext(s.ctx, s.stmts["balanceGet"])

	// обновим заказ
	if o.Status == gophermart.StatusProcessed {
		// записываем начисление только для заказов со статусом выполнено
		_, err = txUpdateOrder.ExecContext(s.ctx, o.ID, o.Status, o.Accrual)
		if err != nil {
			return fmt.Errorf("failed to update order - %w", err)
		}

		// обновим баланс пользователя: сначала получим текущее значение
		b := &gophermart.Balance{}
		row := txGetBalance.QueryRowContext(s.ctx, o.UserID)
		err = row.Scan(&b.UserID, &b.Current, &b.Withdrawn)
		if err != nil {
			return fmt.Errorf("failed to get user balance - %w", err)
		}
		current := b.Current + o.Accrual // прибавим начисленные баллы
		// обновим баланс с новым значением
		_, err = txUpdateBalance.ExecContext(s.ctx, b.UserID, current, b.Withdrawn)
		if err != nil {
			return fmt.Errorf("failed to update user balance - %w", err)
		}
	} else {
		// для всех остальных статусов - начисления не записываем, баланс не обновляем
		_, err = txUpdateOrder.ExecContext(s.ctx, o.ID, o.Status, o.Accrual)
		if err != nil {
			return fmt.Errorf("failed to update order - %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("update order transaction failed - %w", err)
	}

	return nil
}
