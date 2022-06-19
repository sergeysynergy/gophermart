package basicstorage

import (
	"fmt"
	"github.com/sergeysynergy/hardtest/internal/gophermart"
)

func (s *Storage) AddSession(userSession *gophermart.Session) error {
	s.sessionsBySessionTokenMu.Lock()
	s.sessionsBySessionToken[userSession.Token] = userSession
	s.sessionsBySessionTokenMu.Unlock()

	return nil
}

func (s *Storage) GetSession(token string) (*gophermart.Session, error) {
	s.sessionsBySessionTokenMu.RLock()
	defer s.sessionsBySessionTokenMu.RUnlock()

	userSession, ok := s.sessionsBySessionToken[token]
	if !ok {
		return nil, fmt.Errorf("session token not found")
	}

	return userSession, nil
}

func (s *Storage) DeleteSession(token string) error {
	s.sessionsBySessionTokenMu.Lock()
	delete(s.sessionsBySessionToken, token)
	s.sessionsBySessionTokenMu.Unlock()

	return nil
}
