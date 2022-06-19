package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/sergeysynergy/hardtest/internal/gophermart"
	"log"
)

func (s *Storage) initBalance(ctx context.Context) error {
	dbName := "balance"
	_, err := s.db.ExecContext(ctx, "select * from "+dbName+";")
	if err != nil {
		queryCreateTable := `
			CREATE TABLE ` + dbName + ` (
				user_id bigint NOT NULL,
				current bigint NOT NULL,
				withdrawn bigint NOT NULL,
				PRIMARY KEY (user_id)
			);
		`

		_, err = s.db.ExecContext(ctx, queryCreateTable)
		if err != nil {
			return err
		}

		log.Printf("[DEBUG] table `%s` created", dbName)
	}

	err = s.initBalanceStatements()
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) initBalanceStatements() error {
	tableName := "balance"
	var err error
	var stmt *sql.Stmt

	// добавляем баланс пользвателя
	stmt, err = s.db.PrepareContext(
		s.ctx,
		"INSERT INTO "+tableName+" (user_id, current, withdrawn) VALUES ($1, 0, 0)",
	)
	if err != nil {
		return err
	}
	s.stmts["balanceInsert"] = stmt

	// запрос текущего баланса пользователя
	stmt, err = s.db.PrepareContext(
		s.ctx,
		"SELECT * FROM "+tableName+" WHERE user_id=$1",
	)
	if err != nil {
		return err
	}
	s.stmts["balanceGet"] = stmt

	// обновление баланса
	stmt, err = s.db.PrepareContext(
		s.ctx,
		"UPDATE "+tableName+" SET current = $2, withdrawn = $3 WHERE user_id = $1",
	)
	if err != nil {
		return err
	}
	s.stmts["balanceUpdate"] = stmt

	return nil
}

func (s *Storage) GetBalance(userID uint64) (*gophermart.Balance, error) {
	b := &gophermart.Balance{}

	row := s.stmts["balanceGet"].QueryRowContext(s.ctx, userID)
	err := row.Scan(&b.UserID, &b.Current, &b.Withdrawn)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user balance not found - %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user balance - %w", err)
	}

	return b, nil
}

func (s *Storage) UpdateBalance(b *gophermart.Balance) error {
	result, err := s.stmts["balanceUpdate"].ExecContext(s.ctx, b.UserID, b.Current, b.Withdrawn)
	if err != nil {
		return fmt.Errorf("failed to update user balance - %w", err)
	}
	_, err = result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to update user balance - %w", err)
	}

	return nil
}
