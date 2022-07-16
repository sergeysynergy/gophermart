package handlers

import (
	"fmt"
	"github.com/sergeysynergy/gophermart/internal/gophermart"
	"net/http"
)

func (h *handler) authCheck(w http.ResponseWriter, r *http.Request) (*gophermart.Session, error) {
	// извлечём токен сессии
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			h.error(w, r, gophermart.ErrUnauthorizedAccess, http.StatusUnauthorized)
			return nil, gophermart.ErrUnauthorizedAccess
		}
		h.error(w, r, err, http.StatusBadRequest)
		return nil, err
	}
	sessionToken := c.Value

	// получим сессию из хранилища по токену
	session, err := h.gm.Sessions.Get(sessionToken)
	if err != nil {
		err = fmt.Errorf("session token is not present")
		h.error(w, r, err, http.StatusUnauthorized)
		return nil, err
	}

	// Удаляем сессию и выходим, если прошёл срок годности
	if session.IsExpired() {
		h.gm.Sessions.Delete(sessionToken)
		err = fmt.Errorf("session has expired")
		h.error(w, r, err, http.StatusUnauthorized)
		return nil, err
	}

	return session, nil
}
