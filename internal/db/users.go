package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/sergeysynergy/hardtest/internal/gophermart"
)

func (s *Storage) initUsers(ctx context.Context) error {
	tableName := "users"
	_, err := s.db.ExecContext(ctx, "select * from "+tableName+";")
	if err != nil {
		queryCreateTable := `
			CREATE TABLE ` + tableName + ` (
				id serial PRIMARY KEY,
				login varchar NOT NULL, 
				password bytea NOT NULL
			);
		`

		_, err = s.db.ExecContext(ctx, queryCreateTable)
		if err != nil {
			return err
		}

		log.Printf("[DEBUG] table `%s` created", tableName)
	}

	err = s.initUsersStatements()
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) initUsersStatements() error {
	tableName := "users"
	var err error
	var stmt *sql.Stmt

	// добавляем пользвателя
	stmt, err = s.db.PrepareContext(
		s.ctx,
		"INSERT INTO "+tableName+" (login, password) VALUES ($1, $2)",
	)
	if err != nil {
		return err
	}
	s.stmts["usersInsert"] = stmt

	// запрос пользователя по логину
	stmt, err = s.db.PrepareContext(
		s.ctx,
		"SELECT * FROM "+tableName+" WHERE login=$1",
	)
	if err != nil {
		return err
	}
	s.stmts["usersGetByLogin"] = stmt

	// запрос пользователя по ID
	stmt, err = s.db.PrepareContext(
		s.ctx,
		"SELECT * FROM "+tableName+" WHERE id=$1",
	)
	if err != nil {
		return err
	}
	s.stmts["usersGetByID"] = stmt

	stmt, err = s.db.PrepareContext(
		s.ctx,
		"DELETE FROM "+tableName+" WHERE login=$1",
	)
	if err != nil {
		return err
	}
	s.stmts["usersDelete"] = stmt

	return nil
}

func (s *Storage) AddUser(u *gophermart.User) (uint64, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	txInsert := tx.StmtContext(s.ctx, s.stmts["usersInsert"])
	txGet := tx.StmtContext(s.ctx, s.stmts["usersGetByLogin"])
	txInsertBalance := tx.StmtContext(s.ctx, s.stmts["balanceInsert"])

	row := txGet.QueryRowContext(s.ctx, u.Login)
	blankUser := gophermart.User{}
	err = row.Scan(&blankUser.ID, &blankUser.Login, &blankUser.Password)
	if err == sql.ErrNoRows {
		// добавим новую запись в случае отсутствия результата
		_, err = txInsert.ExecContext(s.ctx, u.Login, u.Password)
		if err != nil {
			return 0, err
		}

		// получим сгенерённый ID
		row = txGet.QueryRowContext(s.ctx, u.Login)
		err = row.Scan(&u.ID, &u.Login, &u.Password)
		if err != nil {
			return 0, err
		}

		// добавим запись в таблицу балансов пользователей
		_, err = txInsertBalance.ExecContext(s.ctx, u.ID)
		if err != nil {
			return 0, err
		}

	} else if err != nil {
		return 0, err
	} else {
		return 0, gophermart.ErrLoginAlreadyTaken
	}

	err = tx.Commit()
	if err != nil {
		return 0, fmt.Errorf("add user transaction failed - %w", err)
	}

	return u.ID, nil
}

func (s *Storage) GetUser(byKey interface{}) (*gophermart.User, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	txGetByLogin := tx.StmtContext(s.ctx, s.stmts["usersGetByLogin"])
	txGetByID := tx.StmtContext(s.ctx, s.stmts["usersGetByID"])

	var u gophermart.User
	var row *sql.Row

	switch key := byKey.(type) {
	case string:
		row = txGetByLogin.QueryRowContext(s.ctx, key)
	case uint64:
		row = txGetByID.QueryRowContext(s.ctx, key)
	default:
		return nil, fmt.Errorf("given type not implemented")
	}

	err = row.Scan(&u.ID, &u.Login, &u.Password)
	if err == sql.ErrNoRows {
		return nil, gophermart.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("get user transaction failed - %w", err)
	}

	return &u, nil
}

func (s *Storage) DeleteUser(login string) error {
	res, err := s.stmts["usersDelete"].ExecContext(s.ctx, login)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}
