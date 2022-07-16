package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/sergeysynergy/gophermart/internal/gophermart"
	"log"
	"time"
)

func (s *Storage) initWithdrawals(ctx context.Context) error {
	tableName := "withdrawals"
	_, err := s.db.ExecContext(ctx, "select * from "+tableName+";")
	if err != nil {
		queryCreateTable := `
			CREATE TABLE ` + tableName + ` (
				order_id bigint NOT NULL,
				user_id bigint NOT NULL,
				sum bigint NOT NULL,
				processed_at timestamp NOT NULL, 
				PRIMARY KEY (order_id)
			);
		`

		_, err = s.db.ExecContext(ctx, queryCreateTable)
		if err != nil {
			return err
		}

		log.Printf("[DEBUG] table `%s` created", tableName)
	}

	err = s.initWithdrawalsStatements()
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) initWithdrawalsStatements() error {
	tableName := "withdrawals"
	var err error
	var stmt *sql.Stmt

	// запись о списании средств
	stmt, err = s.db.PrepareContext(
		s.ctx,
		"INSERT INTO "+tableName+" (order_id, user_id, sum, processed_at) VALUES ($1, $2, $3, $4)",
	)
	if err != nil {
		return err
	}
	s.stmts["withdrawalsInsert"] = stmt

	// запрос одного списания по уникальному номеру заказа
	stmt, err = s.db.PrepareContext(
		s.ctx,
		"SELECT * FROM "+tableName+" WHERE order_id=$1",
	)
	if err != nil {
		return err
	}
	s.stmts["withdrawalsGetByID"] = stmt

	// запрос списка расходов пользователя
	stmt, err = s.db.PrepareContext(
		s.ctx,
		"SELECT * FROM "+tableName+" WHERE user_id=$1 ORDER BY processed_at desc",
	)
	if err != nil {
		return err
	}
	s.stmts["withdrawalsGetForUser"] = stmt

	return nil
}

func (s *Storage) AddWithdraw(withdraw *gophermart.Withdraw) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txGetByID := tx.StmtContext(s.ctx, s.stmts["withdrawalsGetByID"])
	txInsertWithdrawal := tx.StmtContext(s.ctx, s.stmts["withdrawalsInsert"])
	txGetBalance := tx.StmtContext(s.ctx, s.stmts["balanceGet"])
	txUpdateBalance := tx.StmtContext(s.ctx, s.stmts["balanceUpdate"])

	// проверим баланс
	var balance gophermart.Balance
	row := txGetBalance.QueryRowContext(s.ctx, withdraw.UserID)
	err = row.Scan(&balance.UserID, &balance.Current, &balance.Withdrawn)
	if err == sql.ErrNoRows {
		return fmt.Errorf("user balance not found - %w", err)
	}
	if err != nil {
		return fmt.Errorf("failed to get user balance - %w", err)
	}
	if balance.Current < withdraw.Sum {
		return gophermart.ErrNotEnoughFunds
	}

	// средств достаточно, обновим баланс
	current := balance.Current - withdraw.Sum
	withdrawn := balance.Withdrawn + withdraw.Sum
	_, err = txUpdateBalance.ExecContext(s.ctx, withdraw.UserID, current, withdrawn)
	if err != nil {
		return fmt.Errorf("failed to update user balance - %w", err)
	}

	// добавим историю списаний
	var bw gophermart.Withdraw
	date := new(string)
	row = txGetByID.QueryRowContext(s.ctx, withdraw.OrderID)
	err = row.Scan(&bw.OrderID, &bw.UserID, &bw.Sum, date)
	if err != nil {
		if err == sql.ErrNoRows {
			// добавим новую запись в случае отсутствия результата
			_, err = txInsertWithdrawal.ExecContext(s.ctx, withdraw.OrderID, withdraw.UserID, withdraw.Sum, time.Now())
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

	if withdraw.UserID == bw.UserID {
		return fmt.Errorf("withdraw already recorded by this user")
	}
	return fmt.Errorf("withdraw already recorded by another user")
}

func (s *Storage) GetUserWithdrawals(userID uint64) ([]*gophermart.Withdraw, error) {
	ws := make([]*gophermart.Withdraw, 0)

	rows, err := s.stmts["withdrawalsGetForUser"].QueryContext(s.ctx, userID)
	if err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var w gophermart.Withdraw
		date := new(string)

		err = rows.Scan(&w.OrderID, &w.UserID, &w.Sum, date)
		if err != nil {
			return nil, err
		}

		if w.ProcessedAt, err = time.Parse(time.RFC3339, *date); err != nil {
			return nil, err
		}

		ws = append(ws, &w)
	}

	return ws, nil
}
