package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/sergeysynergy/hardtest/internal/gophermart"
)

func (s *Storage) initSessions(ctx context.Context) error {
	dbName := "sessions"
	_, err := s.db.ExecContext(ctx, "select * from "+dbName+";")
	if err != nil {
		queryCreateTable := `
			CREATE TABLE ` + dbName + ` (
				user_id bigint NOT NULL,
				token varchar NOT NULL, 
				expiry time NOT NULL
			);
		`

		_, err = s.db.ExecContext(ctx, queryCreateTable)
		if err != nil {
			return err
		}

		log.Printf("[DEBUG] table `%s` created", dbName)
	}

	err = s.initSessionsStatements()
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) initSessionsStatements() error {
	dbName := "sessions"
	var err error
	var stmt *sql.Stmt

	stmt, err = s.db.PrepareContext(
		s.ctx,
		"INSERT INTO "+dbName+" (user_id, token, expiry) VALUES ($1, $2, $3)",
	)
	if err != nil {
		return err
	}
	s.stmts["sessionsInsert"] = stmt

	stmt, err = s.db.PrepareContext(
		s.ctx,
		"SELECT * FROM "+dbName+" WHERE token=$1",
	)
	if err != nil {
		return err
	}
	s.stmts["sessionsGet"] = stmt

	stmt, err = s.db.PrepareContext(
		s.ctx,
		"DELETE FROM "+dbName+" WHERE token=$1",
	)
	if err != nil {
		return err
	}
	s.stmts["sessionsDelete"] = stmt

	return nil
}

func (s *Storage) AddSession(session *gophermart.Session) error {
	_, err := s.stmts["sessionsInsert"].ExecContext(s.ctx, session.UserID, session.Token, session.Expiry)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) GetSession(token string) (*gophermart.Session, error) {
	session := &gophermart.Session{}
	row := s.stmts["sessionsGet"].QueryRowContext(s.ctx, token)
	err := row.Scan(&session.UserID, &session.Token, &session.Expiry)
	if err == sql.ErrNoRows {
		return nil, gophermart.ErrSessionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session - %w", err)
	}

	return session, nil
}

func (s *Storage) DeleteSession(token string) error {
	res, err := s.stmts["sessionsDelete"].ExecContext(s.ctx, token)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("session not found")
	}

	return nil
}
