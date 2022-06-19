package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
	_ "github.com/mattn/go-sqlite3"
)

const (
	initTimeOut = 60 * time.Second
	//queryTimeOut = 60 * time.Second
)

type Storage struct {
	db     *sql.DB
	ctx    context.Context
	cancel context.CancelFunc
	dsn    string
	stmts  map[string]*sql.Stmt
}

type Option func(*Storage)

func New(dsn string, opts ...Option) (*Storage, error) {
	if dsn == "" {
		return nil, fmt.Errorf("database DSN needed")
	}

	ctx, cancel := context.WithCancel(context.Background())

	s := &Storage{
		ctx:    ctx,
		cancel: cancel,
		dsn:    dsn,
		stmts:  make(map[string]*sql.Stmt),
	}
	// применяем в цикле каждую опцию
	for _, opt := range opts {
		opt(s) // *Storage как аргумент
	}

	// проинициализируем подключение к БД
	err := s.init(s.dsn)
	if err != nil {
		return nil, fmt.Errorf("database initialization failed - %w", err)
	}

	return s, nil
}

func (s *Storage) init(dsn string) error {
	var err error
	s.db, err = sql.Open("pgx", dsn)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(s.ctx, initTimeOut)
	defer cancel()

	err = s.initUsers(ctx)
	if err != nil {
		return fmt.Errorf(`failed to create 'users' table - %w`, err)
	}

	err = s.initSessions(ctx)
	if err != nil {
		return fmt.Errorf(`failed to create 'sessions' table - %w`, err)
	}

	err = s.initOrders(ctx)
	if err != nil {
		return fmt.Errorf(`failed to create 'orders' table - %w`, err)
	}

	err = s.initBalance(ctx)
	if err != nil {
		return fmt.Errorf(`failed to create 'balance' table - %w`, err)
	}

	err = s.initWithdrawals(ctx)
	if err != nil {
		return fmt.Errorf(`failed to create 'withdrawals' table - %w`, err)
	}

	s.db.SetMaxOpenConns(40)
	s.db.SetMaxIdleConns(20)
	s.db.SetConnMaxIdleTime(time.Second * 60)

	return nil
}

func (s *Storage) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := s.db.PingContext(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Storage) Shutdown() error {
	// Пошлём сигнал завершения через контекст: завершить все текущие запросы
	s.cancel()

	// Закроем все открытые стейтменты
	for _, stmt := range s.stmts {
		err := stmt.Close()
		if err != nil {
			return fmt.Errorf("failed to close staitement - %w", err)
		}
	}

	// Закроем соединение с БД
	err := s.db.Close()
	if err != nil {
		return err
	}

	log.Println("[DEBUG] Connection to database closed")
	return nil
}
